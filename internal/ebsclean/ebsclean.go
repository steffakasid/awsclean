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
}

func NewInstance(awsClient *internal.AWS, olderthen time.Duration, dryrun bool) *EBSClean {
	return &EBSClean{
		awsClient: awsClient,
		olderthen: olderthen,
		dryrun:    dryrun,
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
		if volume.State != types.VolumeStateInUse {
			if volume.CreateTime.Before(olderThenDate) {
				// now we could delete!
				fmt.Printf("Delete %s creationDate %v type %s state %s\n", *volume.VolumeId, volume.CreateTime, volume.VolumeType, volume.State)
				e.awsClient.DeleteVolume(*volume.VolumeId, e.dryrun)
				deleted++
			} else {
				fmt.Printf("Skipping %s creationDate %v type %s state %s\n", *volume.VolumeId, volume.CreateTime, volume.VolumeType, volume.State)
				skipped++
			}
		} else {
			fmt.Printf("Filtered out %s creationDate %v type %s state %s\n", *volume.VolumeId, volume.CreateTime, volume.VolumeType, volume.State)
			filtered++
		}
	}
	logger.Debugf("Deleted %d, Skipped %d, Filtered out %d EBS volumes", deleted, skipped, filtered)
}
