/*
Copyright © 2023 steffakasid
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/secgrp"
	"github.com/xhit/go-str2duration/v2"
)

// Constants used in command flags
const (
	onlyUnusedFlag = "only-unused"
	createdAgoFlag = "created-ago"
	startTimeFlag  = "start-time"
	endTimeFlag    = "end-time"
)

// secgrpCmd represents the secgrp command
var secgrpCmd = &cobra.Command{
	Use:   "secgrp",
	Short: "Cleanup or list SecurityGroups",
	Long: `
	`,
}

// secGrpListCmd represents the list command
var secGrpListCmd = &cobra.Command{
	Use:   "list [options]",
	Short: "Just lists SecurityGroups",
	Long: `Just list all SecurityGroups from connected AWS account.
	
Also the command tries to get the CreationTime from CloudTrail. CloudTrail only has this information for the past 90 days.
So older SecurityGroups will have no CreationTime / Creator information.
	
Examples:
  awsclean secgrp list --older-then 5w  list all SecurityGroup which are older then 5w and are not used
  awsclean secgrp list --dry-run        --dry-run has not effect here it will just list the security groups
  awsclean secgrp list --show-tags      print out the SecurityGroups with their tags`,
	Run: func(cmd *cobra.Command, args []string) {

		secgrp, startDatetime, endDatetime := setup()

		err := secgrp.GetSecurityGroups(startDatetime, endDatetime)
		internal.CheckError(err, internal.Logger.Fatalf)

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			secGrpPrintJSON(secgrp.GetAllSecurityGroups())
		default:
			secGrpPrintTable(secgrp.GetAllSecurityGroups())
		}

	},
}

var secGrpDeleteCmd = &cobra.Command{
	Use:   "delete [options]",
	Short: "Delte older securityGrp from connected AWS account",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		secgrp, startDatetime, endDatetime := setup()

		err := secgrp.DeleteSecurityGroups(startDatetime, endDatetime)
		internal.CheckError(err, internal.Logger.Fatalf)
	},
}

func secGrpBindFlags() {
	rootCmd.AddCommand(secgrpCmd)

	ninetyDayOffset, err := str2duration.ParseDuration("90d")
	internal.CheckError(err, internal.Logger.Fatalf)

	// Implement flags here
	secGrpListFlags := secGrpListCmd.Flags()
	secGrpListFlags.BoolP(onlyUnusedFlag, "u", false, "defines if only-unused SecurityGroups are listed or all [Default: false]")
	secGrpListFlags.StringP(createdAgoFlag, "c", "", "only list security groups which were created x-days ago. We can only reach back 90 days (e.g. 1m)")
	internal.CheckError(viper.BindPFlags(secGrpListFlags), internal.Logger.Fatalf)

	secGrpPersistentFlags := secgrpCmd.PersistentFlags()
	ninetyDaysAgo := time.Now().Add(ninetyDayOffset * -1)
	secGrpPersistentFlags.StringP(startTimeFlag, "s", ninetyDaysAgo.Format(time.RFC3339), fmt.Sprintf("Set start datetime using format: %s [default: %s]", time.RFC3339, ninetyDaysAgo.Format(time.RFC3339)))
	secGrpPersistentFlags.StringP(endTimeFlag, "e", time.Now().Format(time.RFC3339), fmt.Sprintf("Set end datetime using format: %s [default: %s]", time.RFC3339, time.Now().Format(time.RFC3339)))
	internal.CheckError(viper.BindPFlags(secGrpPersistentFlags), internal.Logger.Fatalf)

	// Add Child commands here
	secgrpCmd.AddCommand(secGrpListCmd)
	secgrpCmd.AddCommand(secGrpDeleteCmd)
}

func setup() (*secgrp.SecGrp, time.Time, time.Time) {
	olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

	createdAgoDuration := internal.ParseDuration(viper.GetString(createdAgoFlag))

	awsClient := internal.NewAWSClient()
	secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, &createdAgoDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag), viper.GetBool(showtagsFlag))

	startDatetime, err := time.Parse(time.RFC3339, viper.GetString(startTimeFlag))
	internal.CheckError(err, internal.Logger.Fatalf)

	endDatetime, err := time.Parse(time.RFC3339, viper.GetString(endTimeFlag))
	internal.CheckError(err, internal.Logger.Fatalf)
	return secgrp, startDatetime, endDatetime
}

func secGrpPrintTable(grps internal.SecurityGroups) {
	grpsTable := table.New("ID", "Name", "Creation Datetime", "IsUsed")
	for _, grp := range grps {
		grpsTable.AddRow(grp.ID, grp.Name, grp.CreationTime.Format(time.RFC3339), grp.IsUsed)
	}
	grpsTable.Print()
}

func secGrpPrintJSON(grps internal.SecurityGroups) {
	out, err := json.Marshal(grps)
	internal.CheckError(err, internal.Logger.Fatalf)
	fmt.Print(string(out))
}
