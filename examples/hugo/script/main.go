package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gocloud.dev/blob"
	"golang.org/x/oauth2/google"

	"github.com/Khan/genqlient/graphql"
	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/sdk/go/dagger"
	"github.com/dagger/cloak/testutil"

	"hugo/gen/alpine"
	"hugo/gen/core"
	"hugo/gen/hugo"
)

func main() {
	credsFilePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsFilePath == "" {
		log.Fatal("set GOOGLE_APPLICATION_CREDENTIALS to your location of your GCP service account file")
	}

	credsDirPath, credsFileRelPath := filepath.Split(credsFilePath)
	_ = credsFileRelPath

	err := engine.Start(context.Background(), &engine.Config{
		Workdir:    ".",
		ConfigPath: "./cloak.yaml",
		LocalDirs: map[string]string{
			"credsDir": credsDirPath,
			"src":      "./test",
		},
	}, func(ctx engine.Context) error {
		src, ok := ctx.LocalDirs["src"]
		if !ok {
			return errors.New("missing source dir")
		}

		genResp, err := hugo.Generate(ctx, src)
		if err != nil {
			return fmt.Errorf("hugo generate: %w", err)
		}

		genFSID := genResp.Hugo.Generate.ID

		saJSON := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if err != nil {
			return fmt.Errorf("load secret JSON env: %w", err)
		}
		_ = saJSON

		qOpts := testutil.QueryOptions{
			Secrets: map[string]string{
				"GCP_SERVICE_ACCOUNT_JSON": saJSON,
			},
		}
		err = addSecrets(ctx, ctx.Client, &qOpts) //"GCP_SERVICE_ACCOUNT_JSON")
		if err != nil {
			return fmt.Errorf("add secret: %w", err)
		}

		saFilePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		saJSON2, err := os.ReadFile(saFilePath)
		if err != nil {
			return fmt.Errorf("load secret JSON env: %w", err)
		}

		// use GOOGLE_APPLICATION_CREDENTIALS
		creds, err := google.CredentialsFromJSON(ctx, []byte(saJSON2))
		if err != nil {
			return fmt.Errorf("get GCP creds: %w", err)
		}
		_ = creds

		readSecretOutput, err := core.Secret(ctx, dagger.SecretID(creds.JSON))
		if err != nil {
			return fmt.Errorf("failed to read secret: %w", err)
		}

		hugo.Deploy(ctx, genFSID, "gs://gerhard-cnd-deploy-test", "", dagger.SecretID(readSecretOutput.Core.Secret))

		alp, err := alpine.Build(ctx, []string{"curl", "git"})
		if err != nil {
			return err
		}

		curled, err := core.Exec(ctx, alp.Alpine.Build.ID, core.ExecInput{
			Args: []string{"sh", "-c", "curl -L https://github.com/gohugoio/hugo/releases/download/v0.102.3/hugo_0.102.3_Linux-64bit.tar.gz | tar -xz"},
		})
		if err != nil {
			return fmt.Errorf("install hugo: %w", err)
		}
		_ = curled

		themed, err := core.Exec(ctx, curled.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
			Args: []string{"sh", "-c", "git init . && git submodule init && git submodule add https://github.com/theNewDynamic/gohugo-theme-ananke.git ./mnt/themes/ananke"},
		})
		if err != nil {
			return fmt.Errorf("git submodule add: %w", err)
		}
		_ = themed

		wd, err := core.Workdir(ctx)
		if err != nil {
			return fmt.Errorf("workdir: %w", err)
		}

		wdID := wd.Host.Workdir.Read.ID
		_ = wdID

		generated, err := core.ExecGetMount(ctx, themed.Core.Filesystem.Exec.Fs.ID, core.ExecInput{
			Mounts: []core.MountInput{
				{
					Fs:   wdID,
					Path: "/mnt",
				},
			},
			Workdir: "/mnt/test",
			Args:    []string{"sh", "-c", "/hugo --buildFuture"},
		}, "/mnt")
		if err != nil {
			return fmt.Errorf("hugo generate: %w", err)
		}

		_ = generated

		_, err = core.WriteWorkdir(ctx, generated.Core.Filesystem.Exec.Mount.ID)
		if err != nil {
			return fmt.Errorf("write on host: %w", err)
		}

		bucket, err := blob.OpenBucket(ctx, "gs://gerhard-cnd-deploy-test")
		if err != nil {
			return fmt.Errorf("open bucket: %w", err)
		}
		defer bucket.Close()

		dd, err := os.ReadDir("/mnt/test")
		if err != nil {
			return fmt.Errorf("test open dir: %w", err)
		}

		for _, d := range dd {
			log.Println("LOL:", d.Name())
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func addSecrets(ctx context.Context, cl graphql.Client, opts *testutil.QueryOptions) error {
	for name, plaintext := range opts.Secrets {
		addSecret := struct {
			Core struct {
				AddSecret dagger.SecretID
			}
		}{}
		err := cl.MakeRequest(ctx,
			&graphql.Request{
				Query: `query AddSecret($plaintext: String!) {
					core {
						addSecret(plaintext: $plaintext)
					}
				}`,
				Variables: map[string]string{
					"plaintext": plaintext,
				},
			},
			&graphql.Response{Data: &addSecret},
		)
		if err != nil {
			return err
		}
		opts.Variables[name] = addSecret.Core.AddSecret
	}
	return nil
}
