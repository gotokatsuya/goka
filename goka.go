package goka

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"

	"github.com/valyala/fasthttp"
)

type (
	Goka struct {
		prefix                  string
		middleware              []MiddlewareFunc
		maxParam                *int
		defaultHTTPErrorHandler HTTPErrorHandler
		httpErrorHandler        HTTPErrorHandler
		renderer                Renderer
		pool                    sync.Pool
		debug                   bool
		router                  *Router
	}

	Route struct {
		Method  string
		Path    string
		Handler Handler
	}

	HTTPError struct {
		code    int
		message string
	}

	Middleware interface{}

	MiddlewareFunc func(HandlerFunc) HandlerFunc

	Handler interface{}

	HandlerFunc func(*Context) error

	HTTPErrorHandler func(error, *Context)

	Validator interface {
		Validate() error
	}

	Renderer interface {
		Render(w io.Writer, name string, data interface{}) error
	}
)

const (
	indexPage = "index.html"
)

var (
	methods = [...]string{
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
	}

	ErrUnsupportedMediaType  = NewHTTPError(fasthttp.StatusUnsupportedMediaType)
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")

	notFoundHandler = func(c *Context) error {
		return NewHTTPError(fasthttp.StatusNotFound)
	}

	methodNotAllowedHandler = func(c *Context) error {
		return NewHTTPError(fasthttp.StatusMethodNotAllowed)
	}
)

func New() (g *Goka) {
	g = &Goka{maxParam: new(int)}
	g.pool.New = func() interface{} {
		return NewContext(nil, g)
	}
	g.router = NewRouter(g)

	g.defaultHTTPErrorHandler = func(err error, c *Context) {
		code := fasthttp.StatusInternalServerError
		msg := fasthttp.StatusMessage(code)
		if he, ok := err.(*HTTPError); ok {
			code = he.code
			msg = he.message
		}
		if g.debug {
			msg = err.Error()
		}
		c.RequestCtx().Error(msg, code)
		return
	}
	g.SetHTTPErrorHandler(g.defaultHTTPErrorHandler)

	return
}

func (g *Goka) Router() *Router {
	return g.router
}

func (g *Goka) DefaultHTTPErrorHandler(err error, c *Context) {
	g.defaultHTTPErrorHandler(err, c)
}

func (g *Goka) SetHTTPErrorHandler(h HTTPErrorHandler) {
	g.httpErrorHandler = h
}

func (g *Goka) SetRenderer(r Renderer) {
	g.renderer = r
}

func (g *Goka) SetDebug(debug bool) {
	g.debug = debug
}

func (g *Goka) Debug() bool {
	return g.debug
}

func (g *Goka) Use(m ...Middleware) {
	for _, h := range m {
		g.middleware = append(g.middleware, wrapMiddleware(h))
	}
}

func (g *Goka) Delete(path string, h Handler) {
	g.add(DELETE, path, h)
}

func (g *Goka) Get(path string, h Handler) {
	g.add(GET, path, h)
}

func (g *Goka) Head(path string, h Handler) {
	g.add(HEAD, path, h)
}

func (g *Goka) Options(path string, h Handler) {
	g.add(OPTIONS, path, h)
}

func (g *Goka) Patch(path string, h Handler) {
	g.add(PATCH, path, h)
}

func (g *Goka) Post(path string, h Handler) {
	g.add(POST, path, h)
}

func (g *Goka) Put(path string, h Handler) {
	g.add(PUT, path, h)
}

func (g *Goka) Any(path string, h Handler) {
	for _, m := range methods {
		g.add(m, path, h)
	}
}

func (g *Goka) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		g.add(m, path, h)
	}
}

func (g *Goka) add(method, path string, h Handler) {
	path = g.prefix + path
	g.router.Add(method, path, wrapHandler(h), g)
	r := Route{
		Method:  method,
		Path:    path,
		Handler: runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
	}
	g.router.routes = append(g.router.routes, r)
}

func (g *Goka) Index(file string) {
	g.ServeFile("/", file)
}

func (g *Goka) Favicon(file string) {
	g.ServeFile("/favicon.ico", file)
}

func (g *Goka) ServeFile(path, file string) {
	g.Get(path, func(c *Context) {
		fasthttp.ServeFile(c.RequestCtx(), file)
	})
}

func (g *Goka) Group(prefix string, m ...Middleware) *Group {
	group := &Group{*g}
	group.goka.prefix += prefix
	if len(m) == 0 {
		mw := make([]MiddlewareFunc, len(group.goka.middleware))
		copy(mw, group.goka.middleware)
		group.goka.middleware = mw
	} else {
		group.goka.middleware = nil
		group.Use(m...)
	}
	return group
}

func (g *Goka) URI(h Handler, params ...interface{}) string {
	uri := new(bytes.Buffer)
	pl := len(params)
	n := 0
	hn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	for _, r := range g.router.routes {
		if r.Handler == hn {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < pl {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

func (g *Goka) URL(h Handler, params ...interface{}) string {
	return g.URI(h, params...)
}

func (g *Goka) Routes() []Route {
	return g.router.routes
}

func (g *Goka) Serve(rCtx *fasthttp.RequestCtx) {

	c := g.pool.Get().(*Context)
	h, g := g.router.Find(string(rCtx.Method()), string(rCtx.Path()), c)
	c.reset(rCtx, g)

	for i := len(g.middleware) - 1; i >= 0; i-- {
		h = g.middleware[i](h)
	}

	if err := h(c); err != nil {
		g.httpErrorHandler(err, c)
	}

	g.pool.Put(c)
}

func (g *Goka) Run(addr string) {
	fasthttp.ListenAndServe(addr, func(rCtx *fasthttp.RequestCtx) {
		g.Serve(rCtx)
	})
}

func (g *Goka) RunTLS(addr, certFile, keyFile string) {
	fasthttp.ListenAndServeTLS(addr, certFile, keyFile, func(rCtx *fasthttp.RequestCtx) {
		g.Serve(rCtx)
	})
}

func wrapMiddleware(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case MiddlewareFunc:
		return m
	case func(HandlerFunc) HandlerFunc:
		return m
	case HandlerFunc:
		return wrapHandlerFuncMW(m)
	case func(*Context) error:
		return wrapHandlerFuncMW(m)
	default:
		panic("unknown middleware")
	}
}

func wrapHandlerFuncMW(m HandlerFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			if err := m(c); err != nil {
				return err
			}
			return h(c)
		}
	}
}

func wrapHandler(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(*Context) error:
		return h
	case func(*fasthttp.RequestCtx):
		return func(c *Context) error {
			h(c.requestCtx)
			return nil
		}
	default:
		panic("unknown handler")
	}
}
