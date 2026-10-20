package tpl

import (
	"regexp"
	"strings"
	"text/template"
)

// Funcs 返回模板通用函数库:命名转换等(与核心编译器的命名工具保持一致)。
func Funcs() template.FuncMap {
	return template.FuncMap{
		"hump":       HumpName,       // user_name -> UserName
		"lowerFirst": LowerFirstName, // UserName -> userName
		"underline":  UnderlineName,  // UserName -> user_name
		"allUpper":   AllUpperName,   // UserName -> USER_NAME
		"middleLine": MiddleLine,     // UserName -> user-name
		"firstLower": LowerFirst,     // UserName -> userName(纯首字母小写)
	}
}

// HumpName 下划线转驼峰:user_name -> UserName
func HumpName(val string) string {
	if len(val) == 0 {
		return val
	}
	var ret string
	for _, item := range strings.Split(val, "_") {
		ret += strings.ToUpper(item[:1]) + item[1:]
	}
	return ret
}

// LowerFirstName 下划线转驼峰后首字母小写:user_name -> userName
func LowerFirstName(val string) string {
	if len(val) == 0 {
		return val
	}
	var ret string
	for _, item := range strings.Split(val, "_") {
		ret += strings.ToUpper(item[:1]) + item[1:]
	}
	return strings.ToLower(ret[:1]) + ret[1:]
}

// LowerFirst 仅把首字母小写:UserName -> userName
func LowerFirst(val string) string {
	if len(val) == 0 {
		return val
	}
	return strings.ToLower(val[:1]) + val[1:]
}

var upperRe = regexp.MustCompile(`[A-Z_]`)

// UnderlineName 驼峰转下划线:UserName -> user_name
func UnderlineName(val string) string {
	if len(val) == 0 {
		return val
	}
	match := upperRe.FindAllStringSubmatchIndex(val, -1)
	if match == nil {
		return strings.ToLower(val)
	}
	var ret string
	index := 0
	for _, item := range match {
		temp := strings.ToLower(val[index:item[0]])
		if strings.HasPrefix(temp, "_") {
			temp = temp[1:]
		}
		if len(temp) == 0 {
			continue
		}
		if len(ret) == 0 {
			ret += temp
		} else {
			ret += "_" + temp
		}
		index = item[0]
	}
	if index < len(val) {
		temp := strings.ToLower(val[index:])
		if strings.HasPrefix(temp, "_") {
			temp = temp[1:]
		}
		if len(ret) == 0 {
			ret += temp
		} else {
			ret += "_" + temp
		}
	}
	return ret
}

// AllUpperName 驼峰转下划线全大写:UserName -> USER_NAME
func AllUpperName(val string) string {
	return strings.ToUpper(UnderlineName(val))
}

// MiddleLine 驼峰/下划线转中横线:UserName -> user-name
func MiddleLine(val string) string {
	ret := strings.Builder{}
	for i := 0; i < len(val); i++ {
		c := val[i]
		if '_' == c {
			if i > 0 {
				ret.WriteByte('-')
			}
		} else if c >= 'A' && c <= 'Z' {
			if i > 0 {
				ret.WriteByte('-')
			}
			ret.WriteByte(c - 'A' + 'a')
		} else {
			ret.WriteByte(c)
		}
	}
	return ret.String()
}
