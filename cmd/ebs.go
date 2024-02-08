/*
Copyright © 2022 steffakasid
*/
package cmd

import (
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

func ebsBindFlags() {
	rootCmd.AddCommand(ebsCmd)

	ebsFlags := ebsCmd.Flags()

	ebsFlags.BoolP(showtagsFlag, "s", false, "show tags of ebs volumes")

	internal.CheckError(viper.BindPFlags(ebsFlags), internal.Logger.Fatalf)
}
