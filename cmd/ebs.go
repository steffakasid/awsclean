/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/amiclean/internal"
	"github.com/steffakasid/amiclean/internal/ebsclean"
	"github.com/xhit/go-str2duration/v2"
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
  awsclean ebs --dry-run        do not delete any EBS volume just show what you would do`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthen))
		internal.CheckError(err, logger.Fatalf)

		awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig)
		ebsclean := ebsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrun))

		ebsclean.DeleteUnusedEBSVolumes()
	},
}

func init() {
	rootCmd.AddCommand(ebsCmd)
}
