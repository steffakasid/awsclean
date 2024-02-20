package internal

import (
	"fmt"
	"time"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	extendedslog "github.com/steffakasid/extended-slog"
)

type SecurityGroup struct {
	*ec2Types.SecurityGroup
	CreationTime        *time.Time
	Creator             string
	IsUsed              bool
	AttachedToNetIfaces []string
}

type SecurityGroups = map[string]*SecurityGroup

func AddOrUpdate(grps SecurityGroups, grpDetail *ec2Types.SecurityGroup, creator string, creationTime *time.Time, isUsed bool, netIfaces []string) {

	if grp, isMapContainsKey := grps[*grpDetail.GroupName]; isMapContainsKey {

		src := &SecurityGroup{
			SecurityGroup:       grpDetail,
			Creator:             creator,
			CreationTime:        creationTime,
			IsUsed:              isUsed,
			AttachedToNetIfaces: netIfaces,
		}

		err := mergeFields(src, grp)
		CheckError(err, extendedslog.Logger.Errorf)

	} else {
		grps[*grpDetail.GroupName] = &SecurityGroup{
			SecurityGroup:       grpDetail,
			CreationTime:        creationTime,
			Creator:             creator,
			IsUsed:              isUsed,
			AttachedToNetIfaces: netIfaces,
		}
	}
}

func AppendAll(src, target SecurityGroups) {
	for key, val := range src {
		if tgtObj, exists := target[key]; exists {
			err := mergeFields(val, tgtObj)
			if err != nil {
				extendedslog.Logger.Error(err)
			}
			target[key] = tgtObj
		} else {
			target[key] = val
		}
	}
}

func mergeFields(src, tgt *SecurityGroup) error {

	if src.SecurityGroup == nil && tgt.SecurityGroup == nil {
		return fmt.Errorf("error mergig SecurityGroups. Both objects have obj.SecurityGroup = nil")
	}

	if src.SecurityGroup != tgt.SecurityGroup {
		if src.SecurityGroup == nil && tgt.SecurityGroup != nil {
			// nothing to do as target alrwady have the data
		} else if src.SecurityGroup != nil && tgt.SecurityGroup == nil {
			tgt.SecurityGroup = src.SecurityGroup
		} else {
			// do not merge objects...
		}

		// We use GroupName instead of GroupID becauce from CloudTrail we only get the name
		if src.SecurityGroup != nil && *src.SecurityGroup.GroupName != *tgt.SecurityGroup.GroupName {
			return fmt.Errorf("error mergig SecurityGroups: %s != %s", *src.SecurityGroup.GroupName, *tgt.SecurityGroup.GroupName)
		}
	}

	if src.Creator != "" && tgt.Creator == "" {
		tgt.Creator = src.Creator
	}

	if src.CreationTime != nil && tgt.CreationTime == nil {
		tgt.CreationTime = src.CreationTime
	}

	if len(src.AttachedToNetIfaces) > 0 && len(tgt.AttachedToNetIfaces) == 0 {
		tgt.AttachedToNetIfaces = src.AttachedToNetIfaces
	}

	return nil
}
