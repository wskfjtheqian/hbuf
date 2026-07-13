package ts

import (
	"hbuf/pkg/build"
	"sort"
)

func printLanguage(lang map[string]*build.Language, dst *build.Writer) {
	langKeys := build.GetKeysByMap(lang)
	sort.Strings(langKeys)
	for _, langKey := range langKeys {
		l := lang[langKey]
		if 0 >= len(l.Lang) {
			continue
		}

		dst.Tab(0).Code("export const ").Code(build.StringToFirstLower(langKey)).Code("Lang = {\n")
		keys := build.GetMapKeys(l.Key)
		sort.Strings(keys)

		for _, key := range keys {
			dst.Tab(1).Code(build.StringToFirstLower(key)).Code(": {\n")

			names := build.GetMapKeys(l.Lang)
			sort.Strings(names)
			for _, name := range names {
				dst.Tab(2).Code(name).Code(": ").Code("\"").Code(l.Lang[name][key]).Code("\",\n")
			}
			dst.Tab(1).Code("},\n")
		}
		dst.Tab(0).Code("}\n")
	}
}
