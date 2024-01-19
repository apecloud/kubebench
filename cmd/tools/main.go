package main

import (
	"github.com/spf13/cobra"

	"github.com/apecloud/kubebench/pkg/tools"
)

func main() {
	rootCmd := &cobra.Command{Use: "tools"}

	rootCmd.AddCommand(tools.NewMongoDBCmd())
	rootCmd.AddCommand(tools.NewMySQLCmd())
	rootCmd.AddCommand(tools.NewPostgreSQLCmd())
	rootCmd.AddCommand(tools.NewRedisCmd())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
