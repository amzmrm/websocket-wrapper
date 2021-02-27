package ws

import (
	"log"
	"runtime"
)

// Recovery recovers panics and prevents service down
type Recovery struct{}

func (r *Recovery) ServeWs(client *Client, param *Request, next HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			log.Printf("panic: %s", string(stack))
		}
	}()

	next(client, param)
}

// NewRecovery ...
func NewRecovery() *Recovery {
	return &Recovery{}
}
