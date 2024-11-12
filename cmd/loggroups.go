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
		// olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		// awsClient := internal.NewAWSClient()
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
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()
		lgGrps := awsClient.ListLogGrps(olderthenDuration)
		for _, lgGrp := range lgGrps {
			fmt.Println(*lgGrp.LogGroupName)
		}

		// logGrpsclean := logGrpsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

		// logGrpsclean.GetlogGrps()

		// switch viper.GetString(outputFlag) {
		// case "json", "JSON":
		// 	logGrpsPrintJSON(logGrpsclean.GetAllVolumes())
		// default:
		// 	logGrpsPrintTable(logGrpsclean.GetAllVolumes())
		// }
	},
}

func logGrpsBindFlags() {
	logGrpsCmd.AddCommand(logGrpsDeleteCmd)
	logGrpsCmd.AddCommand(logGrpsListCmd)
	rootCmd.AddCommand(logGrpsCmd)

	const objType = "logGrps volumes"

	logGrpsDeleteCmdFlags := logGrpsDeleteCmd.Flags()
	deleteOnlyFlags(logGrpsDeleteCmdFlags)

	logGrpsListCmdFlags := logGrpsListCmd.Flags()
	listOnlyFlags(logGrpsListCmdFlags, objType)

	err := viper.BindPFlags(logGrpsListCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)

	err = viper.BindPFlags(logGrpsDeleteCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)
}

func logGrpsPrintTable(vols []ec2Types.Volume) {
	grpsTable := table.New("Volume ID", "Creation Datetime", "State")
	for _, vol := range vols {
		// TODO: Conditionally add Tags here.
		grpsTable.AddRow(*vol.VolumeId, vol.CreateTime.Format(time.RFC3339), vol.State)
	}
	grpsTable.Print()
}

func logGrpsPrintJSON(vols []ec2Types.Volume) {
	out, err := json.Marshal(vols)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Json.Marshal(vols) failed: %s", err)
	fmt.Print(string(out))
}
