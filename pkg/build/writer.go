package build

import (
	"hbuf/pkg/ast"
	"strings"
)

type Writer struct {
	imp      map[string]map[string]uint
	code     *strings.Builder
	File     *ast.File
	Packages string
	lang     map[string]*Language
	maps     map[string]interface{}
}

func (w *Writer) SetValue(key string, val interface{}) {
	w.maps[key] = val
}

func (w *Writer) GetValue(key string) (interface{}, bool) {
	val, ok := w.maps[key]
	return val, ok
}

func (w *Writer) GetImport(pkg string) map[string]uint {
	return w.imp[pkg]
}

func (w *Writer) Import(pkg string, name ...string) {
	if as, ok := w.imp[pkg]; ok {
		//if level > as.level {
		for _, item := range name {
			as[item] = 0
		}
		//}
		return
	}
	w.imp[pkg] = map[string]uint{}
	for _, item := range name {
		w.imp[pkg][item] = 0
	}
}

func (w *Writer) Code(text string) *Writer {
	_, _ = w.code.WriteString(text)
	return w
}

func (w *Writer) LF() *Writer {
	_, _ = w.code.WriteString("\n")
	return w
}

func (w *Writer) Tab(num int) *Writer {
	for i := 0; i < num; i++ {
		_, _ = w.code.WriteString("\t")
	}
	return w
}

func (w *Writer) String() string {
	return w.code.String()
}

func (w *Writer) ImportByWriter(value *Writer) {
	for key, val := range value.imp {
		w.imp[key] = val
	}
}

func (w *Writer) GetCode() *strings.Builder {
	return w.code

}

func (w *Writer) GetImports() map[string]map[string]uint {
	return w.imp
}

func (w *Writer) AddImports(imp map[string]map[string]uint) {
	for key, val := range imp {
		w.imp[key] = val
	}
}

func (w *Writer) GetLang(name string) *Language {
	if val, ok := w.lang[name]; ok {
		return val
	}
	lang := NewLanguage(name)
	w.lang[name] = lang
	return lang
}

func (w *Writer) GetLangMap() map[string]*Language {
	return w.lang
}

func NewWriter() *Writer {
	return &Writer{
		imp:  map[string]map[string]uint{},
		code: &strings.Builder{},
		lang: map[string]*Language{},
		maps: map[string]interface{}{},
	}
}
