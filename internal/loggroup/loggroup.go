package loggroup

import (
	"strings"
	"time"

	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/eslog"

	cwlogsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type LogGrp struct {
	awsClient  *internal.AWS
	olderthen  *time.Duration
	exclude    string
	dryrun     bool
	onlyUnused bool
	used       []cwlogsTypes.LogGroup
	unused     []cwlogsTypes.LogGroup
}

func NewInstance(awsClient *internal.AWS, olderthen *time.Duration, dryrun, onlyUnused bool) *LogGrp {
	return &LogGrp{
		awsClient:  awsClient,
		olderthen:  olderthen,
		exclude:    "cdaas-agent-aws-cdk-production",
		dryrun:     dryrun,
		onlyUnused: onlyUnused,
		used:       []cwlogsTypes.LogGroup{},
		unused:     []cwlogsTypes.LogGroup{},
	}
}

func (l LogGrp) DeleteUnused() {

	l.GetCloudWatchLogGroups()

	for _, lgGrp := range l.GetUnusedLogGroups() {
		err := l.awsClient.DeleteLogGroup(lgGrp.LogGroupName, l.dryrun)
		eslog.LogIfError(err, eslog.Error)
	}

}

func (l *LogGrp) GetCloudWatchLogGroups() {
	unused := []cwlogsTypes.LogGroup{}
	used := []cwlogsTypes.LogGroup{}
	olderThen := time.Now().Add(-1 * *l.olderthen)

	eslog.Debug("GetCloudWatchLogGroups()")
	logGrps := l.awsClient.ListLogGrps()

	for _, lgGrp := range logGrps {
		eslog.Debugf("CreationTime %d < olderThen %d", *lgGrp.CreationTime, olderThen.UnixMilli())
		if *lgGrp.CreationTime < olderThen.UnixMilli() {
			if !strings.Contains(*lgGrp.LogGroupName, l.exclude) {
				eslog.Debugf("Add logGroup: %s", *lgGrp.LogGroupName)
				unused = append(unused, lgGrp)
			} else {
				eslog.Debugf("Used logGroup: %s", *lgGrp.LogGroupName)
				used = append(used, lgGrp)
			}
		}
	}
	l.unused = unused
	l.used = used
}

func (l LogGrp) GetUnusedLogGroups() []cwlogsTypes.LogGroup {
	return l.unused
}
