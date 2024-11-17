/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/eslog"
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
	cobra.OnInitialize(initConfig)

	bindPersistentFlags()

	amiCmdInit()

	ebsCmdInit()

	logGrpsCmdInit()

	secGrpCmdInit()
}

func bindPersistentFlags() {
	peristentFlags := rootCmd.PersistentFlags()

	peristentFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.amiclean.yaml)")

	peristentFlags.StringP(debugFlag, "", "info", "Enable debugging. Possible Values [debug,info,warn,error,fatal]")
	peristentFlags.StringP(outputFlag, "", "table", "Define how to output results [table, json] (default: table)")
	peristentFlags.StringP(olderthenFlag, olderthenFlagSH, "7d", "Set the duration string (e.g 5d, 1w etc.) how old an object must be to be deleted. E.g. if set to 7d, objects will be delete which are older then 7 days.")

	err := viper.BindPFlags(peristentFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %s", err)
}

func deleteOnlyFlags(flagset *pflag.FlagSet) {
	flagset.BoolP(dryrunFlag, dryrunFlagSH, false, "If set to true nothing will be deleted. And amiclean will just show what it would do!")
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
		eslog.LogIfErrorf(err, eslog.Fatalf, "Can not get os.UserHomeDir(): %w", err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(path.Join(home, ".config", "awsclean"))
		viper.SetConfigType("yaml")
		viper.SetConfigName(".awsclean")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		eslog.Error(err) // Just to show the error from ReadInConfig
	}

	err := eslog.Logger.SetLogLevel(viper.GetString(debugFlag))
	eslog.LogIfError(err, eslog.Error, err)
}

func nilCheck(tocheck *string) string {
	if tocheck == nil {
		return "nil"
	}
	return *tocheck
}
