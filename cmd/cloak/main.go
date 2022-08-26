package main

import (
	"fmt"
	"os"

	"github.com/dagger/cloak/tracing"
	"github.com/spf13/cobra"
)

var (
	configPath string
	workdir    string

	queryFile      string
	queryVarsInput []string
	localDirsInput []string
	secretsInput   []string

	generateOutputDir string
	sdkType           string // TODO: enum?
	generateClients   bool
	generateExtension bool
	generateWorkflow  bool

	devServerPort int
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "project", "p", "", "cloak config file")
	rootCmd.PersistentFlags().StringVar(&workdir, "workdir", "", "workdir as passed to workflows")
	rootCmd.AddCommand(
		doCmd,
		generateCmd,
		devCmd,
	)

	doCmd.Flags().StringVarP(&queryFile, "file", "f", "", "query file")
	doCmd.Flags().StringSliceVarP(&queryVarsInput, "set", "s", []string{}, "query variable")
	doCmd.Flags().StringSliceVarP(&localDirsInput, "local-dir", "l", []string{}, "local directory to import")
	doCmd.Flags().StringSliceVarP(&secretsInput, "secret", "e", []string{}, "secret to import")

	generateCmd.Flags().StringVar(&generateOutputDir, "output-dir", "./", "output directory")
	generateCmd.Flags().BoolVar(&generateClients, "client", true, "generate client stub code")
	generateCmd.Flags().BoolVar(&generateExtension, "extension", false, "generate implementation skeleton code for extension")
	generateCmd.Flags().BoolVar(&generateWorkflow, "workflow", false, "generate implementation skeleton code for workflow")

	devCmd.Flags().IntVar(&devServerPort, "port", 8080, "dev server port")
	devCmd.Flags().StringSliceVarP(&localDirsInput, "local-dir", "l", []string{}, "local directory to import")
	devCmd.Flags().StringSliceVarP(&secretsInput, "secret", "e", []string{}, "secret to import")
}

var rootCmd = &cobra.Command{
	Use: "cloak",
}

func main() {
	closer := tracing.Init()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		closer.Close()
		os.Exit(1)
	}
	closer.Close()
}
