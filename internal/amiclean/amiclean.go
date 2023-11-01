package amiclean

import (
	"time"

	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/eslog"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AmiClean struct {
	awsClient      *internal.AWS
	olderthen      time.Duration
	awsaccount     string
	dryrun         bool
	useLaunchTpls  bool
	onlyUnused     bool
	usedAMIs       []ec2Types.Image
	unusedAMIs     []ec2Types.Image
	ignorePatterns []string
}

func NewInstance(
	awsClient *internal.AWS,
	olderthen time.Duration, awsaccount string,
	dryrun bool,
	onlyunused bool,
	useLaunchTpls bool,
	ignorePatterns []string) *AmiClean {

	return &AmiClean{
		awsClient:      awsClient,
		olderthen:      olderthen,
		awsaccount:     awsaccount,
		dryrun:         dryrun,
		onlyUnused:     onlyunused,
		useLaunchTpls:  useLaunchTpls,
		usedAMIs:       []ec2Types.Image{},
		unusedAMIs:     []ec2Types.Image{},
		ignorePatterns: ignorePatterns,
	}
}

func (a *AmiClean) GetAMIs() error {
	usedAMIs := []string{}
	usedAMIs = append(usedAMIs, a.awsClient.GetUsedAMIsFromEC2()...)

	if a.useLaunchTpls {
		usedAMIs = append(usedAMIs, a.awsClient.GetUsedAMIsFromLaunchTpls()...)
	}

	images, err := a.awsClient.DescribeImages(a.awsaccount)
	if err != nil {
		return err
	}

	for _, image := range images {
		if !internal.Contains(usedAMIs, *image.ImageId) {

			if err != nil {
				eslog.Error(err)
			}
			a.unusedAMIs = append(a.unusedAMIs, image)
		} else {
			a.usedAMIs = append(a.usedAMIs, image)
			eslog.Logger.Infof("Ignored %s", *image.ImageId)
		}
	}
	return nil
}

func (a AmiClean) GetAllAMIs() []ec2Types.Image {
	all := []ec2Types.Image{}

	all = append(all, a.unusedAMIs...)
	if !a.onlyUnused {
		all = append(all, a.usedAMIs...)
	}

	return all
}

func (a AmiClean) DeleteOlderUnusedAMIs() error {
	err := a.GetAMIs()
	if err != nil {
		return err
	}

	today := time.Now()
	olderThenDate := today.Add(a.olderthen * -1)

	for _, ami := range a.unusedAMIs {
		ok, err := internal.MatchAny(*ami.Name, a.ignorePatterns)
		if err != nil {
			return err
		}
		if !ok {
			creationDate, err := time.Parse("2006-01-02T15:04:05.000Z", *ami.CreationDate)
			if err != nil {
				return err
			}
			if creationDate.Before(olderThenDate) {
				eslog.Logger.Infof("Delete %s:%s as it's creationdate %s is older then %s", *ami.ImageId, *ami.Name, *ami.CreationDate, olderThenDate.String())
				err = a.awsClient.DeregisterImage(*ami.ImageId, a.dryrun)
				eslog.LogIfErrorf(err, eslog.Errorf, "Error on DeregisterImage(): %s")
			} else {
				eslog.Logger.Infof("Keeping %s:%s as it's creationdate %s is newer then %s", *ami.ImageId, *ami.Name, *ami.CreationDate, olderThenDate.String())
			}
		} else {
			eslog.Logger.Infof("Skipping %s-%s", *ami.ImageId, *ami.Name)
		}
	}
	return nil
}
