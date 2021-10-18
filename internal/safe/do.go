package safe

import (
	log "github.com/sirupsen/logrus"
	"go.octolab.org/errors"
	"go.octolab.org/safe"
)

// Do should be used to catch panic in goroutines
func Do(f func()) {
	safe.Do(
		// always returns nil, usual errors should be logged inside f
		func() error {
			f()
			return nil
		},
		func(err error) {
			if recovered, is := errors.Unwrap(err).(errors.Recovered); is {
				log.WithError(err).WithField("cause", recovered.Cause()).Error("panic in goroutine")
			}
		},
	)
}
