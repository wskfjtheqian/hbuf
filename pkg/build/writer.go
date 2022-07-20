package build

import "strings"

type Writer struct {
	imp  map[string]struct{}
	code *strings.Builder
	Path string
	Pack string
	lang map[string]*Language
}

func (w *Writer) Import(text string) {
	w.imp[text] = struct{}{}
}

func (w *Writer) Code(text string) {
	_, _ = w.code.WriteString(text)
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

func (w *Writer) GetImports() map[string]struct{} {
	return w.imp
}

func (w *Writer) AddImports(imp map[string]struct{}) {
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

func NewWriter(pack string) *Writer {
	return &Writer{
		imp:  map[string]struct{}{},
		code: &strings.Builder{},
		Pack: pack,
		lang: map[string]*Language{},
	}
}
