package core

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/dagger/cloak/core/filesystem"
	"github.com/dagger/cloak/extension"
	"github.com/dagger/cloak/router"
	"github.com/graphql-go/graphql"
	"golang.org/x/sync/singleflight"
)

// TODO:(sipsma) a lot of the naming of the go types/methods/files doesn't make any sense anymore, need to update w/ project terminology

type Project struct {
	Name         string
	Schema       string
	Operations   string
	Sources      []*extension.Source
	Dependencies []*Project
	schema       *extension.RemoteSchema // internal-only, for convenience in `install` resolver
}

var _ router.ExecutableSchema = &extensionSchema{}

type extensionSchema struct {
	*baseSchema
	compiledSchemas map[string]*extension.CompiledRemoteSchema
	l               sync.RWMutex
	sf              singleflight.Group
	sshAuthSockID   string
}

func (s *extensionSchema) Name() string {
	return "extension"
}

func (s *extensionSchema) Schema() string {
	return `
	"Project representation"
	type Project {
		"name of the project"
		name: String!

		"merged schema of the project's sources (if any)"
		schema: String

		"merged operations of the project's sources (if any)"
		operations: String

		"sources for this project"
		sources: [ProjectSource!]

		"dependencies for this project"
		dependencies: [Project!]

		"install the project, stitching its schema into the API"
		install: Boolean!
	}

	"Project source representation"
	type ProjectSource {
		"path to the source in the project filesystem"
		path: String!

		"schema associated with the source (if any)"
		schema: String

		"operations associated with the source (if any)"
		operations: String

		"sdk of the source"
		sdk: String!
	}

	extend type Filesystem {
		"load a project's metadata"
		loadProject(configPath: String!): Project!
	}

	extend type Core {
		"Look up a project by name"
		project(name: String!): Project!
	}
	`
}

func (s *extensionSchema) Operations() string {
	return ""
}

func (s *extensionSchema) Resolvers() router.Resolvers {
	return router.Resolvers{
		"Filesystem": router.ObjectResolver{
			"loadProject": s.loadProject,
		},
		"Core": router.ObjectResolver{
			"project": s.project,
		},
		"Project": router.ObjectResolver{
			"install": s.install,
		},
	}
}

func (s *extensionSchema) Dependencies() []router.ExecutableSchema {
	return nil
}

func (s *extensionSchema) install(p graphql.ResolveParams) (any, error) {
	obj := p.Source.(*Project)

	executableSchema, err := obj.schema.Compile(p.Context, s.compiledSchemas, &s.l, &s.sf)
	if err != nil {
		return nil, err
	}

	if err := s.router.Add(executableSchema); err != nil {
		return nil, err
	}

	return true, nil
}

func (s *extensionSchema) loadProject(p graphql.ResolveParams) (any, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf(string(debug.Stack()), err)
			panic(err)
		}
	}()
	obj, err := filesystem.FromSource(p.Source)
	if err != nil {
		return nil, err
	}

	configPath := p.Args["configPath"].(string)
	schema, err := extension.Load(p.Context, s.gw, s.platform, obj, configPath, s.sshAuthSockID)
	if err != nil {
		return nil, err
	}

	return remoteSchemaToProject(schema), nil
}

func (s *extensionSchema) project(p graphql.ResolveParams) (any, error) {
	name := p.Args["name"].(string)

	schema := s.router.Get(name)
	if schema == nil {
		return nil, fmt.Errorf("project %q not found", name)
	}

	return routerSchemaToProject(schema), nil
}

// TODO:(sipsma) guard against infinite recursion
func routerSchemaToProject(schema router.ExecutableSchema) *Project {
	ext := &Project{
		Name:       schema.Name(),
		Schema:     schema.Schema(),
		Operations: schema.Operations(),
		//FIXME:(sipsma) SDK is not exposed on router.ExecutableSchema yet
	}
	for _, dep := range schema.Dependencies() {
		ext.Dependencies = append(ext.Dependencies, routerSchemaToProject(dep))
	}
	return ext
}

// TODO:(sipsma) guard against infinite recursion
func remoteSchemaToProject(schema *extension.RemoteSchema) *Project {
	ext := &Project{
		Name:       schema.Name(),
		Schema:     schema.Schema(),
		Operations: schema.Operations(),
		Sources:    schema.Sources(),
		schema:     schema,
	}
	for _, dep := range schema.Dependencies() {
		ext.Dependencies = append(ext.Dependencies, remoteSchemaToProject(dep))
	}
	return ext
}
