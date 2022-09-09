package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/alpine/gen/alpine"
	"github.com/dagger/cloak/sdk/go/dagger"

	"hugo/gen/core"
)

func (r *hugo) generate(ctx context.Context, src dagger.FSID) (*dagger.Filesystem, error) {

	panic("implement me")

}

func main() {
	err := engine.Start(context.Background(), &engine.Config{
		Workdir:    ".",
		ConfigPath: "./cloak.yaml",
	}, func(ctx engine.Context) error {
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
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
