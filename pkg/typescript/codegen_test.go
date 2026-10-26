package ts

import (
	"hbuf/pkg/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleIDL = `package test = "test";

data HBufSubTest {
    int64 A1 = 1
    int64? A2 = 2
}

enum Status {
    Ok = 0
    Fail = 1
}

data HBufTest {
    int8 V1 = 1
    int8? V2 = 2
    int16 V3 = 3
    int32 V4 = 4
    int64 V5 = 5
    uint8 V6 = 6
    uint16 V7 = 7
    uint32 V8 = 8
    uint64 V9 = 9
    float V10 = 10
    double V11 = 11
    bool V12 = 12
    string V13 = 13
    bytes V14 = 14
    date V15 = 15
    decimal V16 = 16
    Status V17 = 17
    Status? V18 = 18
    HBufSubTest V19 = 19
    HBufSubTest? V20 = 20
    int8[] V21 = 21
    int8?[] V22 = 22
    int8<int64> V23 = 23
    int8<string> V24 = 24
    HBufSubTest[] V25 = 25
    date[] V26 = 26
}

data HBufSub1Test:HBufSubTest = 1 {
    int64 V1 = 1
    string? V2 = 2
}
`

// TestCodegenTs 生成样例 IDL 的 TS 代码，验证生成器端到端可运行。
func TestCodegenTs(t *testing.T) {
	dir := t.TempDir()
	hbufFile := filepath.Join(dir, "test.hbuf")
	if err := os.WriteFile(hbufFile, []byte(sampleIDL), 0644); err != nil {
		t.Fatal(err)
	}
	build.AddBuildType("ts", Build)
	if err := build.Build(filepath.Join(dir, "out"), hbufFile, "ts", "", true, "D"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "out", "test.data.ts"))
	if err != nil {
		t.Fatal(err)
	}
	code := string(data)
	// 关键形态断言：类实现 HData + desc 单例 + 类型映射
	for _, want := range []string{
		"export class HBufTest implements Data, HData",
		"export class HBufSub1Test implements Data, HData, HBufSubTest",
		"descriptors(): HBufDescriptor",
		"export const hBufTestDesc: HBufDescriptor = NewSyncDescriptor",
		"v15: Date | null = null",
		"v16: string = \"\"",
		"v23: Map<(bigint), (number)> = new Map()",
		"NewRefDescriptor(hBufSubTestDesc",
		"get: (o: HBufSub1Test) => o, set: (o: HBufSub1Test, v) => {}",
		"Unmarshal(new Uint8Array(_data), ret)",
		"return Marshal(this).slice().buffer",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("missing %q in generated code", want)
		}
	}
	t.Logf("out dir: %s", filepath.Join(dir, "out"))
}
