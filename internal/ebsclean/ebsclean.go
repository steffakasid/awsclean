package ebsclean

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/awsclean/internal"
	extendedslog "github.com/steffakasid/extended-slog"
)

type EBSClean struct {
	awsClient     *internal.AWS
	olderthen     time.Duration
	dryrun        bool
	usedVolumes   []types.Volume
	unusedVolumes []types.Volume
}

func NewInstance(awsClient *internal.AWS, olderthen time.Duration, dryrun bool) *EBSClean {
	return &EBSClean{
		awsClient: awsClient,
		olderthen: olderthen,
		dryrun:    dryrun,
	}
}

func (e *EBSClean) GetEBSVolumes() {
	allVolumes := e.awsClient.GetAvailableEBSVolumes()

	for _, volume := range allVolumes {
		if volume.State != types.VolumeStateInUse {
			e.unusedVolumes = append(e.unusedVolumes, volume)
		} else {
			e.usedVolumes = append(e.usedVolumes, volume)
			extendedslog.Logger.Infof("In use:%s\n\n", *volume.VolumeId)
		}
	}
}

func (e EBSClean) GetAllVolumes() []types.Volume {
	all := []types.Volume{}

	all = append(all, e.unusedVolumes...)
	all = append(all, e.usedVolumes...)

	return all
}

func (e EBSClean) DeleteUnusedEBSVolumes() {
	e.GetEBSVolumes()

	deleted := 0
	skipped := 0

	today := time.Now()
	olderThenDate := today.Add(e.olderthen * -1)
	extendedslog.Logger.Debugf("OlderThenDate %v", olderThenDate)

	for _, volume := range e.unusedVolumes {

		if volume.CreateTime.Before(olderThenDate) {
			extendedslog.Logger.Infof("Delete %s", *volume.VolumeId)
			err := e.awsClient.DeleteVolume(*volume.VolumeId, e.dryrun)
			if err != nil {
				extendedslog.Logger.Errorf("error deleting volume: %s", err)
			}
			fmt.Println()
			deleted++
		} else {
			extendedslog.Logger.Infof("Skipping %s", *volume.VolumeId)
			fmt.Println()
			skipped++
		}
	}

	extendedslog.Logger.Infof("Deleted %d, Skipped %d EBS volumes", deleted, skipped)
}
