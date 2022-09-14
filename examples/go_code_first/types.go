package nogen

import (
	"bytes"
	"context"
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

func (h *Hugo) Generate(ctx context.Context, source FS) (FS, error) {
	return FS{}, nil
}

func (h *Hugo) Deploy(ctx context.Context, source FS) (string, error) {
	return "http://deployed.url/", nil
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
		// we ignore the first arg, as it is the pointer to the method owner
		for j := 1; j < mType.NumIn(); j++ {
			arg := mType.In(j)
			argName := strings.ToLower(arg.Name())
			// FIXME: we need to ignore context.Context only, not other Context types
			if argName == "context" {
				// we ignore context.Context
				continue
			}
			argsIn = append(argsIn, argName)
		}

		var argsOut []string
		for j := 0; j < mType.NumOut(); j++ {
			arg := mType.Out(j)
			argName := strings.ToLower(arg.Name())
			// FIXME: we need to ignore pure error types, not just types named "error"
			if argName == "error" {
				continue
			}
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
		name = strings.ToLower(t.Name())
		pt = reflect.PointerTo(t)
		methods = append(methods, getMethods(pt)...)
		for _, v := range methods {
			methodMap[v.Name] = struct{}{}
		}
	}

	var ptt reflect.Type
	if t.Kind() == reflect.Pointer {
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
	tpl := template.Must(template.New("schema").Funcs(template.FuncMap{"joinOut": joinOut, "formatArgs": formatArgs}).Parse(` {{ .Name }} {
		{{- range .Methods }}
		{{ .Name }}({{ formatArgs ", " .ArgsIn }}) {{- with .ArgsOut }}: {{joinOut ", " .}}{{ end }}
		{{- end }}
}`))

	var b bytes.Buffer
	err := tpl.Execute(&b, m)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func joinOut(sep string, s []string) string {
	if len(s) < 2 {
		return s[0]
	}

	return "(" + strings.Join(s, sep) + ")"
}

func formatArgs(sep string, s []string) string {
	for i, v := range s {
		// we create the arg name + arg type couple
		s[i] = string(v[0]) + ": " + v
	}
	return strings.Join(s, sep)
}
