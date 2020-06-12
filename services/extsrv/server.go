// Package extsrv is a library for implementing & using nats-based external services.
// External services are HTTP-based API services that the client can call.
// An HTTP server will act as a gateway to these external services.
package extsrv

import (

	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/vavas/go_services/db"
	"github.com/vavas/go_services/logger"
	"github.com/vavas/go_services/services/auth"
	"github.com/vavas/go_services/services/internal"
	"github.com/vavas/go_services/utils"
)

// HandlerFunc is the function signature for external service handlers
type HandlerFunc func(dbc *mongo.Database, req *Request) (resp *Response)

// Handler structure
type Handler struct {
	Method      string
	Path        *regexp.Regexp // Use numbered capturing group.
	NoAuth      bool
	RawBody     bool
	HandlerFunc HandlerFunc
}

// Subscription structure
type Subscription struct {
	sync.Mutex
	Service  string
	Subject  string
	Queue    string
	handlers []*Handler
}

func subject(service string) string {
	return service + ".external"
}

func subscribe(service string, subject string, queue string) (sub *Subscription) {
	sub = &Subscription{Service: service, Subject: subject, Queue: queue, handlers: []*Handler{}}

	internal.Subscribe(subject, queue, func(msg *nats.Msg) {

		start := time.Now()

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
				internal.Reply(msg, &Response{StatusCode: http.StatusInternalServerError})
			}
		}()

		req := &Request{}
		if err := json.Unmarshal(msg.Data, req); err != nil {
			internal.LogError(err, msg, "ext.subscribe() > json.Unmarshal() error")
			internal.Reply(msg, &Response{StatusCode: http.StatusBadRequest})
			return
		}
		req.RawRequest = msg.Data

		logFields := map[string]interface{}{
			"request":    "external",
			"request_id": req.RequestID,
			"method":     req.Method,
			"path":       req.Path,
		}
		if len(req.Query) > 0 {
			logFields["query"] = req.Query
		}
		logger.Logger.Debug("External request is handling",
			zap.String("request",    "external"),
			zap.String("request_id", req.RequestID))

		handler, params := sub.getHandler(req)

		if handler == nil {
			err := errors.New("handler not found")
			internal.LogError(err, msg, "ext.subscribe() error")
			internal.Reply(msg, &Response{StatusCode: http.StatusNotFound})
			return
		}

		if req.RawAuth != nil {
			req.Auth = &auth.Auth{}
			if err := json.Unmarshal(req.RawAuth, req.Auth); err != nil {
				req.Auth = nil
			}
		}
		if !handler.NoAuth {
			if req.Auth == nil {
				err := errors.New("request is unauthorized")
				internal.LogError(err, msg, "request is unauthorized")
				internal.Reply(msg, &Response{StatusCode: http.StatusUnauthorized})
				return
			}
		}

		if !handler.RawBody && req.Body != nil {
			if body, ok := req.Body.(map[string]interface{}); ok {
				req.BodyMap = body
			} else {
				err := errors.New("request is bad request")
				internal.LogError(err, msg, "request is bad request")
				internal.Reply(msg, &Response{StatusCode: http.StatusBadRequest})
				return
			}
		}

		req.Params = params

		var dbc *mongo.Database
		if db.HasClient() {
			dbc = db.DB()
		}

		resp := handler.HandlerFunc(dbc, req)

		internal.Reply(msg, resp)

		end := time.Now()
		latency := end.Sub(start)
		timeFormatted := end.Format("2006-01-02 15:04:05")

		logFields["time"] = timeFormatted
		logFields["latency"] = latency.String()
		logFields["status"] = resp.StatusCode
		if len(resp.Headers) > 0 {
			logFields["headers"] = resp.Headers
		}

		logger.Logger.Debug("External request completed",
			zap.Any("logFields",    logFields))

		if latency > internal.DefaultTimeout {
			logFields["request"] = req
			utils.NotifyError(internal.ErrTooLong, logFields)
		}
	})

	return sub
}

func (s *Subscription) addHandler(handler *Handler) {
	s.Lock()
	defer s.Unlock()

	s.handlers = append(s.handlers, handler)
}

// AddHandler adds handler for external service
func (s *Subscription) AddHandler(method string, path *regexp.Regexp, handleFunc HandlerFunc) {
	s.addHandler(&Handler{Method: method, Path: path, NoAuth: false, RawBody: false, HandlerFunc: handleFunc})
}

// AddPublicHandler adds handler for external service, no auth.
func (s *Subscription) AddPublicHandler(method string, path *regexp.Regexp, handleFunc HandlerFunc) {
	s.addHandler(&Handler{Method: method, Path: path, NoAuth: true, RawBody: false, HandlerFunc: handleFunc})
}

// AddRawHandler adds handler for external service, raw body.
func (s *Subscription) AddRawHandler(method string, path *regexp.Regexp, handleFunc HandlerFunc) {
	s.addHandler(&Handler{Method: method, Path: path, NoAuth: false, RawBody: true, HandlerFunc: handleFunc})
}

func (s *Subscription) getHandler(req *Request) (*Handler, map[string]string) {
	s.Lock()
	defer s.Unlock()

	var handler *Handler
	params := map[string]string{}

	for _, h := range s.handlers {
		if req.Method == h.Method {
			match := h.Path.FindStringSubmatch(req.Path)
			if len(match) > 0 {
				handler = h
				for i, name := range h.Path.SubexpNames() {
					if i != 0 {
						params[name] = match[i]
					}
				}
				break
			}
		}
	}

	return handler, params
}

// QueueSubscribe create a queued subscription to gnats, which only one server will receive the message.
func QueueSubscribe(service string) (sub *Subscription) {
	return subscribe(service, subject(service), subject(service))
}
