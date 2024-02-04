package internal

import "time"

type Time interface {
	Add(d time.Duration) time.Time
	GetTimeP() *time.Time
}

type MyTime struct {
	*time.Time
}

func NewTime(t time.Time) Time {
	return &MyTime{
		&t,
	}
}

func (t *MyTime) Add(d time.Duration) time.Time {
	return t.Time.Add(d)
}

func (t MyTime) GetTimeP() *time.Time {
	return t.Time
}
