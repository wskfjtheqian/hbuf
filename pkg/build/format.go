package build

import (
	"hbuf/pkg/ast"
	"strings"
)

// 操作符
type Operator int

// ，， ，， ，，，
const (
	OperatorGt       Operator = 1 + iota //大于 >
	OperatorLt                           // 小于 <
	OperatorEq                           // 等于 ==
	OperatorGte                          // 大于等于 >=
	OperatorLte                          // 小于等于 <=
	OperatorNeq                          // 不等于 !=
	OperatorMatch                        // 匹配 =~
	OperatorNotMatch                     // 匹配 !~
)

// 表达式
type Expr struct {
	//操作符
	Op Operator

	//值
	Val string
}

type Format struct {
	Null bool
	Val  []Expr
	Len  []Expr
}

func GetFormat(tags []*ast.Tag) *Format {
	val, ok := GetTag(tags, "format")
	if !ok {
		return nil
	}
	f := &Format{
		Null: false,
	}
	var err error
	if nil != val.KV {
		for _, item := range val.KV {
			if "null" == item.Name.Name {
				f.Null = "true" == item.Values[0].Value[1:len(item.Values[0].Value)-1]
			} else if "val" == item.Name.Name {
				f.Val, err = parseExpr(item.Values)
				if err != nil {
					return f
				}
			} else if "len" == item.Name.Name {
				f.Len, err = parseExpr(item.Values)
				if err != nil {
					return f
				}
			}
		}
	}
	return f
}

// 解析Expr
func parseExpr(values []*ast.BasicLit) ([]Expr, error) {
	val := make([]Expr, 0)
	for _, value := range values {
		v, err := parseExprItem(value.Value[1 : len(value.Value)-1])
		if err != nil {
			return nil, err
		}
		if v.Op > 0 {
			val = append(val, v)
		}
	}
	return val, nil
}

func parseExprItem(value string) (Expr, error) {
	var e Expr
	if strings.HasPrefix(value, "> ") {
		e.Op = OperatorGt
		e.Val = value[2:]
	} else if strings.HasPrefix(value, "< ") {
		e.Op = OperatorLt
		e.Val = value[2:]
	} else if strings.HasPrefix(value, "== ") {
		e.Op = OperatorEq
		e.Val = value[3:]
	} else if strings.HasPrefix(value, ">= ") {
		e.Op = OperatorGte
		e.Val = value[3:]
	} else if strings.HasPrefix(value, "<= ") {
		e.Op = OperatorLte
		e.Val = value[3:]
	} else if strings.HasPrefix(value, "!= ") {
		e.Op = OperatorNeq
		e.Val = value[3:]
	} else if strings.HasPrefix(value, "=~ ") {
		e.Op = OperatorMatch
		e.Val = value[3:]
	} else if strings.HasPrefix(value, "!~ ") {
		e.Op = OperatorNotMatch
		e.Val = value[3:]
	} else {
		return e, nil
	}
	return e, nil
}
