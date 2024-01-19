package internal

import "time"

type SecurityGroup struct {
	Name         string
	ID           string
	CreationTime *time.Time
	Creator      string
}

type SecurityGroups = []SecurityGroup

func AddDetailsToGrp(groupName, creator string, creationTime *time.Time, grps *SecurityGroups) {
	for _, grp := range *grps {
		if grp.Name == groupName {
			grp.CreationTime = creationTime
			grp.Creator = creator
		}
	}
}
