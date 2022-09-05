package nogen

import (
	"bytes"
	"context"
	"log"
	"reflect"
	"strings"
	"text/template"
)

type Hugo struct {
}

// FS should be actually part of cloak API
type FS struct {
	Path string
}

func (h *Hugo) Generate(ctx context.Context, source FS) FS {
	return FS{}
}

func Schema(h any) (string, error) {
	t := reflect.TypeOf(h)
	pt := reflect.PointerTo(t)
	if t.Kind() != reflect.Pointer {
		log.Println("ptr:", false)
		t = reflect.TypeOf(&h)
	}
	if t.Kind() != reflect.Pointer {
		log.Println("ptr2:", false)
		t = reflect.TypeOf(&h)
	}
	_ = pt
	name := strings.ToLower(t.Name())

	var methods []string
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mName := strings.ToLower(m.Name)
		methods = append(methods, mName)
	}

	type method struct {
		Name    string
		Methods []string
	}

	m := method{name, methods}
	log.Println("name:", name)
	tpl := template.Must(template.New("schema").Parse(` {{ .Name }} {
		{{ range .Methods -}}
		{{ . }}(/*TODO*/ src: FSID!) Filesystem!
		{{- end }}
	}`))

	var b bytes.Buffer
	err := tpl.Execute(&b, m)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
