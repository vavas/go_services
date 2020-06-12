package extsrv

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/vavas/go_services/utils"
)

// BadRequest response
func BadRequest(err interface{}) *Response {
	var errs interface{}
	if err == nil {
		errs = []interface{}{http.StatusText(http.StatusBadRequest)}
	} else if kind := reflect.TypeOf(err).Kind(); kind == reflect.Array || kind == reflect.Slice {
		errs = err
	} else if errError, ok := err.(error); ok {
		errs = []interface{}{errError.Error()}
	} else {
		errs = []interface{}{err}
	}
	return &Response{StatusCode: http.StatusBadRequest, Body: utils.M{"errors": errs}}
}

// NotFound response
func NotFound(messages ...interface{}) *Response {
	msg := errMessage(http.StatusNotFound, messages...)
	return &Response{StatusCode: http.StatusNotFound, Body: utils.M{"errors": []string{msg}}}
}

// Forbidden response
func Forbidden(messages ...interface{}) *Response {
	msg := errMessage(http.StatusForbidden, messages...)
	return &Response{StatusCode: http.StatusForbidden, Body: utils.M{"errors": []string{msg}}}
}

// Unauthorized response
func Unauthorized(messages ...interface{}) *Response {
	msg := errMessage(http.StatusUnauthorized, messages...)
	return &Response{StatusCode: http.StatusUnauthorized, Body: utils.M{"errors": []string{msg}}}
}

// ServerError response
func ServerError(messages ...interface{}) *Response {
	msg := errMessage(http.StatusInternalServerError, messages...)
	return &Response{StatusCode: http.StatusInternalServerError, Body: utils.M{"errors": []string{msg}}}
}

// Success response
func Success(body interface{}) *Response {
	return &Response{StatusCode: http.StatusOK, Body: body}
}

// Created response
func Created(body interface{}) *Response {
	return &Response{StatusCode: http.StatusCreated, Body: body}
}

// NoContent response
func NoContent() *Response {
	return &Response{StatusCode: http.StatusNoContent}
}

// NotModified response
func NotModified() *Response {
	return &Response{StatusCode: http.StatusNotModified}
}

// PaymentRequired response
func PaymentRequired(messages ...interface{}) *Response {
	msg := errMessage(http.StatusPaymentRequired, messages...)
	return &Response{StatusCode: http.StatusPaymentRequired, Body: utils.M{"errors": []string{msg}}}
}

// UnavailableForLegalReasons response
func UnavailableForLegalReasons(messages ...interface{}) *Response {
	msg := errMessage(http.StatusUnavailableForLegalReasons, messages...)
	return &Response{StatusCode: http.StatusUnavailableForLegalReasons, Body: utils.M{"errors": []string{msg}}}
}

// Redirect response
func Redirect(redirectURL *url.URL) *Response {
	headers := map[string]string{"Location": redirectURL.String()}
	return &Response{StatusCode: http.StatusFound, Headers: headers}
}

// RespError responses with error
func RespError(err error, rawData ...map[string]interface{}) *Response {
	errorMsg := strings.ToLower(err.Error())

	utils.NotifyError(err, rawData...)

	if strings.Contains(errorMsg, "not found") {
		return NotFound()
	}

	if strings.Contains(errorMsg, "duplicate key error") {
		return BadRequest("Duplicate Key Error")
	}

	if strings.Contains(errorMsg, "request canceled") || strings.Contains(errorMsg, "context canceled") {
		return BadRequest("Request Canceled")
	}

	return ServerError()
}

func errMessage(status int, messages ...interface{}) (msg string) {
	count := len(messages)
	if count == 0 {
		msg = http.StatusText(status)
	} else if count == 1 {
		msg = fmt.Sprintf("%+v", messages[0])
	} else if format, ok := messages[0].(string); ok {
		msg = fmt.Sprintf(format, messages[1:]...)
	}
	return msg
}
