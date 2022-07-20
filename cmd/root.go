/*
Copyright © 2022 steffakasid

*/
package cmd

import (
	"fmt"
	"os"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	dryrun    = "dry-run"
	olderthen = "older-then"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awsclean",
	Short: "This tool is intended to be used to cleanup certain AWS services",
	Long: `This tool is intended to be used to cleanup certain AWS services.
	
Right now it supports the following:
  - Amazon Machine Images (AMIs)
  - Elastic Blockstore (EBS) Volumes

Preqrequisites:
  amiclean uses already provided credentials in ~/.aws/credentials also it uses the
  central configuration in ~/.aws/config!`,
}

func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger.SetLevel(logger.DebugLevel)
	cobra.OnInitialize(initConfig)

	peristentFlags := rootCmd.PersistentFlags()

	peristentFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amiclean.yaml)")

	peristentFlags.BoolP(dryrun, "d", false, "If set to true nothing will be deleted. And amiclean will just show what it would do!")
	peristentFlags.StringP(olderthen, "o", "7d", "Set the duration string (e.g 5d, 1w etc.) how old AMIs must be to be deleted. E.g. if set to 7d, AMIs will be delete which are older then 7 days.")

	cobra.CheckErr(viper.BindPFlags(peristentFlags))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".amiclean")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
