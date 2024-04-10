package build

import (
	"hbuf/pkg/ast"
	"regexp"
	"strings"
)

type Language struct {
	Name string
	Key  map[string]struct{}
	Lang map[string]map[string]string
}

func NewLanguage(name string) *Language {
	return &Language{
		Name: name,
		Key:  map[string]struct{}{},
		Lang: map[string]map[string]string{},
	}
}

func (l *Language) Add(field string, tags []*ast.Tag) {
	lang := getLanguage(tags)
	l.Lang[field] = lang
	l.Key["en"] = struct{}{}

	for key, _ := range lang {
		l.Key[key] = struct{}{}
	}
}

func getLanguage(tags []*ast.Tag) map[string]string {
	val, ok := GetTag(tags, "lang")
	if !ok {
		return nil
	}

	lang := make(map[string]string, 0)
	if nil != val.KV {
		for _, item := range val.KV {
			lang[StringToFirstLower(item.Name.Name)] = item.Values[0].Value[1 : len(item.Values[0].Value)-1]
		}
	}
	return lang
}

func GetLang(val string, key string, lanMap map[string]string) string {
	if nil != lanMap {
		if text, ok := lanMap[key]; ok {
			return text
		}
	}

	if 0 == len(val) {
		return val
	}

	if !regexp.MustCompile(`[a-z0-9]`).MatchString(val) {
		val = strings.ToLower(val)
	}

	rex := regexp.MustCompile(`[A-Z_]`)
	match := rex.FindAllStringSubmatchIndex(val, -1)
	if nil == match {
		return strings.ToUpper(val[:1]) + strings.ToLower(val[1:])
	}
	var ret string
	var index = 0
	for _, item := range match {
		temp := val[index:item[0]]
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 == len(temp) {
			continue
		}
		if 0 == len(ret) {
			ret += strings.ToUpper(temp[:1])
			ret += strings.ToLower(temp[1:])
		} else {
			ret += " " + strings.ToLower(temp)
		}
		index = item[0]
	}
	if index < len(val) {
		temp := val[index:]
		if 0 == strings.Index(temp, "_") {
			temp = temp[1:]
		}
		if 0 < len(temp) {
			ret += " " + strings.ToLower(temp)
		}
	}
	return ret
}
