package internal

import (
	"time"

	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/xhit/go-str2duration/v2"
)

func ParseDuration(str string) time.Duration {
	duration, err := str2duration.ParseDuration(str)
	extendedslog.Logger.Fatalf("error on str2duration.ParseDuration(str): %w", err)
	return duration
}
