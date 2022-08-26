package extension

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name         string        `yaml:"name"`
	Dependencies []*Dependency `yaml:"dependencies,omitempty"`
	Sources      []*Source     `yaml:"sources,omitempty"`
}

type Source struct {
	Path string `yaml:"path"`
	SDK  string `yaml:"sdk" json:"sdk"`
	// FIXME:(sipsma) convenient for internal use, should not be settable in yaml (yet?)
	Schema     string `yaml:"-"`
	Operations string `yaml:"-"`
}

type Dependency struct {
	Local string         `yaml:"local,omitempty"`
	Git   *GitDependency `yaml:"git,omitempty"`
}

type GitDependency struct {
	Remote string `yaml:"remote,omitempty"`
	Ref    string `yaml:"ref,omitempty"`
	Path   string `yaml:"path,omitempty"`
}

func ParseConfig(data []byte) (*Config, error) {
	cfg := Config{}
	if err := yaml.UnmarshalStrict(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w: %s", err, string(data))
	}
	return &cfg, nil
}
