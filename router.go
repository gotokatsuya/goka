package goka

type (
	Router struct {
		tree   *node
		routes []Route
		goka   *Goka
	}
	node struct {
		kind          kind
		label         byte
		prefix        string
		parent        *node
		children      children
		ppath         string
		pnames        []string
		methodHandler *methodHandler
		goka          *Goka
	}
	kind          uint8
	children      []*node
	methodHandler struct {
		connect HandlerFunc
		delete  HandlerFunc
		get     HandlerFunc
		head    HandlerFunc
		options HandlerFunc
		patch   HandlerFunc
		post    HandlerFunc
		put     HandlerFunc
		trace   HandlerFunc
	}
)

const (
	skind kind = iota
	pkind
	mkind
)

func NewRouter(g *Goka) *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes: []Route{},
		goka:   g,
	}
}

func (r *Router) Add(method, path string, h HandlerFunc, g *Goka) {
	ppath := path
	pnames := []string{}

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil, g)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames, g)
				return
			}
			r.insert(method, path[:i], nil, pkind, ppath, pnames, g)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil, g)
			pnames = append(pnames, "_*")
			r.insert(method, path[:i+1], h, mkind, ppath, pnames, g)
			return
		}
	}

	r.insert(method, path, h, skind, ppath, pnames, g)
}

func (r *Router) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string, g *Goka) {
	l := len(pnames)
	if *g.maxParam < l {
		*g.maxParam = l
	}

	cn := r.tree
	if cn == nil {
		panic("goka => invalid method")
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		switch {
		case l == 0:
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
				cn.goka = g
			}
		case l < pl:
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children, cn.methodHandler, cn.ppath, cn.pnames, cn.goka)

			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil
			cn.goka = nil

			cn.addChild(n)

			if l == sl {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
				cn.goka = g
			} else {
				n = newNode(t, search[l:], cn, nil, new(methodHandler), ppath, pnames, g)
				n.addHandler(method, h)
				cn.addChild(n)
			}
		case l < sl:
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				cn = c
				continue
			}
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames, g)
			n.addHandler(method, h)
			cn.addChild(n)
		default:
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = path
				cn.pnames = pnames
				cn.goka = g
			}
		}

		return
	}
}

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, ppath string, pnames []string, g *Goka) *node {
	return &node{
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		ppath:         ppath,
		pnames:        pnames,
		methodHandler: mh,
		goka:          g,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case GET:
		n.methodHandler.get = h
	case POST:
		n.methodHandler.post = h
	case PUT:
		n.methodHandler.put = h
	case DELETE:
		n.methodHandler.delete = h
	case PATCH:
		n.methodHandler.patch = h
	case OPTIONS:
		n.methodHandler.options = h
	case HEAD:
		n.methodHandler.head = h
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	case GET:
		return n.methodHandler.get
	case POST:
		return n.methodHandler.post
	case PUT:
		return n.methodHandler.put
	case DELETE:
		return n.methodHandler.delete
	case PATCH:
		return n.methodHandler.patch
	case OPTIONS:
		return n.methodHandler.options
	case HEAD:
		return n.methodHandler.head
	default:
		return nil
	}
}

func (n *node) check405() HandlerFunc {
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			return methodNotAllowedHandler
		}
	}
	return notFoundHandler
}

func (r *Router) Find(method, path string, ctx *Context) (h HandlerFunc, g *Goka) {
	h = notFoundHandler
	g = r.goka
	cn := r.tree

	var (
		search = path
		c      *node
		n      int
		nk     kind
		nn     *node
		ns     string
	)

	for {
		if search == "" {
			goto End
		}

		pl := 0
		l := 0

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			search = search[l:]
		} else {
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == mkind {
				goto MatchAny
			} else {
				return
			}
		}

		if search == "" {
			goto End
		}

		c = cn.findChild(search[0], skind)
		if c != nil {
			if cn.label == '/' {
				nk = pkind
				nn = cn
				ns = search
			}
			cn = c
			continue
		}

	Param:
		c = cn.findChildByKind(pkind)
		if c != nil {
			if cn.label == '/' {
				nk = mkind
				nn = cn
				ns = search
			}
			cn = c
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			ctx.values[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

	MatchAny:
		if cn = cn.findChildByKind(mkind); cn == nil {
			return
		}
		ctx.values[len(cn.pnames)-1] = search
		goto End
	}

End:
	ctx.path = cn.ppath
	ctx.names = cn.pnames
	h = cn.findHandler(method)
	if cn.goka != nil {
		g = cn.goka
	}

	if h == nil {
		h = cn.check405()

		if cn = cn.findChildByKind(mkind); cn == nil {
			return
		}
		ctx.values[len(cn.pnames)-1] = ""
		if h = cn.findHandler(method); h == nil {
			h = cn.check405()
		}
	}
	return
}
