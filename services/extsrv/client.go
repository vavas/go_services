package extsrv

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"encoding/json"
	"net/http"
	"net/url"

	"github.com/vavas/go_services/gnats"
	"github.com/vavas/go_services/logger"
	"github.com/vavas/go_services/services/auth"
	"github.com/vavas/go_services/services/internal"
)

// Request structure
type Request struct {
	Service   string          `json:"service"`          // The service name such as 'edm.external' or 'billing.external'
	RequestID string          `json:"request_id"`       // GUID to identify request through service chaining and resp
	RequestIP string          `json:"request_ip"`       // IP address of the client
	Method    string          `json:"method"`           // The method delete/put/patch/post/get
	Path      string          `json:"path"`             // The path of the function ie '/users/123/delete'
	Query     url.Values      `json:"query,omitempty"`  // The gateway should parse this for you. The path of the function ie 'foo=bar&something=something'
	RawAuth   json.RawMessage `json:"auth,omitempty"`   // The auth data from auth service
	Body      interface{}     `json:"body,omitempty"`   // The request body (HTTP body)
	Header    http.Header     `json:"header,omitempty"` // the request header (HTTP header)

	Auth    *auth.Auth             `json:"-"` // The parsed auth data from auth service
	Params  map[string]string      `json:"-"`
	BodyMap map[string]interface{} `json:"-"` // Most of the time the body is a map, so it's made ready here

	RawRequest json.RawMessage `json:"-"` // The raw []byte of the request
}

// Response structure
type Response struct {
	StatusCode int               `json:"status_code"`       // HTTP status codes
	Headers    map[string]string `json:"headers,omitempty"` // HTTP headers
	Body       interface{}       `json:"body,omitempty"`    // The response body
}

// RequestReply makes a request and expects a reply from a service
func RequestReply(req *Request, resp interface{}) error {
	enc, err := gnats.JSONConn()

	if err != nil {
		return err
	}

	enc.Flush()

	// Services like billing service can be slow because it depends on 3rd party server (Stripe server)
	if err := enc.Request(req.subject(), req, resp, internal.DefaultTimeout); err != nil {
		logger.Logger.Info("ext.RequestReply() > enc.Request() error",
			zap.Error(err),
			zap.String("request_id", req.RequestID),
			zap.String("request_ip", req.RequestIP),
			zap.String("request_method", req.Method),
		)
		return err
	}

	return nil
}

func (req *Request) subject() string {
	return subject(req.Service)
}

// Param returns an argument from path parameter or query string.
func (req *Request) Param(key string) string {
	if value, ok := req.Params[key]; ok {
		return value
	}

	return req.Query.Get(key)
}

// ParamObjectID gets a param value as Object ID
func (req *Request) ParamObjectID(key string) (primitive.ObjectID, error) {
	val := req.Param(key)

	id, err := primitive.ObjectIDFromHex(val)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf(`"%s" is not a valid id`, val)
	}

	//if !bson.IsObjectIdHex(val) {
	//	return "", fmt.Errorf(`"%s" is not a valid id`, val)
	//}
	return id, nil
}


// SetAuth set json encoded auth data to the request.
func (req *Request) SetAuth(authData *auth.Auth) (err error) {
	req.RawAuth, err = json.Marshal(authData)
	return err
}