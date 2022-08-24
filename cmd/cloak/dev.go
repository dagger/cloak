package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dagger/cloak/core"
	"github.com/dagger/cloak/engine"
	"github.com/dagger/cloak/sdk/go/dagger"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use: "dev",
	Run: Dev,
}

func Dev(cmd *cobra.Command, args []string) {
	localDirs := getKVInput(localDirsInput)
	startOpts := &engine.Config{
		LocalDirs:  localDirs,
		DevServer:  devServerPort,
		Workdir:    workdir,
		ConfigPath: configPath,
	}

	err := engine.Start(context.Background(), startOpts, func(ctx context.Context, _ *core.Extension, _ map[string]dagger.FSID) error {
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
