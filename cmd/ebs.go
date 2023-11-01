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
	eslog "github.com/steffakasid/eslog"
)

const (
	ebsCmdName       = "ebs"
	ebsDeleteCmdName = "delete"
	ebsListCmdName   = "list"
)

var (
	ebsListCmdAliases   = []string{"ls"}
	ebsDeleteCmdAliases = []string{"del"}
)

var (
	ebsDeleteCmdExamples = fmt.Sprintf(`
  %[1]s %[2]s %[3]s --older-then 5w  delete all EBS volumes which are older then 5w and are not bound
  %[1]s %[2]s %[3]s --dry-run        do not delete any EBS volume just show what should be done
  %[1]s %[2]s %[4]s --older-then 5w     delete all EBS volumes which are older then 5w and are not bound
  %[1]s %[2]s %[4]s --dry-run           do not delete any EBS volume just show what should be done
  `,
		binaryname,
		ebsCmdName,
		ebsDeleteCmdName,
		ebsDeleteCmdAliases[0])
	ebsListCmdExamples = fmt.Sprintf(`
	%[1]s %[2]s %[3]s --show-tags      print out tags of EBS volumes
	%[1]s %[2]s %[4]s --show-tags        print out tags of EBS volumes
	`,
		binaryname,
		ebsCmdName,
		ebsListCmdName,
		ebsListCmdAliases[0])
)

// ebsCmd represents the ebs command
var ebsCmd = &cobra.Command{
	Use:   ebsCmdName,
	Short: "Cleanup unused EBS volumes",
	Long: fmt.Sprintf(`This tool can be used to list or cleanup old and unbound Elastic Block Store (EBS) volumes.

Examples:
%s%s
`,
		ebsDeleteCmdExamples,
		ebsListCmdExamples),
}

var ebsDeleteCmd = &cobra.Command{
	Use:     ebsDeleteCmdName,
	Aliases: ebsDeleteCmdAliases,
	Short:   "Cleanup unused EBS volumes",
	Long: fmt.Sprintf(`This tool can be used to cleanup old and unbound Elastic Block Store (EBS) volumes.

Examples:
%s`,
		ebsDeleteCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()
		ebsclean := ebsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

		ebsclean.DeleteUnusedEBSVolumes()
	},
}

var ebsListCmd = &cobra.Command{
	Use:     ebsListCmdName,
	Aliases: ebsListCmdAliases,
	Short:   "List unused EBS volumes",
	Long: fmt.Sprintf(`This command can be used to list unused Elastic Block Store Volumes.

Examples:
%s`,
		ebsListCmdExamples),
	Run: func(cmd *cobra.Command, args []string) {
		olderthenDuration := internal.ParseDuration(viper.GetString(olderthenFlag))

		awsClient := internal.NewAWSClient()
		ebsclean := ebsclean.NewInstance(awsClient, olderthenDuration, viper.GetBool(dryrunFlag), viper.GetBool(onlyUnusedFlag))

		ebsclean.GetEBSVolumes()

		switch viper.GetString(outputFlag) {
		case "json", "JSON":
			ebsPrintJSON(ebsclean.GetAllVolumes())
		default:
			ebsPrintTable(ebsclean.GetAllVolumes())
		}
	},
}

func ebsBindFlags() {
	ebsCmd.AddCommand(ebsDeleteCmd)
	ebsCmd.AddCommand(ebsListCmd)
	rootCmd.AddCommand(ebsCmd)

	const objType = "EBS volumes"

	ebsDeleteCmdFlags := ebsDeleteCmd.Flags()
	deleteOnlyFlags(ebsDeleteCmdFlags)

	ebsListCmdFlags := ebsListCmd.Flags()
	listOnlyFlags(ebsListCmdFlags, objType)

	err := viper.BindPFlags(ebsListCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)

	err = viper.BindPFlags(ebsDeleteCmdFlags)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Failed to bind Flags: %w", err)
}

func ebsPrintTable(vols []ec2Types.Volume) {
	grpsTable := table.New("Volume ID", "Creation Datetime", "State")
	for _, vol := range vols {
		// TODO: Conditionally add Tags here.
		grpsTable.AddRow(*vol.VolumeId, vol.CreateTime.Format(time.RFC3339), vol.State)
	}
	grpsTable.Print()
}

func ebsPrintJSON(vols []ec2Types.Volume) {
	out, err := json.Marshal(vols)
	eslog.LogIfErrorf(err, eslog.Fatalf, "Json.Marshal(vols) failed: %s", err)
	fmt.Print(string(out))
}
