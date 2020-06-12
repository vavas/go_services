package internal

import (
	"errors"
	"fmt"
	"github.com/vavas/go_services/logger"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/vavas/go_services/gnats"
)

// DefaultTimeout is 15s
var DefaultTimeout = 15 * time.Second

// ErrTooLong is the error thrown if the service took more than DefaultTimeout.
var ErrTooLong = errors.New("Service took too long to finish")

func buildSubscriber(subject string, queue string, handler nats.MsgHandler) gnats.OnActiveHandler {
	return func() error {
		enc, err := gnats.JSONConn()
		if err != nil {
			return err
		}

		log.Printf(`subscribe to subject="%s" queue="%s"`, subject, queue)

		sub, err := enc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			handler(msg)
		})
		if err != nil {
			return err
		}

		if err := sub.SetPendingLimits(10*nats.DefaultSubPendingMsgsLimit, 10*nats.DefaultSubPendingBytesLimit); err != nil {
			return err
		}

		return nil
	}
}

// Subscribe makes subscription to gnatsd.
func Subscribe(subject string, queue string, handler nats.MsgHandler) {
	name := subject + " subscriber"
	subscriber := buildSubscriber(subject, queue, handler)

	gnats.AddOnActiveHandler(name, subscriber)

	if gnats.IsConnected() {
		if err := subscriber(); err != nil {
			logger.Logger.Error("subscriber() error",
				zap.NamedError("error", err),
				zap.String("name", name),
			)
		}
	}
}

// Reply sends a reply to gnatsd server.
func Reply(in *nats.Msg, out interface{}) {
	if len(in.Reply) == 0 {
		return
	}

	enc, err := gnats.JSONConn()
	if err != nil {
		LogError(err, in, "reply() > gnats.JSONConn() error")
		return
	}

	if err := enc.Publish(in.Reply, out); err != nil {
		LogError(err, in, "reply() > enc.Publish() error")
		if strings.Contains(err.Error(), "json") {
			dataStr := fmt.Sprintf("%#v", out)
			logger.Logger.Error("reply() > enc.Publish() error",
				zap.NamedError("error", err),
				zap.Any("data", out),
				zap.Any("string_data", dataStr),
			)
		}
		return
	}
}

// LogError logs an error.
func LogError(err error, in *nats.Msg, text string) {
	logger.Logger.Error(text,
		zap.NamedError("error", err),
		zap.String("subject", in.Subject),
		zap.String("reply", in.Reply),
		zap.String("data", string(in.Data)),
	)
}
