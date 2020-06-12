package utils

import (
	"github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
	"go.uber.org/zap"

	"github.com/vavas/go_services/logger"
)

// NotifyError logs an error and notifies bugsnag server.
func NotifyError(err error, rawData ...map[string]interface{}) {
	if err.Error() == "not found" {
		NotifyWarning(err, rawData...)
		return
	}

	log := logger.Logger
	log.Error("error notify",
		zap.Error(err),
		zap.Any("raw_data", rawData),
	)

	if len(bugsnag.Config.APIKey) > 0 {
		var metaData bugsnag.MetaData
		if len(rawData) > 0 {
			metaData = bugsnag.MetaData{"Data": rawData[0]}
		}
		err = errors.New(err, 1)
		if err := bugsnag.Notify(err, bugsnag.SeverityError, metaData); err != nil {
			log.Error("bugsnag.Notify",
				zap.Error(err),
			)
		}
	}
}

// NotifyWarning logs a warning and notifies bugsnag server.
func NotifyWarning(err error, rawData ...map[string]interface{}) {
	log := logger.Logger
	log.Warn("warning notify",
		zap.Error(err),
		zap.Any("raw_data", rawData),
	)

	if err.Error() == "not found" {
		return
	}

	if len(bugsnag.Config.APIKey) > 0 {
		var metaData bugsnag.MetaData
		if len(rawData) > 0 {
			metaData = bugsnag.MetaData{"Data": rawData[0]}
		}
		err = errors.New(err, 1)
		if err := bugsnag.Notify(err, bugsnag.SeverityWarning, metaData); err != nil {
			log.Warn("bugsnag.Notify",
				zap.Error(err),
			)
		}
	}
}
