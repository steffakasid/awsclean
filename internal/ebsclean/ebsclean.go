package ebsclean

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	logger "github.com/sirupsen/logrus"
	"github.com/steffakasid/amiclean/internal"
)

type EBSClean struct {
	awsClient *internal.AWS
	olderthen time.Duration
	dryrun    bool
	showTags  bool
}

func NewInstance(awsClient *internal.AWS, olderthen time.Duration, dryrun bool, showTags bool) *EBSClean {
	return &EBSClean{
		awsClient: awsClient,
		olderthen: olderthen,
		dryrun:    dryrun,
		showTags:  showTags,
	}
}

func (e EBSClean) DeleteUnusedEBSVolumes() {
	ebsVolumes := e.awsClient.GetAvailableEBSVolumes()

	today := time.Now()
	olderThenDate := today.Add(e.olderthen * -1)
	logger.Debugf("OlderThenDate %v", olderThenDate)

	deleted := 0
	skipped := 0
	filtered := 0
	for _, volume := range ebsVolumes {
		details := fmt.Sprintf("%s creationDate: %v\ttype: %s\tstate: %s\t", *volume.VolumeId, volume.CreateTime, volume.VolumeType, volume.State)
		if e.showTags {
			details += "\n\ttags:\t"
			for i, tag := range volume.Tags {
				var pattern string
				if i == 0 {
					pattern = "%s: %s\n"
				} else if i == len(volume.Tags)-1 {
					pattern = "\t\t%s: %s"
				} else {
					pattern = "\t\t%s: %s\n"
				}
				details += fmt.Sprintf(pattern, *tag.Key, *tag.Value)
			}
		}
		if volume.State != types.VolumeStateInUse {
			if volume.CreateTime.Before(olderThenDate) {
				// now we could delete!
				logger.Infof("Delete %s", details)
				err := e.awsClient.DeleteVolume(*volume.VolumeId, e.dryrun)
				if err != nil {
					logger.Errorf("error deleting volume: %s", err)
				}
				fmt.Println()
				deleted++
			} else {
				logger.Infof("Skipping %s", details)
				fmt.Println()
				skipped++
			}
		} else {
			logger.Infof("Filtered out %s\n\n", details)
			fmt.Println()
			filtered++
		}
	}
	logger.Infof("Deleted %d, Skipped %d, Filtered out %d EBS volumes", deleted, skipped, filtered)
}
