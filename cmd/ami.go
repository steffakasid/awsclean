/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/amiclean"
	"github.com/xhit/go-str2duration/v2"
)

const (
	accountFlag   = "account"
	launchTplFlag = "launch-templates"
	ignoreFlag    = "ignore"
)

// amiCmd represents the ami command
var amiCmd = &cobra.Command{
	Use:   "ami",
	Short: "This tool can be used to cleanup old and unused AWS amis",
	Long: `This tool can be used to cleanup old and unused AWS amis. You can specify the 
owner (AWS account) of AMIs and a duration how much older an AMI must be before
it gets deleted. The default duration is set to 7 days.

Examples:
  awsclean ami                      scan all AMIs owned by self and delete them if they are unused and older then 7 days.                   
  awsclean ami --account 2451251    scan all AMIs of self and were AWS account 2451251 are owner
  awsclean ami --dry-run            do not delete anything just show what you would do
  awsclean ami --older-then 5w	    delete all images which are older then 5w and are unused`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthenFlag))
		internal.CheckError(err, internal.Logger.Fatalf)
		awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig, cloudtrail.NewFromConfig)

		amiclean := amiclean.NewInstance(awsClient, olderthenDuration, viper.GetString(accountFlag), viper.GetBool(dryrunFlag), viper.GetBool(launchTplFlag), viper.GetStringSlice(ignoreFlag))
		amiclean.GetUsedAMIs()
		err = amiclean.DeleteOlderUnusedAMIs()
		internal.CheckError(err, internal.Logger.Fatalf)
	},
}

func init() {
	rootCmd.AddCommand(amiCmd)

	amiFlags := amiCmd.Flags()

	amiFlags.StringArrayP(ignoreFlag, "i", []string{}, "Set ignore regex patterns. If a ami name matches the pattern it will be exclueded from cleanup.")
	amiFlags.BoolP(launchTplFlag, "l", false, "Additionally scan launch templates for used AMIs.")
	amiFlags.StringP(accountFlag, "a", "", "Set AWS account number to cleanup AMIs. Used to set owner information when selecting AMIs. If not set only 'self' is used.")

	internal.CheckError(viper.BindPFlags(amiFlags), internal.Logger.Fatalf)
}
