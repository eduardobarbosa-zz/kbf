package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	filePath string
	rootCmd  = &cobra.Command{Use: "kbf"}

	cmdPwd = &cobra.Command{
		Use:   "port-forward",
		Short: "port-forward to k8s services using yml",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			f, err := loadFromFile(filePath)
			if err != nil {
				showErrorAndExit(cmd, err)
			}
			err = portForwarding(f)
			if err != nil {
				showErrorAndExit(cmd, err)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	cmdPwd.Flags().StringVarP(&filePath, "file", "f", "forward.yml", "forward file path")
	rootCmd.AddCommand(cmdPwd)
}

func showErrorAndExit(cmd *cobra.Command, err error) {
	fmt.Println("Error: " + err.Error())
	fmt.Println(cmd.UsageString())
	os.Exit(1)
}

func initConfig() {
	fmt.Println(`
 _  ______  _____
| |/ | __ )|  ___|
| ' /|  _ \| |_
| . \| |_) |  _|
|_|\_|____/|_|

Kubernetes yaml port-forward ;)
`)
}
