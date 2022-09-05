package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/wp-to-hugo/gen/core"
)

// cloak generate
func main() {
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx engine.Context) error {
		workdirResp, err := core.Workdir(ctx)
		if err != nil {
			return err
		}

		// Start from Alpine, mount workdir, install wget
		image, err := core.Image(ctx, "alpine:3.15")
		if err != nil {
			return err
		}

		fs := &image.Core.Image

		output, err := core.Exec(ctx, fs.ID, core.ExecInput{
			Args:    []string{"apk", "add", "-U", "--no-cache", "wget", "bash"},
			Workdir: "/mnt",
		})
		if err != nil {
			return fmt.Errorf("failed to installs: %s", err)
		}
		fs = output.Core.Filesystem.Exec.Fs

		output, err = core.Exec(ctx, fs.ID, core.ExecInput{
			Args:    []string{"./run.sh"},
			Workdir: "/mnt",
			Mounts: []core.MountInput{
				core.MountInput{Fs: workdirResp.Host.Workdir.Read.ID, Path: "/mnt/"},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to run script: %s", err)
		}
		fs = output.Core.Filesystem.Exec.Fs

		return nil
	}); err != nil {
		panic(err)
	}
}
