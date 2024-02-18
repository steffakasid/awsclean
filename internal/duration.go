package internal

import (
	"time"

	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/xhit/go-str2duration/v2"
)

func ParseDuration(str string) time.Duration {
	duration, err := str2duration.ParseDuration(str)
	CheckError(err, extendedslog.Logger.Fatalf)
	return duration
}
