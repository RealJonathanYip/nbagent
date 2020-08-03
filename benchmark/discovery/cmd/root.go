package cmd

import (
	"sync"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "discovery",
	Short: "A benchmark tool for discovery",
	Long:  "A benchmark tool for discovery",
}

var (
	szAgents      string
	nTotalClients uint
	nCallTimeout  uint
	bOnlyAgent    bool

	objWaitGroup sync.WaitGroup
)

func init() {
	RootCmd.PersistentFlags().StringVar(&szAgents, "agents", "127.0.0.1:8800", "agent points")
	RootCmd.PersistentFlags().UintVar(&nTotalClients, "clients", 1, "Total number of clients")
	RootCmd.PersistentFlags().UintVar(&nCallTimeout, "call-timeout", 10, "Call timeout for client")
	RootCmd.PersistentFlags().BoolVar(&bOnlyAgent, "only-agent", true, "connect only to the one agent node")
}
