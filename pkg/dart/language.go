package dart

import (
	"hbuf/pkg/build"
)

func printLanguge(dst *build.Writer) {
	for _, l := range dst.GetLangs() {
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
		for key, _ := range l.Key {
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

		for key, _ := range l.Lang {
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

		for f, lan := range l.Lang {
			dst.Code("  @override\n")
			dst.Code("  String get " + f + " => \"" + build.GetLang(f, "____", lan) + "\";\n")
			dst.Code("\n")
		}
		dst.Code("}\n")
		dst.Code("\n")

		for key, _ := range l.Key {
			keyName := "_" + build.StringToHumpName(key)
			dst.Code("class " + keyName + l.Name + "Localizations implements " + l.Name + "Localizations {\n")
			dst.Code("  const " + keyName + l.Name + "Localizations();\n")
			for f, lan := range l.Lang {
				dst.Code("  @override\n")
				dst.Code("  String get " + f + " => \"" + build.GetLang(f, key, lan) + "\";\n")
				dst.Code("\n")
			}
			dst.Code("}\n")
			dst.Code("\n")
		}
	}
}
