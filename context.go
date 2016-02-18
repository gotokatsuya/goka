package goka

import (
	"bytes"
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/valyala/fasthttp"
)

type (
	Context struct {
		context.Context
		requestCtx *fasthttp.RequestCtx
		path       string
		names      []string
		values     []string
		query      *fasthttp.Args
		store      store
		goka       *Goka
	}
	store map[string]interface{}
)

func NewContext(reqCtx *fasthttp.RequestCtx, g *Goka) *Context {
	return &Context{
		requestCtx: reqCtx,
		goka:       g,
		values:     make([]string, *g.maxParam),
		store:      make(store),
	}
}

func (c *Context) RequestCtx() *fasthttp.RequestCtx {
	return c.requestCtx
}

func (c *Context) ParamNames() []string {
	return c.names
}

func (c *Context) Path() string {
	return c.path
}

func (c *Context) ParamByIndex(index int) (value string) {
	l := len(c.names)
	if index < l {
		value = c.values[index]
	}
	return
}

func (c *Context) ParamByName(name string) (value string) {
	l := len(c.names)
	for i, n := range c.names {
		if n == name && i < l {
			value = c.values[i]
			break
		}
	}
	return
}

func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.requestCtx.URI().QueryArgs()
	}
	return string(c.query.Peek(name))
}

func (c *Context) Form(name string) string {
	return string(c.requestCtx.FormValue(name))
}

func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

func (c *Context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = val
}

func (c *Context) Render(code int, name string, data interface{}) (err error) {
	if c.goka.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.goka.renderer.Render(buf, name, data); err != nil {
		return
	}
	c.requestCtx.SetContentType(TextHTMLCharsetUTF8)
	c.requestCtx.SetStatusCode(code)
	c.requestCtx.SetBody(buf.Bytes())
	return
}

func (c *Context) HTML(code int, html string) (err error) {
	c.requestCtx.SetContentType(TextHTMLCharsetUTF8)
	c.requestCtx.SetStatusCode(code)
	c.requestCtx.SetBodyString(html)
	return
}

func (c *Context) String(code int, s string) (err error) {
	c.requestCtx.SetContentType(TextPlainCharsetUTF8)
	c.requestCtx.SetStatusCode(code)
	c.requestCtx.SetBodyString(s)
	return
}

func (c *Context) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.requestCtx.SetContentType(ApplicationJSONCharsetUTF8)
	c.requestCtx.SetStatusCode(code)
	c.requestCtx.SetBody(b)
	return
}

func (c *Context) JSONIndent(code int, i interface{}, prefix string, indent string) (err error) {
	b, err := json.MarshalIndent(i, prefix, indent)
	if err != nil {
		return err
	}
	c.requestCtx.SetContentType(ApplicationJSONCharsetUTF8)
	c.requestCtx.SetStatusCode(code)
	c.requestCtx.SetBody(b)
	return
}

func (c *Context) NoContent(code int) error {
	c.requestCtx.SetStatusCode(code)
	return nil
}

func (c *Context) Redirect(code int, url string) error {
	if code < fasthttp.StatusMovedPermanently || code > fasthttp.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	c.requestCtx.Redirect(url, code)
	return nil
}

func (c *Context) Error(err error) {
	c.goka.httpErrorHandler(err, c)
}

func (c *Context) Goka() *Goka {
	return c.goka
}

func (c *Context) reset(rCtx *fasthttp.RequestCtx, g *Goka) {
	c.requestCtx = rCtx
	c.query = nil
	c.store = nil
	c.goka = g
}
