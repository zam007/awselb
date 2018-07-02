package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var PassWord string

var rootCmd = &cobra.Command{
	Use:   "awselb",
	Short: "delete or reg ec2 to aws elb service",
	Long:  ``,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}



func init() {
	// add global flags on the root
	rootCmd.PersistentFlags().StringVarP(&PassWord, "password", "p", "", "You need an authorisation to use this app (required)")

	// required flag
	rootCmd.MarkFlagRequired("password")
}
