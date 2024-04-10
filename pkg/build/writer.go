package build

import (
	"hbuf/pkg/ast"
	"strings"
)

type Writer struct {
	imp      map[string]string
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

func (w *Writer) Import(text string, s string) string {
	if as, ok := w.imp[text]; ok {
		return as
	}
	w.imp[text] = s
	return s
}

func (w *Writer) Code(text string) *Writer {
	_, _ = w.code.WriteString(text)
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

func (w *Writer) GetImports() map[string]string {
	return w.imp
}

func (w *Writer) AddImports(imp map[string]string) {
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

func (w *Writer) GetLangs() map[string]*Language {
	return w.lang
}

func NewWriter() *Writer {
	return &Writer{
		imp:  map[string]string{},
		code: &strings.Builder{},
		lang: map[string]*Language{},
		maps: map[string]interface{}{},
	}
}
