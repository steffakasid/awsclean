/*
Copyright © 2023 steffakasid
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/secgrp"
	"github.com/xhit/go-str2duration/v2"
)

// Constants used in command flags
const (
	onlyUnused = "only-unused"
	createdAgo = "created-ago"
)

// secgrpCmd represents the secgrp command
var secgrpCmd = &cobra.Command{
	Use:   "secgrp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// secGrpListCmd represents the list command
var secGrpListCmd = &cobra.Command{
	Use:   "secgrp list [options]",
	Short: "Just lists securityGrps",
	Long: `Jus list all securityGrp from connected AWS account.
	
Also the command tries to get the CreationTime from CloudTrail. CloudTrail only has this information for the past 90 days.
So older SecurityGroups will have no CreationTime.
	
Examples:
  awsclean secgrp list --older-then 5w  list all SecurityGroup which are older then 5w and are not used
  awsclean secgrp list --dry-run        --dry-run has not effect here it will just list the security groups
  awsclean secgrp list --show-tags      print out the SecurityGroups with their tags`,
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthen))
		cobra.CheckErr(err)

		createdAgoDuration, err := str2duration.ParseDuration(viper.GetString(createdAgo))
		cobra.CheckErr(err)

		awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig, cloudtrail.NewFromConfig)
		secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, &createdAgoDuration, viper.GetBool(dryrun), viper.GetBool(onlyUnused), viper.GetBool(showtags))

		secGrps, err := secgrp.GetSecurityGroups()
		cobra.CheckErr(err)

		fmt.Println("ID\t\tName\t\tCreationDate")
		for _, secGrp := range secGrps {
			fmt.Printf("%s\t\t%s\t\t%s", secGrp.ID, secGrp.Name, secGrp.CreationTime)
		}
	},
}

var secGrpDeleteCmd = &cobra.Command{
	Use:   "secgrp delte [options]",
	Short: "Delte older securityGrp from connected AWS account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Nothing to see here until now")
		olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthen))
		internal.CheckError(err, logger.Fatalf)

		createdAgoDuration, err := str2duration.ParseDuration(viper.GetString(createdAgo))
		cobra.CheckErr(err)

		awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig, cloudtrail.NewFromConfig)

		secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, &createdAgoDuration, viper.GetBool(dryrun), viper.GetBool(onlyUnused), viper.GetBool(showtags))
		secgrp.DeleteUnusedSecurityGroups()
	},
}

func init() {
	rootCmd.AddCommand(secgrpCmd)

	// Implement flags here
	secGrpListFlags := secGrpListCmd.Flags()
	secGrpListFlags.BoolP(onlyUnused, "u", false, "defines if only-unused SecurityGroups are listed or all [Default: false]")
	secGrpListFlags.StringP(createdAgo, "c", "", "only list security groups which were created x-days ago. We can only reach back 90 days (e.g. 1m)")

	// Add Child commands here
	secgrpCmd.AddCommand(secGrpListCmd)
	secgrpCmd.AddCommand(secGrpDeleteCmd)
}
