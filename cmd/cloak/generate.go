package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Khan/genqlient/graphql"
	"github.com/dagger/cloak/core"
	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/extension"
	"github.com/dagger/cloak/sdk/go/dagger"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	Run: Generate,
}

func Generate(cmd *cobra.Command, args []string) {
	sourceDir, err := filepath.Rel(workdir, generateOutputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	startOpts := &engine.Config{
		Workdir:     workdir,
		ConfigPath:  configPath,
		SkipInstall: true,
	}

	if err := engine.Start(context.Background(), startOpts, func(ctx context.Context, proj *core.Project, localDirs map[string]dagger.FSID) error {
		var source *extension.Source
		for _, s := range proj.Sources {
			if s.Path == sourceDir {
				source = s
				break
			}
		}
		if source == nil {
			return fmt.Errorf("no source found at %s", sourceDir)
		}

		cl, err := dagger.Client(ctx)
		if err != nil {
			return err
		}

		coreProj, err := loadCore(ctx, cl)
		if err != nil {
			return err
		}

		if generateExtension {
			switch source.SDK {
			case "go":
				if err := generateGoExtensionStub(source, coreProj); err != nil {
					return err
				}
			case "":
			default:
				return fmt.Errorf("unknown sdk type for extension stub %s", source.SDK)
			}
		}

		if generateWorkflow {
			switch source.SDK {
			case "go":
				if err := generateGoWorkflowStub(); err != nil {
					return err
				}
			case "":
			default:
				return fmt.Errorf("unknown sdk type for workflow stub %s", source.SDK)
			}
		}

		if generateClients {
			for _, dep := range append(proj.Dependencies, coreProj) {
				subdir := filepath.Join(generateOutputDir, "gen", dep.Name)
				if err := os.MkdirAll(subdir, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subdir, ".gitattributes"), []byte("** linguist-generated=true"), 0600); err != nil {
					return err
				}
				schemaPath := filepath.Join(subdir, "schema.graphql")

				// TODO:(sipsma) ugly hack to make each schema/operation work independently when referencing core types.
				fullSchema := dep.Schema
				if dep.Name != "core" {
					fullSchema = coreProj.Schema + "\n\n" + fullSchema
				}
				if err := os.WriteFile(schemaPath, []byte(fullSchema), 0600); err != nil {
					return err
				}
				operationsPath := filepath.Join(subdir, "operations.graphql")
				if err := os.WriteFile(operationsPath, []byte(dep.Operations), 0600); err != nil {
					return err
				}

				switch source.SDK {
				case "go":
					if err := generateGoClientStubs(subdir); err != nil {
						return err
					}
				case "":
				default:
					fmt.Fprintf(os.Stderr, "Error: unknown sdk type %s\n", source.SDK)
					os.Exit(1)
				}
			}
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadCore(ctx context.Context, cl graphql.Client) (*core.Project, error) {
	data := struct {
		Core struct {
			Project core.Project
		}
	}{}
	resp := &graphql.Response{Data: &data}

	err := cl.MakeRequest(ctx,
		&graphql.Request{
			Query: `
			query {
				core {
					project(name: "core") {
						name
						schema
						operations
					}
				}
			}`,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}
	return &data.Core.Project, nil
}
