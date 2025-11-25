package cmd

import (
	"goclash-cli/internal/proxy"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		proxy.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}