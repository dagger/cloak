package main

import (
	"context"

	"github.com/dagger/cloak/examples/alpine/gen/core"
	"github.com/dagger/cloak/sdk/go/dagger"
)

func (r *terraform) apply(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	return tfExec(ctx, config, token, "plan")
}

func (r *terraform) plan(ctx context.Context, config dagger.FSID, token dagger.SecretID) (*dagger.Filesystem, error) {
	return tfExec(ctx, config, token, "plan")
}

func tfExec(ctx context.Context, config dagger.FSID, token dagger.SecretID, command string) (*dagger.Filesystem, error) {
	fs := &dagger.Filesystem{}

	tf, err := core.Image(ctx, "hashicorp/terraform:latest")
	if err != nil {
		return fs, err
	}

	exec, err := core.Exec(ctx, tf.Core.Image.ID, core.ExecInput{
		Args:    []string{command},
		Workdir: "/src",
		Mounts: []core.MountInput{
			{
				Fs:   config,
				Path: "/src",
			},
		},
		SecretEnv: []core.ExecSecretEnvInput{
			{
				Name: "TF_TOKEN_app_terraform_io",
				Id:   token,
			},
		},
	})
	return exec.Core.Filesystem.Exec.Fs, err
}
