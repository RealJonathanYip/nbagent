package main

import (
	"fmt"
	"git.yayafish.com/nbagent/tools/NB_Cli/cmd"
	"os"
)

func init() {
	//_ = log.SetLogLevel("info")
}

func Run(stopCh chan bool) {
	for {
		select {
		case <-stopCh:
			{
				return
			}
		}
	}
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	//stopCh := make(chan bool, 1)
	//Run(stopCh)
}
