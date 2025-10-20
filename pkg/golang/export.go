package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
	"strconv"
)

func (b *Builder) printExportCode(dst *build.Writer, data *ast.DataType) error {
	maps := make(map[string][]*ast.Field)
	lists := make([]*ast.Field, 0)
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		export, err := build.GetExport(field.Tags)
		if err != nil {
			return err
		}
		if export == nil {
			return nil
		}
		lists = append(lists, field)
		for _, key := range export.Filter {
			maps[key] = append(maps[key], field)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(maps) == 0 && len(lists) == 0 {
		return nil
	}

	keys := make([]string, 0, len(maps))
	for key := range maps {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	err = b.printExportHeaderCode(dst, data, keys, maps, lists)
	if err != nil {
		return err
	}

	err = b.printExportDataCode(dst, data, keys, maps, lists)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) printExportHeaderCode(dst *build.Writer, data *ast.DataType, keys []string, maps map[string][]*ast.Field, lists []*ast.Field) error {
	dName := build.StringToHumpName(data.Name.Name)
	dst.Code("func (g ").Code(dName).Code(") ExportHeader(key string) []string {\n")
	dst.Tab(1).Code("switch key {\n")
	for _, key := range keys {
		dst.Tab(1).Code("case \"").Code(key).Code("\":\n")
		dst.Tab(2).Code("return []string{\n")
		for _, item := range maps[key] {
			dst.Tab(3).Code("\"").Code(build.StringToHumpName(item.Name.Name)).Code("\",\n")
		}
		dst.Tab(2).Code("}\n")
	}
	dst.Tab(1).Code("default :\n")
	dst.Tab(2).Code("return []string{\n")
	for _, item := range lists {
		dst.Tab(3).Code("\"").Code(build.StringToHumpName(item.Name.Name)).Code("\",\n")
	}
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Code("}\n\n")
	return nil
}

func (b *Builder) printExportDataCode(dst *build.Writer, data *ast.DataType, keys []string, maps map[string][]*ast.Field, lists []*ast.Field) error {
	dName := build.StringToHumpName(data.Name.Name)
	dst.Code("func (g ").Code(dName).Code(") ExportData(key string, zoneOffset int32) ([]any, error){\n")

	for _, item := range lists {
		if build.GetBaseType(item.Type.Type()) == build.Date {
			dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hutl", "hutl")
			dst.Tab(1).Code("loc := hutl.ZoneByOffset(zoneOffset)\n")
			break
		}
	}

	dst.Tab(1).Code("switch key {\n")
	for _, key := range keys {
		dst.Tab(1).Code("case \"").Code(key).Code("\":\n")
		dst.Tab(2).Code("list := make([]any, ").Code(strconv.Itoa(len(maps[key]))).Code(")\n")
		for i, item := range maps[key] {
			err := b.printExportDataItemCode(dst, i, item)
			if err != nil {
				return err
			}
		}
		dst.Tab(2).Code("return list, nil\n")
	}
	dst.Tab(1).Code("default :\n")
	dst.Tab(2).Code("list := make([]any, ").Code(strconv.Itoa(len(lists))).Code(")\n")
	for i, item := range lists {
		err := b.printExportDataItemCode(dst, i, item)
		if err != nil {
			return err
		}
	}
	dst.Tab(2).Code("return list, nil\n")
	dst.Tab(1).Code("}\n")

	dst.Code("}\n\n")
	return nil
}

func (b *Builder) printExportDataItemCode(dst *build.Writer, i int, item *ast.Field) error {
	name := build.StringToHumpName(item.Name.Name)
	isNil := build.IsNil(item.Type)
	tab := 2
	if isNil {
		dst.Tab(2).Code("if g.").Code(name).Code(" != nil {\n")
		tab++
	}

	if build.IsArray(item.Type) || build.IsMap(item.Type) {
		dst.Import("encoding/json", "")
		dst.Tab(tab).Code("bytes, err := json.Marshal(g.Get").Code(name).Code("())\n")
		dst.Tab(tab).Code("if err!= nil {\n")
		dst.Tab(tab + 1).Code("return nil, err\n")
		dst.Tab(tab).Code("}\n")
		dst.Tab(tab).Code("list[").Code(strconv.Itoa(i)).Code("] = string(bytes)\n")
	} else if build.IsEnum(item.Type.Type()) {
		dst.Tab(tab).Code("list[").Code(strconv.Itoa(i)).Code("] = g.Get").Code(name).Code("().ToName()\n")
	} else if build.GetBaseType(item.Type.Type()) == build.Date {
		dst.Import("time", "")
		dst.Tab(tab).Code("list[").Code(strconv.Itoa(i)).Code("] = time.Time(g.Get").Code(name).Code("()).In(loc)\n")
	} else {
		dst.Tab(tab).Code("list[").Code(strconv.Itoa(i)).Code("] = g.Get").Code(name).Code("()\n")
	}
	if isNil {
		dst.Tab(2).Code("}\n")
	}
	return nil
}
