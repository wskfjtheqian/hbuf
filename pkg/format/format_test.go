package format

import (
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"os"
	"strings"
	"testing"
)

const fmtSrc = `package go = "admin";
package java = "com.hbuf.admin";

import "/home/user/hbuf/pkg/parser/parser.go";

[verify: regex="123" min="1"];

//学生信息
data Student : Base = 2, Ext2 = 3 {
    //编号
    int64? no = 3; //注释

    string[] names = 5;

    string<int> maps = 6 "json:\"maps\"";
};

enum Status {
    player = 0;

    stop = 1;
};

//服务定义
server BaseServer {
    //查询学生
    Student GetStudent(Student req) = 1;

    void SendMessage(Message message) = 2;
};
`

// TestRoundTrip 格式化输出应可被 parser 重新解析,且再次格式化结果一致。
func TestRoundTrip(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/a/base.hbuf", []byte(fmtSrc), parser.AllErrors|parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var sb strings.Builder
	if err := Node(&sb, file); err != nil {
		t.Fatalf("format: %v", err)
	}
	out := sb.String()

	// 输出必须可被重新解析
	file2, err := parser.ParseFile(fset, "/a/base2.hbuf", []byte(out), parser.AllErrors|parser.ParseComments)
	if err != nil {
		t.Fatalf("re-parse: %v\n%s", err, out)
	}

	// 二次格式化输出应与第一次一致(幂等)
	var sb2 strings.Builder
	if err := Node(&sb2, file2); err != nil {
		t.Fatalf("re-format: %v", err)
	}
	if sb2.String() != out {
		t.Errorf("format not idempotent:\nfirst:\n%s\nsecond:\n%s", out, sb2.String())
	}
}

// TestNodeUnsupported 不支持的节点类型应报错。
func TestNodeUnsupported(t *testing.T) {
	if err := Node(os.Stdout, "not-a-node"); err == nil {
		t.Error("expect error for unsupported node")
	}
}
