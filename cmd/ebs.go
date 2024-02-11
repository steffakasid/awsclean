/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/ebsclean"
)

const (
	showtagsFlag = "show-tags"
)

// ebsCmd represents the ebs command
var ebsCmd = &cobra.Command{
	Use:   "ebs",
	Short: "Cleanup unused EBS volumes",
	Long: `This tool can be used to cleanup old and unbound Elastic Block Store (EBS) volumes.
You can specify a duration on how old a EBS volums should be to be deleted. The default duration
is set to 7 days.
TODO: needs rework
Examples:
  awsclean ebs --older-then 5w  delete all EBS volumes which are older then 5w and are not bound
  awsclean ebs --dry-run        do not delete any EBS volume just show what should be done
  awsclean ebs --show-tags      print out tags of EBS volumes`,
}

var ebsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Cleanup unused EBS volumes",
	Long: `This tool can be used to cleanup old and unbound Elastic Block Store (EBS) volumes.
You can specify a duration on how old a EBS volums should be to be deleted. The default duration
is set to 7 days.

Examples:
  awsclean ebs --older-then 5w  delete all EBS volumes which are older then 5w and are not bound
  awsclean ebs --dry-run        do not delete any EBS volume just show what should be done
  awsclean ebs --show-tags      print out tags of EBS volumes`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()
		ebsclean := ebsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(showtagsFlag))

		ebsclean.DeleteUnusedEBSVolumes()
	},
}

var ebsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List unused EBS volumes",
	Long: `
	`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()
		ebsclean := ebsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(showtagsFlag))

		ebsclean.GetEBSVolumes()

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			secEBSPrintJSON(ebsclean.GetAllVolumes())
		default:
			secEBSPrintTable(ebsclean.GetAllVolumes())
		}
	},
}

func ebsBindFlags() {
	ebsCmd.AddCommand(ebsDeleteCmd)
	ebsCmd.AddCommand(ebsListCmd)
	rootCmd.AddCommand(ebsCmd)

	ebsFlags := ebsCmd.PersistentFlags()

	ebsFlags.BoolP(showtagsFlag, "s", false, "show tags of ebs volumes")

	internal.CheckError(viper.BindPFlags(ebsFlags), internal.Logger.Fatalf)
}

func secEBSPrintTable(vols []ec2Types.Volume) {
	grpsTable := table.New("Volume ID", "Creation Datetime", "State")
	for _, vol := range vols {
		grpsTable.AddRow(vol.VolumeId, vol.CreateTime.Format(time.RFC3339), vol.State)
	}
	grpsTable.Print()
}

func secEBSPrintJSON(vols []ec2Types.Volume) {
	out, err := json.Marshal(vols)
	internal.CheckError(err, internal.Logger.Fatalf)
	fmt.Print(string(out))
}
