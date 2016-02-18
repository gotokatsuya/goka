package goka

type Group struct {
	goka Goka
}

func (g *Group) Use(m ...Middleware) {
	for _, h := range m {
		g.goka.middleware = append(g.goka.middleware, wrapMiddleware(h))
	}
}

func (g *Group) Delete(path string, h Handler) {
	g.goka.Delete(path, h)
}

func (g *Group) Get(path string, h Handler) {
	g.goka.Get(path, h)
}

func (g *Group) Head(path string, h Handler) {
	g.goka.Head(path, h)
}

func (g *Group) Options(path string, h Handler) {
	g.goka.Options(path, h)
}

func (g *Group) Patch(path string, h Handler) {
	g.goka.Patch(path, h)
}

func (g *Group) Post(path string, h Handler) {
	g.goka.Post(path, h)
}

func (g *Group) Put(path string, h Handler) {
	g.goka.Put(path, h)
}

func (g *Group) Any(path string, h Handler) {
	for _, m := range methods {
		g.goka.add(m, path, h)
	}
}

func (g *Group) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		g.goka.add(m, path, h)
	}
}

func (g *Group) ServeFile(path, file string) {
	g.goka.ServeFile(path, file)
}

func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.goka.Group(prefix, m...)
}
