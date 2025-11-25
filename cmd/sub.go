package cmd

import (
	"fmt"
	"goclash-cli/internal/config"
	"goclash-cli/internal/subscription"
	"log"

	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscriptions",
}

var subAddCmd = &cobra.Command{
	Use:   "add [url]",
	Short: "Add and update subscription",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Println("[Subscription Agent] Fetching from:", url)
		
		nodes, err := subscription.FetchAndParse(url)
		if err != nil {
			log.Fatalf("Failed to fetch: %v", err)
		}

		err = config.SaveNodes(nodes)
		if err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}

		fmt.Printf("Success! Imported %d nodes to config.yaml\n", len(nodes))
	},
}

func init() {
	subCmd.AddCommand(subAddCmd)
	rootCmd.AddCommand(subCmd)
}