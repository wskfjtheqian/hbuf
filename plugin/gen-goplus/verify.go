package goplus

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strconv"
	"strings"
	"time"
)

// VerifyContext verify 输出模板的 data 类型上下文。
type VerifyContext struct {
	Name    string
	LowName string
	Fields  []*VerifyFieldContext
	Maps    bool // 类型带 verify 注解时生成 VerifyMaps 与 Verify 分发
}

// VerifyFieldContext 单个字段的校验上下文:Name 为方法名,Key 为字段原名,
// Body 为校验分支(含缩进,预渲染)。
type VerifyFieldContext struct {
	Name string
	Key  string
	Body string
}

// buildVerify 为带 verify 注解的 data 类型构建 verify 模板上下文。
func (c *Ctx) buildVerify(imp Imports, data *wire.DataType) (*VerifyContext, error) {
	imp.Import("context", "")
	vc := &VerifyContext{Name: tpl.HumpName(data.Name), LowName: tpl.LowerFirstName(data.Name)}
	err := c.enumField(data, func(field *wire.Field, data *wire.DataType) error {
		verify := getVerify(field.Tags)
		if verify == nil {
			return nil
		}
		body, err := c.buildVerifyBody(imp, field, verify)
		if err != nil {
			return err
		}
		vc.Fields = append(vc.Fields, &VerifyFieldContext{
			Name: tpl.HumpName(field.Name),
			Key:  field.Name,
			Body: body,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	vc.Maps = getVerify(data.Tags) != nil
	return vc, nil
}

// buildVerifyBody 渲染字段校验分支,与进程内 printVerifyFieldCode 一致。
func (c *Ctx) buildVerifyBody(imp Imports, field *wire.Field, verify *Verify) (string, error) {
	fName := tpl.HumpName(field.Name)
	var sb strings.Builder

	first := true
	for i, val := range verify.Format {
		enum, item, pack := c.verifyEnumRef(val, imp)
		if enum == nil || item == nil {
			continue
		}
		f := getFormat(item.Tags)
		if f == nil {
			continue
		}

		if isNil(field.Type) && 0 == i {
			imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
			if !f.Null {
				sb.WriteString("\tif nil == i." + fName)
				if baseType(field.Type) == "string" {
					sb.WriteString(" || len(i.Get" + fName + "()) == 0")
				}
				sb.WriteString(" {\n")
				sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
				sb.WriteString("\t}\n")
			} else {
				sb.WriteString("\tif nil == i." + fName)
				if baseType(field.Type) == "string" {
					sb.WriteString(" || len(i.Get" + fName + "()) == 0")
				}
				sb.WriteString(" {\n")
				sb.WriteString("\t\treturn nil\n")
				sb.WriteString("\t}\n")
			}
		}
		if isEnum(field.Type) {
			imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
			sb.WriteString("\tif 0 == len(i.Get" + fName + "().ToName()) {\n")
			sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
			sb.WriteString("\t}\n")
		} else if isMap(field.Type) {
			// 与进程内后端一致:map 无格式校验
		} else if isArray(field.Type) {
			for _, item := range f.Len {
				imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
				imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
				sb.WriteString("\tif !(len(i." + fName + ") " + getOperator(item.Op) + " " + item.Val + ") {\n")
				sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
				sb.WriteString("\t}\n")
			}
		} else {
			t := baseType(field.Type)
			switch t {
			case "int8", "int16", "int32", "uint8", "uint16", "uint32", "float", "double", "uint64", "int64":
				for _, item := range f.Val {
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")

					sb.WriteString("\tif !(i.Get" + fName + "() " + getOperator(item.Op) + " " + item.Val + ") {\n")
					sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
					sb.WriteString("\t}\n")
				}
			case "date":
				for _, item := range f.Val {
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")

					parse, err := time.Parse("2006-01-02T15:04:05Z", item.Val)
					if err != nil {
						return "", err
					}

					sb.WriteString("\tif !(i.Get" + fName + "() " + getOperator(item.Op) + " " + strconv.FormatInt(parse.UnixMilli(), 10) + ") {\n")
					sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
					sb.WriteString("\t}\n")
				}
			case "decimal":
				for _, item := range f.Val {
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
					imp.Import("github.com/shopspring/decimal", "")

					sb.WriteString("\tif ")
					if OperatorGt == item.Op {
						sb.WriteString("!i.Get" + fName + "().GreaterThan")
					} else if OperatorLt == item.Op {
						sb.WriteString("!i.Get" + fName + "().LessThan")
					} else if OperatorGte == item.Op {
						sb.WriteString("!i.Get" + fName + "().GreaterThanOrEqual")
					} else if OperatorLte == item.Op {
						sb.WriteString("!i.Get" + fName + "().LessThanOrEqual")
					} else if OperatorEq == item.Op {
						sb.WriteString("!i.Get" + fName + "().Equal")
					} else if OperatorNeq == item.Op {
						sb.WriteString("i.Get" + fName + "().Equal")
					}
					sb.WriteString("(decimal.NewFromFloat(" + item.Val + ")) {\n")
					sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
					sb.WriteString("\t}\n")
				}

			case "string":
				for _, item := range f.Len {
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
					imp.Import("unicode/utf8", "")
					sb.WriteString("\tif !(utf8.RuneCountInString(i.Get" + fName + "()) " + getOperator(item.Op) + " " + item.Val + ") {\n")
					sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
					sb.WriteString("\t}\n")
				}
				for _, item := range f.Val {
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "")
					if item.Op == OperatorMatch || item.Op == OperatorNotMatch {
						sb.WriteString("\tmatch, err ")
						if first {
							sb.WriteString(":")
						}
						imp.Import("regexp", "")
						sb.WriteString("= regexp.MatchString(\"" + item.Val + "\", i.Get" + fName + "())\n")
						sb.WriteString("\tif err != nil {\n")
						sb.WriteString("\t\treturn err\n")
						sb.WriteString("\t}\n")
						sb.WriteString("\tif false " + getMatch(item.Op) + " match {\n")
						sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
						sb.WriteString("\t}\n")
						first = false
					} else {
						sb.WriteString("\tif !(i.Get" + fName + "() " + getOperator(item.Op) + " " + item.Val + ") {\n")
						sb.WriteString("\t\treturn hrpc.NewResult[hbuf.Data](int32(" + pack + "), " + pack + ".ToName(), nil)\n")
						sb.WriteString("\t}\n")
					}
				}
			}
		}
	}

	sb.WriteString("\treturn nil\n")
	return sb.String(), nil
}

// findEnumItem 在枚举声明中按名称查找条目。
func findEnumItem(enum *wire.EnumType, name string) *wire.EnumItem {
	if enum == nil {
		return nil
	}
	for _, item := range enum.Items {
		if item != nil && item.Name == name {
			return item
		}
	}
	return nil
}

// verifyEnumRef 解析 verify 格式引用,返回枚举声明、条目与渲染前缀(含包前缀)。
func (c *Ctx) verifyEnumRef(val VerifyEnum, imp Imports) (*wire.EnumType, *wire.EnumItem, string) {
	enumType, enumFile := c.Find(val.Enum)
	if enumType == nil || enumType.Enum == nil {
		return nil, nil, ""
	}
	item := findEnumItem(enumType.Enum, val.Item)
	if item == nil {
		return nil, nil, ""
	}
	pack := ""
	if enumFile != nil {
		pack = c.refPrefix(enumFile.Packages["go"], imp)
	}
	return enumType.Enum, item, pack + tpl.HumpName(val.Enum) + tpl.HumpName(val.Item)
}

func getOperator(op Operator) string {
	switch op {
	case OperatorGt:
		return ">"
	case OperatorLt:
		return "<"
	case OperatorGte:
		return ">="
	case OperatorLte:
		return "<="
	case OperatorEq:
		return "=="
	case OperatorNeq:
		return "!="
	}
	return ""
}

func getMatch(op Operator) string {
	switch op {
	case OperatorMatch:
		return "=="
	case OperatorNotMatch:
		return "!="
	}
	return ""
}
