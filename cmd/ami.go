/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/amiclean"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
	// TODO: needs rework
	Long: `This tool can be used to cleanup old and unused AWS amis. You can specify the 
owner (AWS account) of AMIs and a duration how much older an AMI must be before
it gets deleted. The default duration is set to 7 days.

Examples:
  awsclean ami                      scan all AMIs owned by self and delete them if they are unused and older then 7 days.                   
  awsclean ami --account 2451251    scan all AMIs of self and were AWS account 2451251 are owner
  awsclean ami --dry-run            do not delete anything just show what you would do
  awsclean ami --older-then 5w	    delete all images which are older then 5w and are unused`,
}

var amiListCmd = &cobra.Command{
	Use:   "list",
	Short: "This tool can be used to cleanup old and unused AWS amis",
	//TODO: define
	Long: `
	`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()

		amiclean := amiclean.NewInstance(awsClient, olderthenDuration, viper.GetString(accountFlag), viper.GetBool(dryrunFlag), viper.GetBool(launchTplFlag), viper.GetStringSlice(ignoreFlag))

		err := amiclean.GetAMIs()
		internal.CheckError(err, internal.Logger.Fatalf)

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			amiPrintJSON(amiclean.GetAllAMIs())
		default:
			amiPrintTable(amiclean.GetAllAMIs())
		}
	},
}

var amiDeleteCmd = &cobra.Command{
	Use:   "delete",
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
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()

		amiclean := amiclean.NewInstance(awsClient, olderthenDuration, viper.GetString(accountFlag), viper.GetBool(dryrunFlag), viper.GetBool(launchTplFlag), viper.GetStringSlice(ignoreFlag))

		err := amiclean.DeleteOlderUnusedAMIs()
		internal.CheckError(err, internal.Logger.Fatalf)
	},
}

func amiBindFlags() {
	amiCmd.AddCommand(amiDeleteCmd)
	amiCmd.AddCommand(amiListCmd)
	rootCmd.AddCommand(amiCmd)

	amiFlags := amiCmd.Flags()

	amiFlags.StringArrayP(ignoreFlag, "i", []string{}, "Set ignore regex patterns. If a ami name matches the pattern it will be exclueded from cleanup.")
	amiFlags.BoolP(launchTplFlag, "l", false, "Additionally scan launch templates for used AMIs.")
	amiFlags.StringP(accountFlag, "a", "", "Set AWS account number to cleanup AMIs. Used to set owner information when selecting AMIs. If not set only 'self' is used.")

	internal.CheckError(viper.BindPFlags(amiFlags), internal.Logger.Fatalf)
}

func amiPrintTable(amis []ec2Types.Image) {
	grpsTable := table.New("ID", "Name", "Creation DateTime")
	for _, ami := range amis {
		grpsTable.AddRow(*ami.ImageId, *ami.Name, *ami.CreationDate)
	}
	grpsTable.Print()
}

func amiPrintJSON(amis []ec2Types.Image) {
	out, err := json.Marshal(amis)
	internal.CheckError(err, internal.Logger.Fatalf)
	fmt.Print(string(out))
}
