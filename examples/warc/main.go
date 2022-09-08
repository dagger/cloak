package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/warc/gen/core"
)

func main() {
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx engine.Context) error {
		workdirResp, err := core.Workdir(ctx)
		if err != nil {
			return err
		}

		// Start from Alpine
		image, err := core.Image(ctx, "alpine:3.15")
		if err != nil {
			return err
		}

		fs := &image.Core.Image

		// Install deps & declare workdir
		deps, err := core.Exec(ctx, fs.ID, core.ExecInput{
			Args:    []string{"apk", "add", "-U", "--no-cache", "wget", "bash"},
			Workdir: "/mnt",
		})
		if err != nil {
			return fmt.Errorf("failed to install: %s", err)
		}
		fs = deps.Core.Filesystem.Exec.Fs

		// Mount current directory & run script
		webArchive, err := core.ExecGetMount(ctx, fs.ID, core.ExecInput{
			Args:    []string{"./run.sh"},
			Workdir: "/mnt",
			Mounts: []core.MountInput{
				core.MountInput{Fs: workdirResp.Host.Workdir.Read.ID, Path: "/mnt/"},
			},
		}, "/mnt")
		if err != nil {
			return fmt.Errorf("failed to run script: %s", err)
		}
		fs = webArchive.Core.Filesystem.Exec.Mount

		// Export downloaded content back to mounted directory
		_, err = core.WriteWorkdir(ctx, fs.ID)
		if err != nil {
			return fmt.Errorf("failed to export downloaded content: %s", err)
		}

		return nil
	}); err != nil {
		panic(err)
	}
}
