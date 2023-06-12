package dart

import (
	"hbuf/pkg/build"
	"sort"
)

func printLanguge(dst *build.Writer) {
	langsKeys := build.GetKeysByMap(dst.GetLangs())
	sort.Strings(langsKeys)
	for _, langsKey := range langsKeys {
		l := dst.GetLangs()[langsKey]
		if 0 >= len(l.Lang) {
			continue
		}
		dst.Import("package:flutter/foundation.dart", "")

		dst.Code("\n")
		dst.Code("class _" + l.Name + "LocalizationsDelegate extends LocalizationsDelegate<" + l.Name + "Localizations> {\n")
		dst.Code("  const _" + l.Name + "LocalizationsDelegate();\n")
		dst.Code("\n")
		dst.Code("  @override\n")
		dst.Code("  bool isSupported(Locale locale) => true;\n")
		dst.Code("\n")
		dst.Code("  @override\n")
		dst.Code("  Future<" + l.Name + "Localizations> load(Locale locale) {\n")
		dst.Code("    switch (locale.languageCode) {\n")

		keyKeys := build.GetKeysByMap(l.Key)
		sort.Strings(keyKeys)
		for _, key := range keyKeys {
			keyName := "_" + build.StringToHumpName(key)
			dst.Code("      case '" + key + "':\n")
			dst.Code("        return SynchronousFuture<" + l.Name + "Localizations>(const " + keyName + l.Name + "Localizations());\n")
		}
		dst.Code("    }\n")
		dst.Code("   return SynchronousFuture<" + l.Name + "Localizations>(const Default" + l.Name + "Localizations());\n")
		dst.Code("  }\n")
		dst.Code("  @override\n")
		dst.Code("  bool shouldReload(_" + l.Name + "LocalizationsDelegate old) => true;\n")
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

		lanKeys := build.GetKeysByMap(l.Lang)
		sort.Strings(lanKeys)
		for _, key := range lanKeys {
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

		for _, key := range lanKeys {
			dst.Code("  @override\n")
			dst.Code("  String get " + key + " => \"" + build.GetLang(key, "____", l.Lang[key]) + "\";\n")
			dst.Code("\n")
		}
		dst.Code("}\n")
		dst.Code("\n")

		for _, key := range keyKeys {
			keyName := "_" + build.StringToHumpName(key)
			dst.Code("class " + keyName + l.Name + "Localizations implements " + l.Name + "Localizations {\n")
			dst.Code("  const " + keyName + l.Name + "Localizations();\n")
			for _, f := range lanKeys {
				dst.Code("  @override\n")
				dst.Code("  String get " + f + " => \"" + build.GetLang(f, key, l.Lang[f]) + "\";\n")
				dst.Code("\n")
			}
			dst.Code("}\n")
			dst.Code("\n")
		}
	}
}
