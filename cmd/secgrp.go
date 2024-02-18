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
	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/xhit/go-str2duration/v2"
)

// Constants used in command flags
const (
	onlyUnusedFlag = "only-unused"
	createdAgoFlag = "created-ago"
	startTimeFlag  = "start-time"
	endTimeFlag    = "end-time"
)

const (
	secGrpCmdName       = "securitzGroups"
	secGrpListCmdName   = "list"
	secGrpDeleteCmdName = "delete"
)

var (
	secGrpCmdAliases       = []string{"secgrp"}
	secGrpListCmdAliases   = []string{"ls"}
	secGrpDeleteCmdAliases = []string{"del"}
)

var (
	secGrpDeleteCmdExamples = fmt.Sprintf(`
  %[1]s %[2]s %[4]s
  %[1]s %[3]s %[4]s
  %[1]s %[3]s %[5]s
`, binaryname,
		secGrpCmdName,
		secGrpCmdAliases[0],
		secGrpDeleteCmdName,
		secGrpDeleteCmdAliases[0])
	secGrpListCmdExamples = fmt.Sprintf(`
		%[1]s %[2]s %[4]s
		%[1]s %[3]s %[4]s
		%[1]s %[3]s %[5]s
	  `, binaryname,
		secGrpCmdName,
		secGrpCmdAliases[0],
		secGrpListCmdName,
		secGrpListCmdAliases[0])
)

// secGrpCmd represents the secgrp command
var secGrpCmd = &cobra.Command{
	Use:     secGrpCmdName,
	Aliases: secGrpCmdAliases,
	Short:   "Cleanup or list SecurityGroups",
	Long: fmt.Sprintf(`

Examples:
%s%s`,
		secGrpDeleteCmdExamples,
		secGrpListCmdExamples),
}

// secGrpListCmd represents the list command
var secGrpListCmd = &cobra.Command{
	Use:     secGrpListCmdName,
	Aliases: secGrpListCmdAliases,
	Short:   "Just lists SecurityGroups",
	Long: fmt.Sprintf(`Just list all SecurityGroups from connected AWS account.
	
Also the command tries to get the CreationTime from CloudTrail. CloudTrail only has this information for the past 90 days.
So older SecurityGroups will have no CreationTime / Creator information.
	
Examples:
%s`,
		secGrpListCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {

		secgrp, startDatetime, endDatetime := setup()

		err := secgrp.GetSecurityGroups(startDatetime, endDatetime)
		internal.CheckError(err, extendedslog.Logger.Fatalf)

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			secGrpPrintJSON(secgrp.GetAllSecurityGroups())
		default:
			secGrpPrintTable(secgrp.GetAllSecurityGroups())
		}

	},
}

var secGrpDeleteCmd = &cobra.Command{
	Use:     secGrpDeleteCmdName,
	Aliases: secGrpDeleteCmdAliases,
	Short:   "Delte older securityGrp from connected AWS account",
	Long: fmt.Sprintf(`Delte older securityGrp from connected AWS account

For security groups we only get the creation date of the past 90 days. So if older then date is specified less then 90d all SecurityGroups will be deleted which are older then this duration or do not have a CreationDate set as we couldn't get it from CloudTrail (in fact that means they are older then 90d).

Examples:
%s`,
		secGrpDeleteCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {

		secgrp, startDatetime, endDatetime := setup()

		err := secgrp.DeleteSecurityGroups(startDatetime, endDatetime)
		internal.CheckError(err, extendedslog.Logger.Fatalf)
	},
}

func secGrpBindFlags() {
	rootCmd.AddCommand(secGrpCmd)

	ninetyDayOffset, err := str2duration.ParseDuration("90d")
	internal.CheckError(err, extendedslog.Logger.Fatalf)

	const objType = "SecurityGroups"

	secGrpListCmdFlags := secGrpListCmd.Flags()
	secGrpListCmdFlags.BoolP(onlyUnusedFlag, "u", false, "defines if only-unused SecurityGroups are listed or all [Default: false]")
	listOnlyFlags(secGrpListCmdFlags, objType)

	secGrpDeleteCmdFlags := secGrpDeleteCmd.Flags()
	deleteOnlyFlags(secGrpDeleteCmdFlags, objType)

	secGrpCmdPersistentFlags := secGrpCmd.PersistentFlags()
	ninetyDaysAgo := time.Now().Add(ninetyDayOffset * -1)
	secGrpCmdPersistentFlags.StringP(startTimeFlag, "s", ninetyDaysAgo.Format(time.RFC3339), fmt.Sprintf("Set start datetime using format: %s [default: %s]", time.RFC3339, ninetyDaysAgo.Format(time.RFC3339)))
	secGrpCmdPersistentFlags.StringP(endTimeFlag, "e", time.Now().Format(time.RFC3339), fmt.Sprintf("Set end datetime using format: %s [default: %s]", time.RFC3339, time.Now().Format(time.RFC3339)))

	secGrpCmd.AddCommand(secGrpListCmd)
	secGrpCmd.AddCommand(secGrpDeleteCmd)

	internal.CheckError(viper.BindPFlags(secGrpCmdPersistentFlags), extendedslog.Logger.Fatalf)
	internal.CheckError(viper.BindPFlags(secGrpDeleteCmdFlags), extendedslog.Logger.Fatalf)
	internal.CheckError(viper.BindPFlags(secGrpListCmdFlags), extendedslog.Logger.Fatalf)
}

func setup() (*secgrp.SecGrp, time.Time, time.Time) {
	olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

	awsClient := internal.NewAWSClient()
	secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

	startDatetime, err := time.Parse(time.RFC3339, viper.GetString(startTimeFlag))
	internal.CheckError(err, extendedslog.Logger.Fatalf)

	endDatetime, err := time.Parse(time.RFC3339, viper.GetString(endTimeFlag))
	internal.CheckError(err, extendedslog.Logger.Fatalf)
	return secgrp, startDatetime, endDatetime
}

func secGrpPrintTable(grps internal.SecurityGroups) {
	grpsTable := table.New("ID", "Name", "Creation Datetime", "IsUsed")
	for _, grp := range grps {
		// TODO: conditionally add tags here.
		grpsTable.AddRow(grp.SecurityGroup.GroupId, grp.SecurityGroup.GroupName, grp.CreationTime.Format(time.RFC3339), grp.IsUsed)
	}
	grpsTable.Print()
}

func secGrpPrintJSON(grps internal.SecurityGroups) {
	out, err := json.Marshal(grps)
	internal.CheckError(err, extendedslog.Logger.Fatalf)
	fmt.Print(string(out))
}
