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
	eslog "github.com/steffakasid/eslog"
)

const (
	secGrpCmdName       = "securityGroups"
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

		secgrp, startDatetime, endDatetime := setupSecGrpClient()

		err := secgrp.GetSecurityGroups(startDatetime, endDatetime)
		eslog.LogIfErrorf(err, eslog.Fatalf, "secgrp.GetSecurityGroups() failed: %s", err)

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

		secgrp, startDatetime, endDatetime := setupSecGrpClient()

		err := secgrp.DeleteSecurityGroups(startDatetime, endDatetime)
		eslog.LogIfErrorf(err, eslog.Fatalf, "secgrp.DeleteSecurityGroups() failed: %s", err)
	},
}

func secGrpBindFlags() {
	rootCmd.AddCommand(secGrpCmd)

	const objType = "SecurityGroups"

	secGrpListCmdFlags := secGrpListCmd.Flags()
	secGrpListCmdFlags.BoolP(onlyUnusedFlag, onlyUnusedFlagSH, false, "defines if only-unused SecurityGroups are listed or all [Default: false]")
	listOnlyFlags(secGrpListCmdFlags, objType)

	secGrpDeleteCmdFlags := secGrpDeleteCmd.Flags()
	deleteOnlyFlags(secGrpDeleteCmdFlags)

	secGrpCmdPersistentFlags := secGrpCmd.PersistentFlags()

	secGrpCmd.AddCommand(secGrpListCmd)
	secGrpCmd.AddCommand(secGrpDeleteCmd)

	err := viper.BindPFlags(secGrpCmdPersistentFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)

	err = viper.BindPFlags(secGrpDeleteCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)

	err = viper.BindPFlags(secGrpListCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)
}

func setupSecGrpClient() (*secgrp.SecGrp, time.Time, time.Time) {
	olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

	awsClient := internal.NewAWSClient()
	secgrp := secgrp.NewInstance(awsClient, &olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

	startDatetime, err := time.Parse(time.RFC3339, viper.GetString(startTimeFlag))
	eslog.LogIfErrorf(err, eslog.Fatalf, "Error parsing given %s: %s", startTimeFlag, err)

	endDatetime, err := time.Parse(time.RFC3339, viper.GetString(endTimeFlag))
	eslog.LogIfErrorf(err, eslog.Fatalf, "Error parsing given %s: %s", endTimeFlag, err)
	return secgrp, startDatetime, endDatetime
}

func secGrpPrintTable(grps internal.SecurityGroups) {
	grpsTable := table.New("ID", "Name", "Creation Datetime", "Created by", "IsUsed")
	for _, grp := range grps {
		// TODO: conditionally add tags here.
		if grp.SecurityGroup != nil {
			if grp.CreationTime == nil {
				grp.CreationTime = &time.Time{}
			}
			grpsTable.AddRow(nilCheck(grp.SecurityGroup.GroupId), nilCheck(grp.SecurityGroup.GroupName), grp.CreationTime.Format(time.RFC3339), grp.Creator, grp.IsUsed)
		}
	}
	grpsTable.Print()
}

func secGrpPrintJSON(grps internal.SecurityGroups) {
	out, err := json.Marshal(grps)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Json.Marshal(grps) failed: %s", err)
	fmt.Print(string(out))
}
