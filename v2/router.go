package v2

// RequestImpl 请求实现
type RequestImpl struct {
	conn Connection
	msg  Message
}

func (r *RequestImpl) GetConnection() Connection {
	return r.conn
}

func (r *RequestImpl) GetMessage() Message {
	return r.msg
}

// Router 路由
type Router struct {
	handlers map[uint32]RouterHandler
}

// NewRouter 创建路由
func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint32]RouterHandler),
	}
}

// AddRouter 添加路由
func (r *Router) AddRouter(msgID uint32, handler RouterHandler) {
	r.handlers[msgID] = handler
}

// Handle 处理请求
func (r *Router) Handle(req Request) {
	handler, ok := r.handlers[req.GetMessage().GetMsgID()]
	if !ok {
		return
	}

	handler.PreHandle(req)
	handler.Handle(req)
	handler.PostHandle(req)
}

// BaseHandler 基础处理器
type BaseHandler struct{}

func (b *BaseHandler) PreHandle(request Request)  {}
func (b *BaseHandler) Handle(request Request)     {}
func (b *BaseHandler) PostHandle(request Request) {}
