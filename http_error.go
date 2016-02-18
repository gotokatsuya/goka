package goka

import (
	"github.com/valyala/fasthttp"
)

func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{code: code, message: fasthttp.StatusMessage(code)}
	if len(msg) > 0 {
		m := msg[0]
		he.message = m
	}
	return he
}

func (e *HTTPError) SetCode(code int) {
	e.code = code
}

func (e *HTTPError) Code() int {
	return e.code
}

func (e *HTTPError) Error() string {
	return e.message
}
