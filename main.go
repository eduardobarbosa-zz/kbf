package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

func main() {

	var filePath string

	var cmdPwd = &cobra.Command{
		Use:   "connect",
		Short: "connect is a port-forwarding k8s services",
		Long: `connect is used to bind ports from k8s service to host
		like kubectl port-forward works`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dat, err := ioutil.ReadFile("banner")
			if err != nil {
				log.Panic(err.Error())
			}
			fmt.Println(string(dat))
			portForwarding(loadFromFile(filePath))
		},
	}

	cmdPwd.Flags().StringVarP(&filePath, "file", "f", "forward.yml", "forward file path")

	var rootCmd = &cobra.Command{Use: "./kbf"}
	rootCmd.AddCommand(cmdPwd)
	rootCmd.Execute()
}
