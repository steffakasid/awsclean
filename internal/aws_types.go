package internal

import "time"

type SecurityGroup struct {
	Name         string
	ID           string
	CreationTime *time.Time
}

type SecurityGroups = []SecurityGroup
