/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
)

// Constants used in command flags
const (
	dryrunFlag    = "dry-run"
	olderthenFlag = "older-then"
	debugFlag     = "debug"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awsclean",
	Short: "This tool is intended to be used to cleanup certain AWS services",
	Long: `This tool is intended to be used to cleanup certain AWS services.
	
Right now it supports the following:
  - Unused Amazon Machine Images (AMIs)
  - Unused Elastic Blockstore (EBS) Volumes
  - Unused SecurityGroups

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

	internal.InitLogger()

	cobra.OnInitialize(initConfig)

	peristentFlags := rootCmd.PersistentFlags()

	peristentFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amiclean.yaml)")

	peristentFlags.BoolP(dryrunFlag, "d", false, "If set to true nothing will be deleted. And amiclean will just show what it would do!")
	peristentFlags.StringP(olderthenFlag, "o", "7d", "Set the duration string (e.g 5d, 1w etc.) how old objeccts must be to be deleted or listed. E.g. if set to 7d, AMIs will be delete which are older then 7 days. For security groups we only get the creation date of the past 90 days.")
	peristentFlags.StringP(debugFlag, "", "info", "Enable debugging. Possible Values [debug,info,warn,error,fatal]")

	internal.CheckError(viper.BindPFlags(peristentFlags), internal.Logger.Fatalf)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		internal.CheckError(err, internal.Logger.Fatalf)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".awsclean")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		internal.Logger.Error(err) // Just to show the error from ReadInConfig
	}

	err := internal.Logger.SetLogLevel(viper.GetString(debugFlag))
	internal.Logger.Error(err)
}
