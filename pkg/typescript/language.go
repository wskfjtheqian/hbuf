package ts

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

		dst.Tab(0).Code("export const ").Code(build.StringToFirstLower(langsKey)).Code("Lang = {\n")
		for key, _ := range l.Key {
			dst.Tab(1).Code(build.StringToFirstLower(key)).Code(": {\n")
			for name, lang := range l.Lang {
				dst.Tab(2).Code(name).Code(": ").Code("\"").Code(lang[key]).Code("\",\n")
			}
			dst.Tab(1).Code("},\n")
		}
		dst.Tab(0).Code("}\n")
	}
}
