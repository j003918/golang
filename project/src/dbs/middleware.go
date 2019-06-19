// middleware
package main

import "net/http"

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Chain struct {
	middlewares []Middleware
}

func NewChain(ms ...Middleware) Chain {
	return Chain{ms}
}

// func (p Pipeline) Pipe(ms ...Middleware) Pipeline {
// 	return Pipeline{append(p.middlewares, ms...)}
// }

func (p Chain) ThenFunc(h http.HandlerFunc) http.HandlerFunc {
	for i := range p.middlewares {
		h = p.middlewares[len(p.middlewares)-1-i](h)
	}
	return h
}
