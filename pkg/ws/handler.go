package ws

// A Handler responds to an WebSocket request.
type Handler interface {
	ServeWs(client *Client, param *Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as WebSocket handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(client *Client, param *Request)

// ServeHTTP calls w(client, param).
func (w HandlerFunc) ServeWs(client *Client, param *Request) {
	w(client, param)
}

// Handle binds a RequestCode to a Handler
func Handle(request RequestCode, handler Handler) {
	DefaultWsServerMux.m[request] = handler
}

// HandleFunc binds a RequestCode to a HandlerFunc
func HandleFunc(request RequestCode, handler func(client *Client, param *Request)) {
	Handle(request, HandlerFunc(handler))
}

// ServerMux maps all the RequestCodes to Handlers
type ServerMux struct {
	m map[RequestCode]Handler
}

func NewWsRouter() *ServerMux {
	return &ServerMux{
		m: make(map[RequestCode]Handler),
	}
}

func (mux *ServerMux) Handle(request RequestCode, handler Handler) {
	mux.m[request] = handler
}

func (mux *ServerMux) HandleFunc(request RequestCode, handler func(client *Client, param *Request)) {
	mux.m[request] = HandlerFunc(handler)
}

func (mux *ServerMux) ServeWs(client *Client, param *Request) {
	h, ok := mux.m[param.Code]
	if !ok {
		resp := &Response{
			Error: false,
			Data:  nil,
		}
		client.SendResponse(resp)
		return
	}
	h.ServeWs(client, param)
}

var DefaultWsServerMux = &ServerMux{
	m: make(map[RequestCode]Handler),
}

type MiddlewareHandler interface {
	ServeWs(client *Client, param *Request, next HandlerFunc)
}

type MiddlewareHandlerFunc func(client *Client, param *Request, next HandlerFunc)

func (m MiddlewareHandlerFunc) ServeWs(client *Client, param *Request, next HandlerFunc) {
	m(client, param, next)
}

type middleware struct {
	handler MiddlewareHandler
	next    *middleware
}

func (m middleware) ServeWs(client *Client, param *Request) {
	m.handler.ServeWs(client, param, m.next.ServeWs)
}

type Tequila struct {
	middleware middleware
	handlers   []MiddlewareHandler
}

func (t *Tequila) ServeWs(client *Client, param *Request) {
	t.middleware.ServeWs(client, param)
}

func NewTequila(handlers ...MiddlewareHandler) *Tequila {
	t := Tequila{
		middleware: build(handlers),
		handlers:   handlers,
	}

	hub = newHub(&t)
	go hub.run()

	return &t
}

func (t *Tequila) Use(handler MiddlewareHandler) {
	if handler == nil {
		panic("middleware handler can not be nil")
	}

	t.handlers = append(t.handlers, handler)
	t.middleware = build(t.handlers)
}

func (t *Tequila) UseFunc(handlerFunc func(client *Client, param *Request, next HandlerFunc)) {
	t.Use(MiddlewareHandlerFunc(handlerFunc))
}

func (t *Tequila) UseWsHandler(handler Handler) {
	t.Use(Wrap(handler))
}

func Wrap(handler Handler) MiddlewareHandler {
	return MiddlewareHandlerFunc(func(client *Client, param *Request, next HandlerFunc) {
		handler.ServeWs(client, param)
		next(client, param)
	})
}

func WrapFunc(handlerFunc HandlerFunc) MiddlewareHandler {
	return MiddlewareHandlerFunc(func(client *Client, param *Request, next HandlerFunc) {
		handlerFunc(client, param)
		next(client, param)
	})
}

func build(handlers []MiddlewareHandler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = build(handlers[1:])
	} else {
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}

func voidMiddleware() middleware {
	return middleware{
		handler: MiddlewareHandlerFunc(func(client *Client, param *Request, next HandlerFunc) {}),
		next:    &middleware{},
	}
}
