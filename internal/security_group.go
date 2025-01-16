package internal

import (
	"errors"
	"fmt"
	"maps"
	"time"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/eslog"
)

type SecurityGroup struct {
	*ec2Types.SecurityGroup
	CreationTime        *time.Time
	Creator             string
	IsUsed              bool
	AttachedToNetIfaces []string
}

// the key must be the SecurityGroup.GroupName of the SecurityGroup value
type SecurityGroups map[string]*SecurityGroup

func (grps SecurityGroups) AddOrUpdate(grpToAdd SecurityGroup) error {

	if grpToAdd.SecurityGroup == nil {
		return errors.New("AddOrUpdate() grpToAdd.SecurityGroup is nil")
	}
	if grpToAdd.SecurityGroup.GroupName == nil {
		return errors.New("AddOrUpdate() grpToAdd.SecurityGroup.GroupName is nil")
	}

	groupName := *grpToAdd.SecurityGroup.GroupName

	if grp, exists := grps[groupName]; exists {
		err := grp.mergeFields(grpToAdd)
		eslog.LogIfErrorf(err, eslog.Errorf, "error in AddOrUpdate(): %s", &err)
	} else {
		grps[groupName] = &grpToAdd
	}
	return nil
}

func (grps SecurityGroups) AppendAll(src SecurityGroups) {
	for key, val := range src {
		if tgtObj, exists := grps.getValueByIDorName(key); exists {
			err := tgtObj.mergeFields(*val)
			eslog.LogIfErrorf(err, eslog.Errorf, "error in AppendAll(): %s", err)

			grps[key] = tgtObj
		} else {
			grps[key] = val
		}
	}
}

func (grps SecurityGroups) UpdateIfExists(src SecurityGroups) (skipped SecurityGroups) {
	skipped = SecurityGroups{}
	for key, val := range src {
		if tgtObj, exists := grps.getValueByIDorName(key); exists {
			err := tgtObj.mergeFields(*val)
			eslog.LogIfErrorf(err, eslog.Errorf, "error in UpdateIfExists(): %s", err)

			grps[key] = tgtObj
		} else {
			eslog.Logger.Infof("%s doesn't seem to exist anymore. Skipping Update of result set.", key)
			skipped[key] = val
		}
	}
	return skipped
}

func (grps SecurityGroups) DeleteSkipped(skipped SecurityGroups) {
	maps.DeleteFunc(grps, func(k string, v *SecurityGroup) bool {
		_, exists := skipped.getValueByIDorName(k)
		return exists
	})
}

func (grps SecurityGroups) getValueByIDorName(idOrName string) (value *SecurityGroup, exists bool) {
	// if it's the name we can just get it ad key from the map by convention.
	if grp, exists := grps[idOrName]; exists {
		return grp, true
	}

	for _, grp := range grps {
		if grp.SecurityGroup != nil &&
			*grp.SecurityGroup.GroupId == idOrName {
			return grp, true
		}
	}

	return nil, false
}

// Basically add details from src to tgt.
func (tgt *SecurityGroup) mergeFields(src SecurityGroup) error {
	if src.SecurityGroup == nil && tgt.SecurityGroup == nil {
		return fmt.Errorf("error mergin SecurityGroups. Both objects have obj.SecurityGroup = nil")
	}

	if src.SecurityGroup != tgt.SecurityGroup {

		//nolint:staticcheck
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
		src.SecurityGroup != nil &&
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
