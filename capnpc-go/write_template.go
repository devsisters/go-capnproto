package main

import (
	"io"
	"strings"
	"text/template"
)

func writeTemplate(w io.Writer, templ string, param interface{}) {
	t, err := template.New("").Parse(strings.TrimSpace(templ))
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, param)
	if err != nil {
		panic(err)
	}
}
