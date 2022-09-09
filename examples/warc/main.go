package main

import (
	"context"
	"fmt"

	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/examples/warc/gen/core"
)

func main() {
	if err := engine.Start(context.Background(), &engine.Config{}, func(ctx engine.Context) error {
		// Configure current directory for mounting
		workdirResp, err := core.Workdir(ctx)
		if err != nil {
			return err
		}
		workdirFs := workdirResp.Host.Workdir.Read.ID

		// Start with the Alpine image
		image, err := core.Image(ctx, "alpine:3.15")
		if err != nil {
			return fmt.Errorf("failed to resolve image: %s", err)
		}

		alpineFs := &image.Core.Image

		// Install deps
		deps, err := core.Exec(ctx, alpineFs.ID, core.ExecInput{
			Args: []string{"apk", "add", "-U", "--no-cache", "bash", "git", "wget"},
		})
		if err != nil {
			return fmt.Errorf("failed to install: %s", err)
		}
		depsFs := deps.Core.Filesystem.Exec.Fs

		// Generate Web ARChive
		webArchive, err := core.ExecGetMount(ctx, depsFs.ID, core.ExecInput{
			Args:    []string{"./webarchive.sh"},
			Workdir: "/mnt",
			Mounts: []core.MountInput{
				core.MountInput{Fs: workdirFs, Path: "/mnt"},
			},
		}, "/mnt")
		if err != nil {
			return fmt.Errorf("failed to generate Web ARChive: %s", err)
		}
		webArchiveFs := webArchive.Core.Filesystem.Exec.Fs

		// Git commit & push content
		_, err = core.ExecGetMount(ctx, depsFs.ID, core.ExecInput{
			Args:    []string{"ls", "-lah"},
			Workdir: "/mnt",
			Mounts: []core.MountInput{
				core.MountInput{Fs: webArchiveFs.ID, Path: "/mnt/"},
			},
		}, "/mnt")
		if err != nil {
			return fmt.Errorf("failed to git commit: %s", err)
		}

		// MAYBE export downloaded content back to mounted directory
		// _, err = core.WriteWorkdir(ctx, webArchiveFs.ID)
		// if err != nil {
		// 	return fmt.Errorf("failed to export Web ARChive: %s", err)
		// }

		return nil
	}); err != nil {
		panic(err)
	}
}
