package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/dagger/cloak/sdk/go/dagger"
)

func (r *terraform) apply(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	cmd := []string{"terraform", "apply", "-auto-approve"}
	return tfExec(ctx, config, token, cmd)
}

func (r *terraform) plan(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	cmd := []string{"terraform", "plan"}
	return tfExec(ctx, config, token, cmd)
}

func (r *terraform) fmt(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	cmd := []string{"terraform", "fmt"}
	return tfExec(ctx, config, token, cmd)
}

func (r *terraform) destroy(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	cmd := []string{"terraform", "apply", "-destroy", "-auto-approve"}
	return tfExec(ctx, config, token, cmd)
}

func tfExec(ctx context.Context, config dagger.FSID, token dagger.SecretID, command []string) (*dagger.Filesystem, error) {
	client, err := dagger.Client(ctx)
	if err != nil {
		return nil, err
	}

	fsid, err := image(ctx, client, "hashicorp/terraform:latest")
	if err != nil {
		fmt.Printf("cant load image: %v", err)
		return nil, err
	}

	exec, err := exec(ctx, client, fsid, config, token, command)
	if err != nil {
		fmt.Printf("cant execute plan: %v", err)
		return nil, err
	}

	return &dagger.Filesystem{ID: exec}, nil
}

func image(ctx context.Context, client graphql.Client, ref string) (dagger.FSID, error) {
	req := &graphql.Request{
		Query: `
query Image ($ref: String!) {
	core {
		image(ref: $ref) {
			id
		}
	}
}
`,
		Variables: map[string]any{
			"ref": ref,
		},
	}
	resp := struct {
		Core struct {
			Image struct {
				ID dagger.FSID
			}
		}
	}{}
	err := client.MakeRequest(ctx, req, &graphql.Response{Data: &resp})
	if err != nil {
		return "", err
	}

	return resp.Core.Image.ID, nil
}

func exec(ctx context.Context, client graphql.Client, root dagger.FSID, mount dagger.FSID, token dagger.SecretID, args []string) (dagger.FSID, error) {
	flatArgs := "\""
	flatArgs = flatArgs + strings.Join(args, "\", \"")
	flatArgs = flatArgs + "\""

	req := &graphql.Request{
		Query: `
query TfExec ($root: FSID!, $mount: FSID!, $args: String!, $token: SecretID!) {
	core {
		filesystem(id: $root) {
			exec(input: {
				args: [$args],
				workdir: "/src",
				mounts: [
					{
						path: "/src",
						fs: $mount
					}
				],
				secretEnv: [
					{
						name: "TF_TOKEN_app_terraform_io",
						id: $token
					}
				]
			}) {
				fs {
					id
				}
			}
		}
	}
}
`,
		Variables: map[string]any{
			"root":  root,
			"mount": mount,
			"args":  flatArgs,
			"token": token,
		},
	}
	resp := struct {
		Core struct {
			Filesystem struct {
				Exec struct {
					FS struct {
						ID dagger.FSID
					}
				}
			}
		}
	}{}
	err := client.MakeRequest(ctx, req, &graphql.Response{Data: &resp})
	if err != nil {
		return "", err
	}

	return resp.Core.Filesystem.Exec.FS.ID, nil
}
