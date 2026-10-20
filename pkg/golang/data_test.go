package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"regexp"
	"strings"
	"testing"
)

type sqlParam struct {
	text string
}

func TestName(t *testing.T) {
	var allRex = regexp.MustCompile(`(\?{\w+})|(\${\w+})|\$|\?`)
	var text = "adfasfa?afasd$dfasd${fdasd}sfasdf?{afadfasd}"
	match := allRex.FindAllStringSubmatchIndex(text, -1)
	if nil != match {
		var index = 0
		buf := strings.Builder{}
		for _, item := range match {
			buf.WriteString(text[index:item[0]])
			index = item[0] + 1

		}
		if index < len(text) {
			buf.WriteString(text[index:])
		}
		text = buf.String()
	}
}

// TestGetPackageImportPath 跨文件引用类型的 import 路径应为 "<构建包前缀>/<目标包名>"。
func TestGetPackageImportPath(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/a/sms.public.hbuf", `package go = "sms";
data Req {
    SmsProvider provider = 1;
};`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	pubFile, err := parser.ParseFile(fset, "/a/public.hbuf", `package go = "public";
enum SmsProvider {
    aliyun = 0;
};`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}

	// 模拟跨文件引用链接:Req.provider 的类型指向 public.hbuf 的 SmsProvider
	req := file.Specs[0].(*ast.TypeSpec).Type.(*ast.DataType)
	provider := req.Fields.List[0].Type.(*ast.VarType).TypeExpr.(*ast.Ident)
	provider.Obj = &ast.Object{Kind: ast.Enum, Data: pubFile}

	b := &Builder{packages: "github.com/test/hbuf"}
	dst := build.NewWriter()
	dst.Packages = "sms"

	if got := b.getPackage(dst, provider); got != "public." {
		t.Errorf("prefix = %q, want %q", got, "public.")
	}
	if m := dst.GetImport("github.com/test/hbuf/public"); len(m) == 0 {
		t.Errorf("import %q not registered: %v", "github.com/test/hbuf/public", dst.GetImports())
	}

	// 构建包前缀为空时,import 退化为末段包名
	b2 := &Builder{}
	dst2 := build.NewWriter()
	dst2.Packages = "sms"
	if got := b2.getPackage(dst2, provider); got != "public." {
		t.Errorf("prefix(no pack) = %q, want %q", got, "public.")
	}
	if m := dst2.GetImport("public"); len(m) == 0 {
		t.Errorf("import %q not registered: %v", "public", dst2.GetImports())
	}
}
