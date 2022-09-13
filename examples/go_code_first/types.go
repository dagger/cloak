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

type Method struct {
	Name    string
	ArgsIn  []string
	ArgsOut []string
}

func getMethods(t reflect.Type) []Method {
	var methods []Method
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mName := strings.ToLower(m.Name)
		// args
		mType := m.Type
		var argsIn []string
		for j := 0; j < mType.NumIn(); j++ {
			arg := mType.In(j)
			argName := strings.ToLower(arg.Name())
			argsIn = append(argsIn, argName)
		}

		var argsOut []string
		for j := 0; j < mType.NumOut(); j++ {
			arg := mType.Out(j)
			argName := strings.ToLower(arg.Name())
			argsOut = append(argsOut, argName)
		}
		methods = append(methods, Method{mName, argsIn, argsOut})
	}
	return methods
}

func Schema(h any) (string, error) {
	t := reflect.TypeOf(h)

	methodMap := map[string]struct{}{}
	methods := getMethods(t)
	for _, v := range methods {
		methodMap[v.Name] = struct{}{}
	}

	var name string
	var pt reflect.Type
	if t.Kind() != reflect.Pointer {
		log.Println("ptr:", false)
		name = strings.ToLower(t.Name())
		pt = reflect.PointerTo(t)
		methods = append(methods, getMethods(pt)...)
		for _, v := range methods {
			methodMap[v.Name] = struct{}{}
		}
	}

	_ = pt
	//name := strings.ToLower(t.Name())

	var ptt reflect.Type
	if t.Kind() == reflect.Pointer {
		log.Println("ptr:", true)
		ptt = t
		name = strings.ToLower(t.Elem().Name())
		pttMethods := getMethods(ptt)
		for _, v := range pttMethods {
			_, ok := methodMap[v.Name]
			if ok {
				continue
			}
			methodMap[v.Name] = struct{}{}
			methods = append(methods, v)
		}
	}

	type method struct {
		Name    string
		Methods []Method
	}

	m := method{name, methods}
	log.Println("name:", name)
	tpl := template.Must(template.New("schema").Funcs(template.FuncMap{"join": func(sep string, s []string) string { return strings.Join(s, sep) }}).Parse(` {{ .Name }} {
		{{ range .Methods -}}
		{{ .Name }}( {{ join "," .ArgsIn }} ) {{join "," .ArgsOut}}
		{{- end }}
	}`))

	var b bytes.Buffer
	err := tpl.Execute(&b, m)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
