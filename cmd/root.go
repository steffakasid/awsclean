/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	extendedslog "github.com/steffakasid/extended-slog"
)

// Constants used in command flags
const (
	accountFlag    = "account"
	debugFlag      = "debug"
	dryrunFlag     = "dry-run"
	endTimeFlag    = "end-time"
	ignoreFlag     = "ignore"
	launchTplFlag  = "launch-templates"
	olderthenFlag  = "older-then"
	outputFlag     = "output"
	onlyUnusedFlag = "only-unused"
	startTimeFlag  = "start-time"
	showtagsFlag   = "show-tags"
)

// constants used for short hand flags (to avoid collitions)
const (
	accountFlagSH    = "a"
	dryrunFlagSH     = "d"
	endTimeFlagSH    = "e"
	ignoreFlagSH     = "i"
	launchTplFlagSH  = "l"
	noShortHand      = ""
	onlyUnusedFlagSH = "u"
	olderthenFlagSH  = "o"
	startTimeFlagSH  = "s"
	showtagsFlagSH   = "t"
)

const binaryname = "awsclean"

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: binaryname,
	Long: fmt.Sprintf(`This tool is intended to be used to cleanup certain AWS services.
	
Right now it supports the following:
  - Amazon Machine Images (AMIs)
  - Elastic Blockstore (EBS) Volumes
  - SecurityGroups

Preqrequisites:
  amiclean uses already provided credentials in ~/.aws/credentials also it uses the
  central configuration in ~/.aws/config!

Examples:
  %s ami --help  show help for ami subcommand%s%s`, binaryname, amiDeleteCmdExamples, amiListCmdExamples),
}

func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// The one and only init() for package cmd
func init() {

	extendedslog.InitLogger()

	cobra.OnInitialize(initConfig)

	bindPersistentFlags()
	amiBindFlags()
	ebsBindFlags()
	secGrpBindFlags()
}

func bindPersistentFlags() {
	peristentFlags := rootCmd.PersistentFlags()

	peristentFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amiclean.yaml)")

	peristentFlags.StringP(debugFlag, "", "info", "Enable debugging. Possible Values [debug,info,warn,error,fatal]")
	peristentFlags.StringP(outputFlag, "", "table", "Define how to output results [table, json] (default: table)")

	extendedslog.Logger.Fatalf("Failed to bind Flags: %w", viper.BindPFlags(peristentFlags))
}

func deleteOnlyFlags(flagset *pflag.FlagSet, objType string) {
	flagset.BoolP(dryrunFlag, dryrunFlagSH, false, "If set to true nothing will be deleted. And amiclean will just show what it would do!")
	flagset.StringP(olderthenFlag, olderthenFlagSH, "7d", fmt.Sprintf("Set the duration string (e.g 5d, 1w etc.) how old %[1]s must be to be deleted. E.g. if set to 7d, %[1]s will be delete which are older then 7 days.", objType))
}

func listOnlyFlags(flagset *pflag.FlagSet, objType string) {
	flagset.BoolP(showtagsFlag, showtagsFlagSH, false, fmt.Sprintf("show tags of %s", objType))

	ninetyDayOffset := internal.ParseDuration("90d")
	ninetyDaysAgo := time.Now().Add(ninetyDayOffset * -1)
	flagset.StringP(startTimeFlag, startTimeFlagSH, ninetyDaysAgo.Format(time.RFC3339), fmt.Sprintf("Set start datetime using format: %s [default: %s]", time.RFC3339, ninetyDaysAgo.Format(time.RFC3339)))
	flagset.StringP(endTimeFlag, endTimeFlagSH, time.Now().Format(time.RFC3339), fmt.Sprintf("Set end datetime using format: %s [default: %s]", time.RFC3339, time.Now().Format(time.RFC3339)))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		extendedslog.Logger.Fatalf("Can not get os.UserHomeDir(): %w", err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".awsclean")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		extendedslog.Logger.Error(err) // Just to show the error from ReadInConfig
	}

	err := extendedslog.Logger.SetLogLevel(viper.GetString(debugFlag))
	extendedslog.Logger.Error(err)
}

func nilCheck(tocheck *string) string {
	if tocheck == nil {
		return "nil"
	}
	return *tocheck
}
