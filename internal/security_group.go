package internal

import (
	"errors"
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

const name_tag_key = "Name"

type SecurityGroups = map[string]*SecurityGroup

func AddOrUpdate(grps SecurityGroups, grpToAdd SecurityGroup) error {

	if grpToAdd.SecurityGroup == nil {
		return errors.New("AddOrUpdate() grpToAdd.SecurityGroup is nil")
	}
	if grpToAdd.SecurityGroup.GroupName == nil {
		return errors.New("AddOrUpdate() grpToAdd.SecurityGroup.GroupName is nil")
	}

	groupName := *grpToAdd.SecurityGroup.GroupName

	if grp, exists := grps[groupName]; exists {

		err := mergeFields(grpToAdd, grp)
		if err != nil {
			extendedslog.Logger.Error(fmt.Errorf("error in AddOrUpdate(): %w", err))
		}

	} else {
		grps[groupName] = &grpToAdd
	}
	return nil
}

func AppendAll(src, target SecurityGroups) {
	for key, val := range src {
		if tgtObj, exists := target[key]; exists {
			extendedslog.Logger.Debugf("Merge %v with %v", *val.SecurityGroup.GroupName, *tgtObj.SecurityGroup.GroupName)
			err := mergeFields(*val, tgtObj)
			if err != nil {
				extendedslog.Logger.Error(fmt.Errorf("error in AppendAll(): %w", err))
			}
			target[key] = tgtObj
		} else {
			target[key] = val
		}
	}
}

// Basically add details from src to tgt.
func mergeFields(src SecurityGroup, tgt *SecurityGroup) error {

	if src.SecurityGroup == nil && tgt.SecurityGroup == nil {
		return fmt.Errorf("error mergin SecurityGroups. Both objects have obj.SecurityGroup = nil")
	}

	if src.SecurityGroup != tgt.SecurityGroup {
		if src.SecurityGroup == nil && tgt.SecurityGroup != nil {
			// nothing to do as target alrwady have the data
		} else if src.SecurityGroup != nil && tgt.SecurityGroup == nil {
			tgt.SecurityGroup = src.SecurityGroup
		} else {
			// do not merge objects...
		}

		if src.SecurityGroup != nil &&
			src.SecurityGroup.GroupName != nil &&
			tgt.SecurityGroup.GroupName != nil &&
			*src.SecurityGroup.GroupName != *tgt.SecurityGroup.GroupName {
			return fmt.Errorf("error mergin SecurityGroups: %s != %s", *src.SecurityGroup.GroupName, *tgt.SecurityGroup.GroupName)
		}
	}

	// if GroupId not set this should mean other fields are also not set so we overwrite with src.
	if tgt.SecurityGroup != nil &&
		tgt.SecurityGroup.GroupId == nil &&
		src.SecurityGroup.GroupId != nil {
		tgt.SecurityGroup = src.SecurityGroup
	}

	if src.Creator != "" && tgt.Creator == "" {
		tgt.Creator = src.Creator
	}

	// Todo: needs testing
	if src.CreationTime != nil &&
		(tgt.CreationTime == nil || tgt.CreationTime.Compare(time.Time{}) == 0) {
		tgt.CreationTime = src.CreationTime
	} else if src.CreationTime == nil &&
		tgt.CreationTime == nil {
		tgt.CreationTime = &time.Time{}
	}

	if src.IsUsed && !tgt.IsUsed {
		tgt.IsUsed = src.IsUsed
	}

	if len(src.AttachedToNetIfaces) > 0 && len(tgt.AttachedToNetIfaces) == 0 {
		tgt.AttachedToNetIfaces = src.AttachedToNetIfaces
	}

	return nil
}
