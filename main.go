package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	logger "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/steffakasid/amiclean/internal"
	"github.com/xhit/go-str2duration/v2"
)

const (
	dryrun    = "dry-run"
	olderthen = "older-then"
	account   = "account"
	launchTpl = "launch-templates"
	ignore    = "ignore"
)

var version = "0.1-dev"

func init() {
	flag.BoolP(dryrun, "d", false, "If set to true nothing will be deleted. And amiclean will just show what it would do!")
	flag.StringP(olderthen, "o", "7d", "Set the duration string (e.g 5d, 1w etc.) how old AMIs must be to be deleted. E.g. if set to 7d, AMIs will be delete which are older then 7 days.")
	flag.StringP(account, "a", "", "Set AWS account number to cleanup AMIs. Used to set owner information when selecting AMIs. If not set only 'self' is used.")
	flag.StringArrayP(ignore, "i", []string{}, "Set ignore regex patterns. If a ami name matches the pattern it will be exclueded from cleanup.")
	flag.BoolP(launchTpl, "l", false, "Additionally scan launch templates for used AMIs.")
	flag.BoolP("version", "v", false, "Print version information")
	flag.BoolP("help", "?", false, "Print usage information")

	flag.Usage = func() {
		w := os.Stderr

		fmt.Fprintf(w, "Usage of %s: \n", os.Args[0])
		fmt.Fprintln(w, `
This tool can be used to cleanup old and unused AWS amis. You can specify the 
owner (AWS account) of AMIs and a duration how much older an AMI must be before
it gets deleted. The default duration is set to 7 days. 

Usage:
  amiclean [flags]

Preqrequisites:
  amiclean uses already provided credentials in ~/.aws/credentials also it uses the
  central configuration in ~/.aws/config!


Examples:
  amiclean                      scan all AMIs owned by self and delete them if they are unused and older then 7 days.                   
  amiclean --account 2451251    scan all AMIs of self and were AWS account 2451251 are owner
  amiclean --dry-run            do not delete anything just show what you would do
  amiclean --older-then 5w		delete all images which are older then 5w and are unused

Flags:`)

		flag.PrintDefaults()
	}

	flag.Parse()
	err := viper.BindPFlags(flag.CommandLine)
	internal.CheckError(err, logger.Fatalf)
}

func main() {
	if viper.GetBool("version") {
		fmt.Printf("AMICLEAN version: %s\n", version)
	} else if viper.GetBool("help") {
		flag.Usage()
	} else {
		olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthen))
		internal.CheckError(err, logger.Fatalf)
		amiclean := internal.NewInstance(config.LoadDefaultConfig, ec2.NewFromConfig, olderthenDuration, viper.GetString(account), viper.GetBool(dryrun), viper.GetBool(launchTpl), viper.GetStringSlice(ignore))
		amiclean.GetUsedAMIs()
		err = amiclean.DeleteOlderUnusedAMIs()
		internal.CheckError(err, logger.Fatalf)
	}
}
