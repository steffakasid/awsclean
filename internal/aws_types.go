package internal

import (
	"fmt"
	"time"
)

type SecurityGroup struct {
	Name                string
	ID                  string
	CreationTime        *time.Time
	Creator             string
	IsUsed              bool
	AttachedToNetIfaces []string
}

type SecurityGroups = map[string]SecurityGroup

func AddOrUpdate(grps SecurityGroups, name, ID, creator string, creationTime *time.Time, isUsed bool, netIfaces []string) {

	if grp, isMapContainsKey := grps[name]; isMapContainsKey {

		if name != "" {
			grp.Name = name
		}
		if ID != "" {
			grp.ID = ID
		}
		if creator != "" {
			grp.Creator = creator
		}
		if creationTime != nil {
			grp.CreationTime = creationTime
		}
		grp.IsUsed = isUsed
		grp.AttachedToNetIfaces = netIfaces
		grps[name] = grp
	} else {
		grps[name] = SecurityGroup{
			Name:                name,
			ID:                  ID,
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
			mergeFields(&val, &tgtObj)
			target[key] = tgtObj
		} else {
			target[key] = val
		}
	}
}

func mergeFields(src, tgt *SecurityGroup) error {
	if src.Name != tgt.Name {
		return fmt.Errorf("error mergig SecurityGroups: %s != %s", src.Name, tgt.Name)
	}

	if src.ID != "" && tgt.ID == "" {
		tgt.ID = src.ID
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
