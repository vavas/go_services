package intsrv

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/vavas/go_services/logger"
	"go.uber.org/zap"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/vavas/go_services/gnats"
	"github.com/vavas/go_services/services/internal"
)

// Request structure
type Request struct {
	Service   string      `json:"service"`    // The service name such as 'edm' or 'billing'
	Function  string      `json:"function"`   // The function name
	Arguments interface{} `json:"arguments"`  // The arguments for the function
	RequestID string      `json:"request_id"` // GUID to identify request through service chaining and resp
}

// Response structure
type Response struct {
	Error string      `json:"error,omitempty"`
	Body  interface{} `json:"body,omitempty"`
}

func (req *Request) subject() string {
	return subject(req.Service)
}

// Param returns an argument from Arguments.
func (req *Request) Param(key string) interface{} {
	args, _ := req.Arguments.(map[string]interface{})
	result, _ := args[key]
	return result
}

// ParamObjectID gets a param value as Object ID
func (req *Request) ParamObjectID(key string) (bson.ObjectId, error) {
	val := req.Param(key)
	valStr, _ := val.(string)
	if !bson.IsObjectIdHex(valStr) {
		return "", fmt.Errorf(`"%s" is not a valid id`, val)
	}
	return bson.ObjectIdHex(valStr), nil
}

// Publish makes a request and does not expect a reply from a service
func Publish(req *Request) error {
	enc, err := gnats.JSONConn()
	if err != nil {
		return err
	}
	logger.Logger.Debug("Request internal without reply",
		zap.Error(err),
		zap.String("request", "internal"),
		zap.String("service", req.Service),
		zap.String("method", req.Function),
		zap.String("request_id", req.RequestID),
	)
	if err := enc.Publish(req.subject(), req); err != nil {
		return errors.Wrap(err, "enc.Publish")
	}

	return nil
}

// RequestReply makes a request and expects a reply from a service
func RequestReply(req *Request, resp interface{}, customTimeout ...time.Duration) error {
	enc, err := gnats.JSONConn()
	if err != nil {
		return err
	}

	logger.Logger.Debug("Request internal with reply",
		zap.Error(err),
		zap.String("request", "internal"),
		zap.String("service", req.Service),
		zap.String("method", req.Function),
		zap.String("request_id", req.RequestID),
	)
	realTimeout := internal.DefaultTimeout
	if len(customTimeout) > 0 {
		realTimeout = customTimeout[0]
	}

	if err := enc.Flush(); err != nil {
		return errors.Wrap(err, "enc.Flush")
	}

	if err := enc.Request(req.subject(), req, resp, realTimeout); err != nil {
		if strings.Contains(err.Error(), "nats: timeout") {
			if err := enc.Request(req.subject(), req, resp, realTimeout); err != nil {
				return errors.Wrap(err, "enc.Request")
			}
		} else {
			return errors.Wrap(err, "enc.Request")
		}
	}

	return nil
}
