/*
Copyright © 2022 steffakasid
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	cwlogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/loggroup"
	eslog "github.com/steffakasid/eslog"
)

const (
	logGrpsCmdName       = "CloudWatchLogGroups"
	logGrpsDeleteCmdName = "delete"
	logGrpsListCmdName   = "list"
)

var (
	logGrpsCmdAliases       = []string{"loggrps"}
	logGrpsListCmdAliases   = []string{"ls"}
	logGrpsDeleteCmdAliases = []string{"del"}
)

var (
	logGrpsDeleteCmdExamples = fmt.Sprintf(`
  %[1]s %[2]s %[3]s --older-then 5w  delete all logGrps volumes which are older then 5w and are not bound
  %[1]s %[2]s %[3]s --dry-run        do not delete any logGrps volume just show what should be done
  %[1]s %[2]s %[4]s --older-then 5w     delete all logGrps volumes which are older then 5w and are not bound
  %[1]s %[2]s %[4]s --dry-run           do not delete any logGrps volume just show what should be done
  `,
		binaryname,
		logGrpsCmdName,
		logGrpsDeleteCmdName,
		logGrpsDeleteCmdAliases[0])
	logGrpsListCmdExamples = fmt.Sprintf(`
	%[1]s %[2]s %[3]s --show-tags      print out tags of logGrps volumes
	%[1]s %[2]s %[4]s --show-tags        print out tags of logGrps volumes
	`,
		binaryname,
		logGrpsCmdName,
		logGrpsListCmdName,
		logGrpsListCmdAliases[0])
)

// logGrpsCmd represents the logGrps command
var logGrpsCmd = &cobra.Command{
	Use:     logGrpsCmdName,
	Aliases: logGrpsCmdAliases,
	Short:   "Cleanup unused logGrps volumes",
	Long: fmt.Sprintf(`This tool can be used to list or cleanup old and unbound Elastic Block Store (logGrps) volumes.

Examples:
%s%s
`,
		logGrpsDeleteCmdExamples,
		logGrpsListCmdExamples),
}

var logGrpsDeleteCmd = &cobra.Command{
	Use:     logGrpsDeleteCmdName,
	Aliases: logGrpsDeleteCmdAliases,
	Short:   "Cleanup unused logGrps volumes",
	Long: fmt.Sprintf(`This tool can be used to cleanup old and unbound Elastic Block Store (logGrps) volumes.

Examples:
%s`,
		logGrpsDeleteCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {
		lgGrpsService, _, _ := setupLogGroups()

		lgGrpsService.DeleteUnused()
	},
}

var logGrpsListCmd = &cobra.Command{
	Use:     logGrpsListCmdName,
	Aliases: logGrpsListCmdAliases,
	Short:   "List unused logGrps volumes",
	Long: fmt.Sprintf(`This command can be used to list unused Elastic Block Store Volumes.

Examples:
%s`,
		logGrpsListCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {

		lgGrpsService, _, _ := setupLogGroups()

		lgGrpsService.GetCloudWatchLogGroups()

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			logGrpsPrintJSON(lgGrpsService.GetUnusedLogGroups())
		default:
			logGrpsPrintTable(lgGrpsService.GetUnusedLogGroups())
		}
	},
}

func logGrpsCmdInit() {
	logGrpsCmd.AddCommand(logGrpsDeleteCmd)
	logGrpsCmd.AddCommand(logGrpsListCmd)
	rootCmd.AddCommand(logGrpsCmd)
	logGrpsDeleteCmd.PreRun = func(cmd *cobra.Command, args []string) {
		logGrpsDeleteCmdBindFlags()
	}
	logGrpsListCmd.PreRun = func(cmd *cobra.Command, args []string) {
		logGrpsListCmdBindFlags()
	}
	logGrpsDeleteCmdFlags := logGrpsDeleteCmd.Flags()
	deleteOnlyFlags(logGrpsDeleteCmdFlags)
	logGrpsListCmdFlags := logGrpsListCmd.Flags()
	listOnlyFlags(logGrpsListCmdFlags, "CloudWatchLogGrps")
}

func logGrpsDeleteCmdBindFlags() {
	logGrpsDeleteCmdFlags := logGrpsDeleteCmd.Flags()
	err := viper.BindPFlags(logGrpsDeleteCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)
}

func logGrpsListCmdBindFlags() {
	logGrpsListCmdFlags := logGrpsListCmd.Flags()
	err := viper.BindPFlags(logGrpsListCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)
}

func setupLogGroups() (loggrp *loggroup.LogGrp, startDatetime time.Time, endDatetime time.Time) {
	olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

	awsClient := internal.NewAWSClient()
	loggrp = loggroup.NewInstance(awsClient, &olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

	startDatetime, err := time.Parse(time.RFC3339, viper.GetString(startTimeFlag))
	eslog.LogIfErrorf(err, eslog.Fatalf, "Error parsing given %s: %s", startTimeFlag, err)

	endDatetime, err = time.Parse(time.RFC3339, viper.GetString(endTimeFlag))
	eslog.LogIfErrorf(err, eslog.Fatalf, "Error parsing given %s: %s", endTimeFlag, err)
	return loggrp, startDatetime, endDatetime
}

func logGrpsPrintTable(grps []cwlogsTypes.LogGroup) {
	grpsTable := table.New("LogGroup Name", "Creation Datetime", "Retention(d)")
	for _, grp := range grps {
		// TODO: Conditionally add Tags here.
		creationtime := time.UnixMilli(*grp.CreationTime)
		grpsTable.AddRow(*grp.LogGroupName, creationtime, grp.RetentionInDays)
	}
	grpsTable.Print()
}

func logGrpsPrintJSON(grps []cwlogsTypes.LogGroup) {
	out, err := json.Marshal(grps)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Json.Marshal(vols) failed: %s", err)
	fmt.Print(string(out))
}
