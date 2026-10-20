package lsp

import (
	"hbuf/pkg/ast"
	"strings"
)

// completion 补全:基础类型、已声明类型/枚举、字段名、方法名、注解 tag 名。
func (s *Server) completion(doc *Document, pos Position) *CompletionList {
	pf := s.view.parse(doc)
	src := []byte(doc.Text)
	offset := toOffset(src, pos)
	prefix := linePrefix(src, offset)
	trimmed := strings.TrimSpace(prefix)
	word, _ := wordAtPrefix(prefix)

	// tag 上下文:[name 或 [name: key
	if strings.HasPrefix(trimmed, "[") {
		if strings.Contains(trimmed, ":") {
			return &CompletionList{Items: tagKeyItems(pf, doc, word)}
		}
		return &CompletionList{Items: tagNameItems(pf, doc, word)}
	}

	// package 语言名
	if trimmed == "package" || strings.HasPrefix(trimmed, "package ") {
		return &CompletionList{Items: filterItems(languageItems(), word)}
	}

	// import 路径:不补全
	if trimmed == "import" || strings.HasPrefix(trimmed, "import ") {
		return &CompletionList{Items: nil}
	}

	file := pf.pkg.Files[doc.Path]
	ctx := findContext(pf, file, offset)
	words := strings.Fields(trimmed)

	// 块声明行(data/enum/server 类型名位置):不补全
	if ctx.kind == ctxTop || ctx.kind == ctxData || ctx.kind == ctxServer || ctx.kind == ctxEnum {
		first := ""
		if len(words) > 0 {
			first = words[0]
		}
		switch first {
		case "data", "enum", "server", "import", "package":
			return &CompletionList{Items: nil}
		}
	}

	// 块内成员位置:类型已写完,补成员名
	switch ctx.kind {
	case ctxData, ctxDataField:
		if len(words) >= 1 {
			return &CompletionList{Items: filterItems(fieldItems(ctx.data), word)}
		}
	case ctxDataExtends:
		return &CompletionList{Items: filterItems(typeItems(pf, doc), word)}
	case ctxServer, ctxServerMethod:
		if len(words) >= 1 {
			return &CompletionList{Items: filterItems(methodItems(ctx.server), word)}
		}
	case ctxEnum, ctxEnumItem:
		return &CompletionList{Items: filterItems(enumItemItems(ctx.enum), word)}
	}

	// 默认:关键字 + 基础类型 + 类型
	var items []CompletionItem
	items = append(items, keywordItems()...)
	items = append(items, typeItems(pf, doc)...)
	return &CompletionList{Items: filterItems(items, word)}
}

// filterItems 按当前输入词前缀过滤。
func filterItems(items []CompletionItem, word string) []CompletionItem {
	if word == "" {
		return items
	}
	ret := items[:0]
	for _, item := range items {
		if strings.HasPrefix(item.Label, word) {
			ret = append(ret, item)
		}
	}
	return ret
}

// keywordItems 语言关键字与基础类型。
func keywordItems() []CompletionItem {
	words := []string{"package", "import", "data", "enum", "server", "void", "stream"}
	items := make([]CompletionItem, 0, len(words)+len(baseTypes))
	for _, w := range words {
		items = append(items, CompletionItem{Label: w, Kind: CompletionKeyword, Detail: "关键字"})
	}
	for _, w := range []string{
		"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64",
		"bool", "float", "double", "decimal", "bytes", "string", "date",
	} {
		items = append(items, CompletionItem{Label: w, Kind: CompletionValue, Detail: "基础类型"})
	}
	return items
}

// typeItems 已声明类型(当前文件 + import 文件)。
func typeItems(pf *parsedFile, doc *Document) []CompletionItem {
	var items []CompletionItem
	seen := map[string]bool{}

	file := pf.pkg.Files[doc.Path]
	if file != nil && file.Scope != nil {
		for _, obj := range file.Scope.Objects {
			ts, ok := obj.Decl.(*ast.TypeSpec)
			if !ok || seen[obj.Name] {
				continue
			}
			seen[obj.Name] = true
			items = append(items, CompletionItem{
				Label:  obj.Name,
				Kind:   typeKindOf(ts.Type),
				Detail: typeWordOf(ts.Type) + " 类型",
			})
		}
	}

	if file != nil {
		for _, spec := range file.Imports {
			dep := pf.pkg.Files[spec.Path.Value]
			if dep == nil || dep.Scope == nil {
				continue
			}
			for _, obj := range dep.Scope.Objects {
				ts, ok := obj.Decl.(*ast.TypeSpec)
				if !ok || seen[obj.Name] {
					continue
				}
				seen[obj.Name] = true
				items = append(items, CompletionItem{
					Label:  obj.Name,
					Kind:   typeKindOf(ts.Type),
					Detail: typeWordOf(ts.Type) + " 类型(import)",
				})
			}
		}
	}
	return items
}

// fieldItems data 已有字段名。
func fieldItems(data *ast.DataType) []CompletionItem {
	if data == nil {
		return nil
	}
	var items []CompletionItem
	for _, field := range data.Fields.List {
		items = append(items, CompletionItem{
			Label:  field.Name.Name,
			Kind:   CompletionField,
			Detail: typeString(field.Type) + " = " + field.Id.Value,
		})
	}
	return items
}

// methodItems server 已有方法名。
func methodItems(server *ast.ServerType) []CompletionItem {
	if server == nil {
		return nil
	}
	var items []CompletionItem
	for _, method := range server.Methods {
		items = append(items, CompletionItem{
			Label:  method.Name.Name,
			Kind:   CompletionMethod,
			Detail: typeString(method.Result) + "(" + typeString(method.Param) + " " + method.ParamName.Name + ")",
		})
	}
	return items
}

// enumItemItems enum 已有项名。
func enumItemItems(enum *ast.EnumType) []CompletionItem {
	if enum == nil {
		return nil
	}
	var items []CompletionItem
	for _, item := range enum.Items {
		items = append(items, CompletionItem{
			Label:  item.Name.Name,
			Kind:   CompletionEnumMember,
			Detail: "= " + item.Id.Value,
		})
	}
	return items
}

// tagNameItems 注解 tag 名补全:内置常用 tag + 当前文件已出现的 tag。
func tagNameItems(pf *parsedFile, doc *Document, word string) []CompletionItem {
	builtin := []string{"verify", "db", "marshal"}
	items := make([]CompletionItem, 0, len(builtin))
	for _, name := range builtin {
		items = append(items, CompletionItem{Label: name, Kind: CompletionKeyword, Detail: "注解"})
	}
	// 当前文件与依赖文件中已出现的 tag 名
	seen := map[string]bool{}
	for _, file := range pf.pkg.Files {
		collectTagNames(file, seen)
	}
	for name := range seen {
		if name == "verify" || name == "db" || name == "marshal" {
			continue
		}
		items = append(items, CompletionItem{Label: name, Kind: CompletionKeyword, Detail: "注解(已使用)"})
	}
	return filterItems(items, word)
}

// collectTagNames 收集文件中出现的注解名。
func collectTagNames(file *ast.File, seen map[string]bool) {
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		var tags []*ast.Tag
		switch t := ts.Type.(type) {
		case *ast.DataType:
			tags = t.Tags
		case *ast.ServerType:
			tags = t.Tags
		case *ast.EnumType:
			tags = t.Tags
		}
		for _, tag := range tags {
			if tag.Name != nil {
				seen[tag.Name.Name] = true
			}
		}
	}
}

// tagKeyItems tag 键名补全:内置常用键 + 当前文件已出现的键。
func tagKeyItems(pf *parsedFile, doc *Document, word string) []CompletionItem {
	builtin := []string{
		"regex", "min", "max", // verify
		"key", "index", "order", "table", "where", "limit", "offset", "force",
		"insert", "inserts", "update", "set", "del", "get", "list", "map", "count",
		"rm", "group", "converter", // db
	}
	items := make([]CompletionItem, 0, len(builtin))
	for _, name := range builtin {
		items = append(items, CompletionItem{Label: name, Kind: CompletionConstant, Detail: "注解键"})
	}
	// 当前文件已出现的键
	seen := map[string]bool{}
	for _, file := range pf.pkg.Files {
		collectTagKeyNames(file, seen)
	}
	for name := range seen {
		found := false
		for _, b := range builtin {
			if b == name {
				found = true
				break
			}
		}
		if !found {
			items = append(items, CompletionItem{Label: name, Kind: CompletionConstant, Detail: "注解键(已使用)"})
		}
	}
	return filterItems(items, word)
}

// collectTagKeyNames 收集文件中的注解键名。
func collectTagKeyNames(file *ast.File, seen map[string]bool) {
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		var all [][]*ast.Tag
		switch t := ts.Type.(type) {
		case *ast.DataType:
			all = append(all, t.Tags)
			for _, f := range t.Fields.List {
				all = append(all, f.Tags)
			}
		case *ast.ServerType:
			all = append(all, t.Tags)
			for _, m := range t.Methods {
				all = append(all, m.Tags)
			}
		case *ast.EnumType:
			all = append(all, t.Tags)
			for _, item := range t.Items {
				all = append(all, item.Tags)
			}
		}
		for _, tags := range all {
			for _, tag := range tags {
				for _, kv := range tag.KV {
					if kv.Name != nil {
						seen[kv.Name.Name] = true
					}
				}
			}
		}
	}
}

// languageItems package 语言名。
func languageItems() []CompletionItem {
	return []CompletionItem{
		{Label: "go", Kind: CompletionValue, Detail: "Go 包名"},
		{Label: "dart", Kind: CompletionValue, Detail: "Dart 包名"},
		{Label: "java", Kind: CompletionValue, Detail: "Java 包名"},
		{Label: "ts", Kind: CompletionValue, Detail: "TypeScript 包名"},
	}
}
