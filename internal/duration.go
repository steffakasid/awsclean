package internal

import (
	"time"

	eslog "github.com/steffakasid/eslog"
	"github.com/xhit/go-str2duration/v2"
)

func ParseDuration(str string) time.Duration {
	duration, err := str2duration.ParseDuration(str)
	eslog.Logger.Fatalf("error on str2duration.ParseDuration(str): %w", err)
	return duration
}
