package main

import (
	_ "embed"
	"fmt"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen"
	gqlconfig "github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/Khan/genqlient/generate"
	"github.com/dagger/cloak/core"
	"github.com/dagger/cloak/extension"
	"github.com/vektah/gqlparser/v2/ast"
)

//go:embed templates/go.main.gotpl
var mainTmpl string

//go:embed templates/go.generated.gotpl
var generatedTmpl string

func generateGoWorkflowStub() error {
	if err := os.WriteFile(filepath.Join(generateOutputDir, "main.go"), []byte(workflowMain), 0644); err != nil {
		return err
	}

	cmd := exec.Command("go", "mod", "edit", "-replace=github.com/docker/docker=github.com/docker/docker@v20.10.3-0.20220414164044-61404de7df1a+incompatible")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	// FIXME:(sipsma) don't hardcode this
	cmd = exec.Command("go", "mod", "edit", "-replace=github.com/dagger/cloak=github.com/sipsma/cloak@workflow-clean")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	// this tidy has to run first to resolve "workflow-clean" (otherwise next commands whine)
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("go", "get", "github.com/dagger/cloak")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("go", "mod", "vendor")
	cmd.Dir = generateOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Env = append(cmd.Env, "GOPRIVATE=github.com/sipsma/cloak,github.com/dagger/cloak")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func generateGoExtensionStub(ext *extension.Source, coreProj *core.Project) error {
	cfg := gqlconfig.DefaultConfig()
	cfg.SkipModTidy = true
	cfg.Exec = gqlconfig.ExecConfig{Filename: filepath.Join(generateOutputDir, "_deleteme.go"), Package: "main"}
	cfg.SchemaFilename = nil
	cfg.Sources = []*ast.Source{{Input: ext.Schema}}
	cfg.Model = gqlconfig.PackageConfig{
		Filename: filepath.Join(generateOutputDir, "models.go"),
		Package:  "main",
	}
	cfg.Models = gqlconfig.TypeMap{
		"SecretID": gqlconfig.TypeMapEntry{
			Model: gqlconfig.StringList{"github.com/dagger/cloak/sdk/go/dagger.SecretID"},
		},
		"FSID": gqlconfig.TypeMapEntry{
			Model: gqlconfig.StringList{"github.com/dagger/cloak/sdk/go/dagger.FSID"},
		},
		"Filesystem": gqlconfig.TypeMapEntry{
			Model: gqlconfig.StringList{"github.com/dagger/cloak/sdk/go/dagger.Filesystem"},
			Fields: map[string]gqlconfig.TypeMapField{
				"exec":        {Resolver: false},
				"dockerbuild": {Resolver: false},
				"file":        {Resolver: false},
			},
		},
		"Exec": gqlconfig.TypeMapEntry{
			Model: gqlconfig.StringList{"github.com/dagger/cloak/sdk/go/dagger.Exec"},
			Fields: map[string]gqlconfig.TypeMapField{
				"fs":       {Resolver: false},
				"stdout":   {Resolver: false},
				"stderr":   {Resolver: false},
				"exitcode": {Resolver: false},
				"mount":    {Resolver: false},
			},
		},
	}

	if err := gqlconfig.CompleteConfig(cfg); err != nil {
		return fmt.Errorf("error completing config: %w", err)
	}
	defer os.Remove(cfg.Exec.Filename)
	if err := api.Generate(cfg, api.AddPlugin(plugin{
		mainPath:      filepath.Join(generateOutputDir, "main.go"),
		generatedPath: filepath.Join(generateOutputDir, "generated.go"),
		coreSchema:    coreProj.Schema,
	})); err != nil {
		return fmt.Errorf("error generating code: %w", err)
	}
	return nil
}

func generateGoClientStubs(subdir string) error {
	cfg := &generate.Config{
		Schema:     generate.StringList{"schema.graphql"},
		Operations: generate.StringList{"operations.graphql"},
		Generated:  "generated.go",
		Bindings: map[string]*generate.TypeBinding{
			"Filesystem": {Type: "github.com/dagger/cloak/sdk/go/dagger.Filesystem"},
			"Exec":       {Type: "github.com/dagger/cloak/sdk/go/dagger.Exec"},
			"FSID":       {Type: "github.com/dagger/cloak/sdk/go/dagger.FSID"},
			"SecretID":   {Type: "github.com/dagger/cloak/sdk/go/dagger.SecretID"},
		},
		ClientGetter: "github.com/dagger/cloak/sdk/go/dagger.Client",
	}
	if err := cfg.ValidateAndFillDefaults(subdir); err != nil {
		return err
	}
	generated, err := generate.Generate(cfg)
	if err != nil {
		return err
	}
	for filename, content := range generated {
		if err := os.WriteFile(filename, content, 0600); err != nil {
			return err
		}
	}
	return nil
}

type plugin struct {
	mainPath      string
	generatedPath string
	coreSchema    string
}

func (plugin) Name() string {
	return "cloakgen"
}

func (p plugin) InjectSourceEarly() *ast.Source {
	return &ast.Source{BuiltIn: true, Input: p.coreSchema}
}

func (p plugin) GenerateCode(data *codegen.Data) error {
	file := File{}

	typesByName := make(map[string]types.Type)
	for _, o := range data.Objects {
		if o.Name == "Query" {
			// only include fields under query from the current schema, not any external imported ones like `core`
			var queryFields []*codegen.Field
			for _, f := range o.Fields {
				if !f.TypeReference.Definition.BuiltIn {
					queryFields = append(queryFields, f)
				}
			}
			o.Fields = queryFields
		} else if o.BuiltIn || o.IsReserved() {
			continue
		}
		var hasResolvers bool
		for _, f := range o.Fields {
			if !f.IsReserved() {
				hasResolvers = true
			}
		}
		if !hasResolvers {
			continue
		}
		file.Objects = append(file.Objects, o)
		typesByName[o.Reference().String()] = o.Reference()
		for _, f := range o.Fields {
			f.MethodHasContext = true
			resolver := Resolver{o, f, "", ""}
			file.Resolvers = append(file.Resolvers, &resolver)
			typesByName[f.TypeReference.GO.String()] = f.TypeReference.GO
			for _, arg := range f.Args {
				typesByName[arg.TypeReference.GO.String()] = arg.TypeReference.GO
			}
		}
	}

	resolverBuild := &ResolverBuild{
		File:        &file,
		PackageName: "main",
		HasRoot:     true,
		typesByName: typesByName,
	}

	if err := templates.Render(templates.Options{
		PackageName: "main",
		Filename:    p.mainPath,
		Data:        resolverBuild,
		Packages:    data.Config.Packages,
		Template:    mainTmpl,
	}); err != nil {
		return err
	}

	if err := templates.Render(templates.Options{
		PackageName:     "main",
		Filename:        p.generatedPath,
		Data:            resolverBuild,
		Packages:        data.Config.Packages,
		Template:        generatedTmpl,
		GeneratedHeader: true,
	}); err != nil {
		return err
	}

	return nil
}

type ResolverBuild struct {
	*File
	HasRoot      bool
	PackageName  string
	ResolverType string
	typesByName  map[string]types.Type
}

func (r ResolverBuild) ShortTypeName(name string) string {
	shortName := templates.CurrentImports.LookupType(r.typesByName[name])
	if shortName == "*<nil>" || shortName == "<nil>" {
		return ""
	}
	return shortName
}

func (r ResolverBuild) PointedToShortTypeName(name string) string {
	t, ok := r.typesByName[name].(*types.Pointer)
	if !ok {
		return ""
	}
	shortName := templates.CurrentImports.LookupType(t.Elem())
	if shortName == "*<nil>" || shortName == "<nil>" {
		return ""
	}
	return shortName
}

type File struct {
	// These are separated because the type definition of the resolver object may live in a different file from the
	// resolver method implementations, for example when extending a type in a different graphql schema file
	Objects         []*codegen.Object
	Resolvers       []*Resolver
	RemainingSource string
}

type Resolver struct {
	Object         *codegen.Object
	Field          *codegen.Field
	Comment        string
	Implementation string
}

func (r *Resolver) HasArgs() bool {
	return len(r.Field.Args) > 0
}

func (r *Resolver) IncludeParentObject() bool {
	return !r.HasArgs() && !r.Object.Root
}

func (r *Resolver) MethodSignature() string {
	if r.Object.Kind == ast.InputObject {
		return fmt.Sprintf("(ctx context.Context, obj %s, data %s) error",
			templates.CurrentImports.LookupType(r.Object.Reference()),
			templates.CurrentImports.LookupType(r.Field.TypeReference.GO),
		)
	}

	res := "(ctx context.Context"

	if r.IncludeParentObject() {
		res += fmt.Sprintf(", obj %s", templates.CurrentImports.LookupType(r.Object.Reference()))
	}
	for _, arg := range r.Field.Args {
		res += fmt.Sprintf(", %s %s", arg.VarName, templates.CurrentImports.LookupType(arg.TypeReference.GO))
	}

	result := templates.CurrentImports.LookupType(r.Field.TypeReference.GO)
	res += fmt.Sprintf(") (%s, error)", result)
	return res
}

// FIXME:(sipsma)
const workflowMain = `package main

import (
  "context"
  coretypes "github.com/dagger/cloak/core"
  "github.com/dagger/cloak/engine"
  "github.com/dagger/cloak/sdk/go/dagger"
)

func main() {
  if err := engine.Start(context.Background(), &engine.Config{}, func(ctx context.Context, _ *coretypes.Project, _ map[string]dagger.FSID) error {
    panic("implement me")
  }); err != nil {
    panic(err)
  }
}
`
