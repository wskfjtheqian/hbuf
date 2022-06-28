package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"regexp"
	"strings"
)

type Language struct {
	Name string
	key  map[string]struct{}
	lang map[string]map[string]string
}

func NewLanguage(name string) *Language {
	return &Language{
		Name: name,
		key:  map[string]struct{}{},
		lang: map[string]map[string]string{},
	}
}

func (l *Language) Add(field string, tags []*ast.Tag) {
	lang := getLanguage(tags)
	l.lang[field] = lang
	l.key["en"] = struct{}{}

	for key, _ := range lang {
		l.key[key] = struct{}{}
	}
}

func getLanguage(tags []*ast.Tag) map[string]string {
	val, ok := build.GetTag(tags, "lang")
	if !ok {
		return nil
	}

	lang := make(map[string]string, 0)
	if nil != val.KV {
		for _, item := range val.KV {
			lang[strings.ToLower(item.Name.Name)] = item.Value.Value[1 : len(item.Value.Value)-1]
		}
	}
	return lang
}

func (l *Language) printLanguage(dst *Writer) {
	if 0 == len(l.lang) {
		return
	}
	dst.Import("package:flutter/foundation.dart")

	dst.Code("\n")
	dst.Code("class _" + l.Name + "LocalizationsDelegate extends LocalizationsDelegate<" + l.Name + "Localizations> {\n")
	dst.Code("  const _" + l.Name + "LocalizationsDelegate();\n")
	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  bool isSupported(Locale locale) => locale.languageCode == 'en';\n")
	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  Future<" + l.Name + "Localizations> load(Locale locale) {\n")
	dst.Code("    switch (locale.languageCode) {\n")
	for key, _ := range l.key {
		keyName := "_" + build.StringToHumpName(key)
		dst.Code("      case '" + key + "':\n")
		dst.Code("        return SynchronousFuture<" + l.Name + "Localizations>(const " + keyName + l.Name + "Localizations());\n")
	}
	dst.Code("    }\n")
	dst.Code("   return SynchronousFuture<" + l.Name + "Localizations>(const Default" + l.Name + "Localizations());\n")
	dst.Code("  }\n")
	dst.Code("  @override\n")
	dst.Code("  bool shouldReload(_" + l.Name + "LocalizationsDelegate old) => false;\n")
	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code(" String toString() => 'Default" + l.Name + "Localizations.delegate(en_US)';\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("abstract class " + l.Name + "Localizations {\n")
	dst.Code("  static " + l.Name + "Localizations of(BuildContext context) {\n")
	dst.Code("    return Localizations.of<" + l.Name + "Localizations>(context, " + l.Name + "Localizations) ?? const Default" + l.Name + "Localizations();\n")
	dst.Code("  }\n")
	dst.Code("\n")
	dst.Code("  static const LocalizationsDelegate<" + l.Name + "Localizations> delegate = _" + l.Name + "LocalizationsDelegate();\n")
	dst.Code("\n")

	for key, _ := range l.lang {
		dst.Code("  String get " + key + ";\n")
		dst.Code("\n")
	}
	dst.Code("\n")
	dst.Code("}\n")
	dst.Code("\n")

	dst.Code("class Default" + l.Name + "Localizations implements " + l.Name + "Localizations {\n")
	dst.Code("  const Default" + l.Name + "Localizations();\n")
	dst.Code("\n")
	dst.Code("  static Future<" + l.Name + "Localizations> load(Locale locale) {\n")
	dst.Code("    return SynchronousFuture<" + l.Name + "Localizations>(const Default" + l.Name + "Localizations());\n")
	dst.Code("  }\n")
	dst.Code("\n")

	for f, lan := range l.lang {
		dst.Code("  @override\n")
		dst.Code("  String get " + f + " => \"" + getLang(f, "____", lan) + "\";\n")
		dst.Code("\n")
	}
	dst.Code("}\n")
	dst.Code("\n")

	for key, _ := range l.key {
		keyName := "_" + build.StringToHumpName(key)
		dst.Code("class " + keyName + l.Name + "Localizations implements " + l.Name + "Localizations {\n")
		dst.Code("  const " + keyName + l.Name + "Localizations();\n")
		for f, lan := range l.lang {
			dst.Code("  @override\n")
			dst.Code("  String get " + f + " => \"" + getLang(f, key, lan) + "\";\n")
			dst.Code("\n")
		}
		dst.Code("}\n")
		dst.Code("\n")
	}

}

func getLang(val string, key string, lanMap map[string]string) string {
	if text, ok := lanMap[key]; ok {
		return text
	}

	if 0 == len(val) {
		return val
	}

	rex := regexp.MustCompile(`[A-Z]`)
	match := rex.FindAllStringSubmatchIndex(val, -1)
	if nil == match {
		return strings.ToLower(val)
	}
	var ret string
	var index = 0
	for _, item := range match {
		temp := val[index:item[0]]
		if 0 == len(ret) {
			ret += strings.ToUpper(temp[:1])
			ret += strings.ToLower(temp[1:])
		} else {
			ret += " " + strings.ToLower(val[index:item[0]])
		}
		index = item[0]
	}
	if index < len(val) {
		ret += " " + strings.ToLower(val[index:])
	}
	return ret
}
