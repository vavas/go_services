// Package intsrv is a library for implementing & using nats-based internal services.
// For example, getting user and/or account based on an id.
package intsrv

import (

	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/vavas/go_services/db"
	"github.com/vavas/go_services/logger"
	"github.com/vavas/go_services/services/internal"
	"github.com/vavas/go_services/utils"
)

// HandlerFunc is the function signature for internal service handlers
type HandlerFunc func(dbc *mongo.Database, req *Request) (resp *Response, err error)

// Handler structure
type Handler struct {
	Function    string
	HandlerFunc HandlerFunc
}

// Subscription structure
type Subscription struct {
	sync.Mutex
	Service  string
	Subject  string
	Queue    string
	handlers map[string]*Handler
}

func subject(service string) string {
	return service + ".internal"
}

func subscribe(service string, subject string, queue string) (sub *Subscription) {
	sub = &Subscription{Service: service, Subject: subject, Queue: queue, handlers: map[string]*Handler{}}

	internal.Subscribe(subject, queue, func(msg *nats.Msg) {

		start := time.Now()

		var respErr string

		req := &Request{}
		if err := json.Unmarshal(msg.Data, req); err != nil {
			internal.LogError(err, msg, "int.subscribe() > json.Unmarshal() error")
			internal.Reply(msg, &Response{Error: err.Error()})
			respErr = err.Error()
			return
		}

		logFields := map[string]interface{}{
			"request":    "internal",
			"function":   req.Function,
			"request_id": req.RequestID,
		}

		logger.Logger.Debug("Internal request is handling",
			zap.Any("logFields", logFields))

		defer func() {
			if panicVal := recover(); panicVal != nil {
				err, ok := panicVal.(error)
				if !ok {
					err = fmt.Errorf("%+v", panicVal)
				}

				meta := utils.M{"error": err}
				if msg != nil {
					meta["request"] = string(msg.Data)
					meta["subject"] = msg.Subject
				}
				utils.NotifyError(err, meta)
				internal.Reply(msg, &Response{Error: err.Error()})
			}

			end := time.Now()
			latency := end.Sub(start)
			timeFormatted := end.Format("2006-01-02 15:04:05")

			logFields["time"] = timeFormatted
			logFields["latency"] = latency.String()
			if len(respErr) > 0 {
				logFields["error"] = respErr
			}

			logger.Logger.Debug("Internal request is completed",
				zap.Any("logFields", logFields))

			if latency > internal.DefaultTimeout {
				logFields["request"] = req
				utils.NotifyError(internal.ErrTooLong, logFields)
			}
		}()

		h := sub.getHandler(req)
		if h == nil {
			err := errors.New("handler not found")
			internal.LogError(err, msg, "int.subscribe() > json.Unmarshal() error")
			internal.Reply(msg, &Response{Error: err.Error()})
			respErr = err.Error()
			return
		}

		var dbc *mongo.Database
		if db.HasClient() {
			dbc = db.DB()
			//defer dbc.Session.Close()
		}

		resp, err := h.HandlerFunc(dbc, req)
		if err != nil {
			internal.LogError(err, msg, "int.subscribe() > handler() error")
			if resp == nil {
				resp = &Response{}
			}
			resp.Error = err.Error()
		}

		internal.Reply(msg, resp)

		if resp != nil {
			respErr = resp.Error
		}
	})

	return sub
}

// AddHandler adds handler for internal service
func (s *Subscription) AddHandler(name string, handler HandlerFunc) {
	s.Lock()
	defer s.Unlock()

	h := Handler{name, handler}

	s.handlers[name] = &h
}

func (s *Subscription) getHandler(req *Request) *Handler {
	s.Lock()
	defer s.Unlock()

	h, _ := s.handlers[req.Function]

	return h
}

// Subscribe create a subscription to gnatsd.
func Subscribe(service string) (sub *Subscription) {
	return subscribe(service, subject(service), "")
}

// QueueSubscribe create a queued subscription to gnats, which only one server will receive the message.
func QueueSubscribe(service string) (sub *Subscription) {
	return subscribe(service, subject(service), subject(service))
}
