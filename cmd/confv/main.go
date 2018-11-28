package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/dunjut/confv/pkg/flexvolume"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "confv",
		Short:        "confv is a volume plugin for dynamic config rendering in Kubernetes.",
		Long:         "confv is a volume plugin for dynamic config rendering in Kubernetes.",
		SilenceUsage: true,
	}

	flexvolume.AddCobraCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
