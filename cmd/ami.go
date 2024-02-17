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

const (
	amiCmdName       = "ami"
	amiListCmdName   = "list"
	amiDeleteCmdName = "delete"
)

var (
	amiListCmdAliases   = []string{"ls"}
	amiDeleteCmdAliases = []string{"del"}
)

var (
	amiDeleteCmdExamples = fmt.Sprintf(`
  %[1]s %[2]s %[3]s --account 2451251 scan all AMIs of self and were AWS account 2451251 are owner  
  %[1]s %[2]s %[3]s --dry-run         do not delete anything just show what you would do
  %[1]s %[2]s %[3]s --older-then      5w delete all images which are older then 5w and are unused
  %[1]s %[2]s %[3]s --help            show help for this sub-command
	`,
		binaryname,
		amiCmdName,
		amiDeleteCmdName)
	amiListCmdExamples = fmt.Sprintf(`
  %[1]s %[2]s %[3]s --account 2451251 scan all AMIs of self and were AWS account 2451251 are owner  
  %[1]s %[2]s %[3]s --dry-run         do not delete anything just show what you would do 
  %[1]s %[2]s %[3]s --help            show help for this sub-command
	`,
		binaryname,
		amiCmdName,
		amiListCmdName)
)

// amiCmd represents the ami command
var amiCmd = &cobra.Command{
	Use:   amiCmdName,
	Short: "This tool can be used to delete or list old and unused AWS Amis",
	Long: fmt.Sprintf(`This tool can be used to list or delete old and unused AWS amis. What you wanna to can
be defined via sub-commands.

Examples:
%s%s`,
		amiDeleteCmdExamples,
		amiListCmdExamples),
}

var amiListCmd = &cobra.Command{
	Use:     amiListCmdName,
	Aliases: amiListCmdAliases,
	Short:   "This tool can be used to list old and unused AWS amis",
	Long: fmt.Sprintf(`Use this command to list AWS amis. Nothing will be delete so it can safely be used to
view which amis exist.
	
Examples:
%s`,
		amiListCmdExamples),
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
	Use:     amiDeleteCmdName,
	Aliases: amiDeleteCmdAliases,
	Short:   "This tool can be used to cleanup old and unused AWS amis",
	Long: fmt.Sprintf(`This tool can be used to delte old and unused AWS amis. You can specify the 
owner (AWS account) of AMIs and a duration how much older an AMI must be, beforeit gets deleted. 
The default duration is set to 7 days. Also there is a try-run flag wich can be used to simulate the delete.
Nothing will be delete in that case.

Examples:
%s`,
		amiDeleteCmdExamples),
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

	amiDeleteFlags := amiDeleteCmd.Flags()
	amiDeleteFlags.StringArrayP(ignoreFlag, "i", []string{}, "Set ignore regex patterns. If a ami name matches the pattern it will be exclueded from cleanup.")
	deleteExtraFlags(amiDeleteFlags, "AMIs")

	amiCmdPersistentFlags := amiCmd.PersistentFlags()
	amiCmdPersistentFlags.BoolP(launchTplFlag, "l", false, "Additionally scan launch templates for used AMIs.")
	amiCmdPersistentFlags.StringP(accountFlag, "a", "", "Set AWS account number to cleanup AMIs. Used to set owner information when selecting AMIs. If not set only 'self' is used.")

	internal.CheckError(viper.BindPFlags(amiCmdPersistentFlags), internal.Logger.Fatalf)
	internal.CheckError(viper.BindPFlags(amiDeleteFlags), internal.Logger.Fatalf)
}

func amiPrintTable(amis []ec2Types.Image) {
	grpsTable := table.New("ID", "Name", "Creation DateTime")
	for _, ami := range amis {
		// TODO: Conditionally add ami.tags here.
		grpsTable.AddRow(*ami.ImageId, *ami.Name, *ami.CreationDate)
	}
	grpsTable.Print()
}

func amiPrintJSON(amis []ec2Types.Image) {
	out, err := json.Marshal(amis)
	internal.CheckError(err, internal.Logger.Fatalf)
	fmt.Print(string(out))
}
