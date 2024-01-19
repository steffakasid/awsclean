/*
Copyright © 2023 steffakasid
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
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
	Run: func(cmd *cobra.Command, args []string) {
		var nextToken string = "empty"

		// olderthenDuration, err := str2duration.ParseDuration(viper.GetString(olderthen))
		// internal.CheckError(err, logger.Fatalf)

		cfg, _ := config.LoadDefaultConfig(context.TODO())
		cloudtrailclient := cloudtrail.NewFromConfig(cfg)

		for nextToken != "" {
			lookup := &cloudtrail.LookupEventsInput{
				LookupAttributes: []types.LookupAttribute{
					{
						AttributeKey:   types.LookupAttributeKeyEventName,
						AttributeValue: aws.String("CreateSecurityGroup"),
					},
				},
			}
			// We only get CloudTrailEvents of the last 90d: https://docs.aws.amazon.com/sdk-for-go/api/service/cloudtrail/#CloudTrail.LookupEvents
			// ResouceName: vpc-a51078cd
			// ResouceName: eksctl-eks-dev-nodegroup-apic-gw-1a-green-SG-16ACVO6XMU6HE
			// ResouceName: sg-018ce2cbe787b04ef
			// Time 2024-01-12 14:37:43 +0000 UTC
			// Wer ist schuld? `email@adress.com`
			// ---------------------------------------------
			out, err := cloudtrailclient.LookupEvents(context.TODO(), lookup)
			if nextToken != "empty" {
				lookup.NextToken = aws.String(nextToken)
			}
			nextToken = *out.NextToken
			cobra.CheckErr(err)

			for _, ev := range out.Events {
				for _, res := range ev.Resources {
					fmt.Println("ResouceName:", *res.ResourceName)
				}
				fmt.Println("Time", ev.EventTime)
				fmt.Println("Wer ist schuld?", *ev.Username)
				fmt.Println("---------------------------------------------")
				fmt.Println()
			}
		}

		// awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig)

		// secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, viper.GetBool(dryrun), viper.GetBool(showtags))
		// secgrp.DeleteUnusedSecurityGroups()
	},
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
		internal.CheckError(err, logger.Fatalf)

		awsClient := internal.NewAWSClient(config.LoadDefaultConfig, ec2.NewFromConfig)
		secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, viper.GetBool(dryrun), viper.GetBool(showtags))

		secGrps, err := secgrp.GetSecurityGroups(viper.GetBool(onlyUnused))
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
		log.Fatal("NOT IMPLEMENTED")
	},
}

func init() {
	rootCmd.AddCommand(secgrpCmd)

	// Implement flags here
	secGrpListFlags := secGrpListCmd.Flags()
	secGrpListFlags.BoolP(onlyUnused, "u", false, "defines if only-unused SecurityGroups are listed or all [Default: false]")

	// Add Child commands here
	secgrpCmd.AddCommand(secGrpListCmd)
	secgrpCmd.AddCommand(secGrpDeleteCmd)
}
