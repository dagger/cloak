package querybuilder

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	"github.com/udacity/graphb"
)

func Query() *Selection {
	s := &Selection{
		subSelection: make(map[string]*Selection),
	}
	s.root = s
	return s
}

type Selection struct {
	name  string
	alias string
	args  map[string]any

	root         *Selection
	subSelection map[string]*Selection
	bind         interface{}
}

func (s *Selection) SelectAs(alias, name string) *Selection {
	sel := &Selection{
		name:         name,
		root:         s.root,
		alias:        alias,
		subSelection: make(map[string]*Selection),
	}

	fieldKey := name
	if alias != "" {
		fieldKey = alias
	}

	if _, ok := s.subSelection[fieldKey]; ok {
		panic("duplicate selection field")
	}

	s.subSelection[fieldKey] = sel
	return sel
}

func (s *Selection) Select(name string) *Selection {
	return s.SelectAs("", name)
}

func (s *Selection) Arg(name string, value any) *Selection {
	if s.args == nil {
		s.args = map[string]any{}
	}
	s.args[name] = value
	return s
}

func (s *Selection) Bind(v interface{}) *Selection {
	s.bind = v
	return s
}

func (s *Selection) resultName() string {
	if s.alias != "" {
		return s.alias
	}
	return s.name
}

func (s *Selection) buildFields() []*graphb.Field {
	fields := []*graphb.Field{}
	subFields := make([]*Selection, 0, len(s.subSelection))
	for _, sub := range s.subSelection {
		subFields = append(subFields, sub)
	}
	// Sort fields so the query is stable
	sort.Slice(subFields, func(i, j int) bool {
		return subFields[i].resultName() < subFields[j].resultName()
	})

	for _, sub := range subFields {
		field := graphb.MakeField(sub.name)
		fields = append(fields, field)
		if sub.alias != "" {
			field.Alias = sub.alias
		}
		for name, value := range sub.args {
			var (
				arg graphb.Argument
				err error
			)

			if v, ok := value.(graphb.Argument); ok {
				arg = graphb.ArgumentCustomType(name, v)
			} else {
				arg, err = graphb.ArgumentAny(name, value)
				if err != nil {
					panic(err)
				}
			}

			field = field.AddArguments(arg)
		}
		field.Fields = sub.buildFields()
	}

	return fields
}

func (s *Selection) Build() (string, error) {
	q := graphb.MakeQuery(graphb.TypeQuery)
	q.SetFields(s.root.buildFields()...)

	strCh, err := q.StringChan()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return graphb.StringFromChan(strCh), nil
}

func (s *Selection) Unpack(data interface{}) error {
	return s.root.unpack(data)
}

func (s *Selection) unpack(data interface{}) error {
	if s.bind != nil {
		marshalled, err := json.Marshal(data)
		if err != nil {
			return err
		}
		json.Unmarshal(marshalled, s.bind)
	}

	for _, sub := range s.subSelection {
		field := data.(map[string]interface{})[sub.resultName()]
		if err := sub.unpack(field); err != nil {
			return err
		}
	}

	return nil
}
