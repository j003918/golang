// loadbalance project loadbalance.go
package loadbalance

import (
	"net/http"
	"net/http/httputil"
	"sync"
)

const (
	ScheduleRand = iota
	ScheduleOrder
)

type node struct {
	activeTime int64 //time.Now().Unix()
	proxy      *httputil.ReverseProxy
	checkHB    bool
	valid      bool
}

type LB struct {
	index  int
	nodes  []*node
	rwLock *sync.RWMutex
}

const maxNodeLen = 10

func NewLB() *LB {
	return &LB{
		nodes:  make([]*node, 0, maxNodeLen),
		rwLock: new(sync.RWMutex),
	}
}

func (this *LB) Register(scheme, host string) {
	this.rwLock.Lock()
	director := func(req *http.Request) {
		req.URL.Scheme = scheme
		req.URL.Host = host
	}

	this.nodes = append(this.nodes, &node{
		proxy: &httputil.ReverseProxy{Director: director},
	})
	this.rwLock.Unlock()
}

func (this *LB) getNode() *node {
	this.rwLock.Lock()
	nl := len(this.nodes)
	if nl <= 0 {
		this.rwLock.Unlock()
		return nil
	} else {
		index := this.index % nl
		this.index = index + 1
		this.rwLock.Unlock()
		return this.nodes[index]
	}
}

//Proxy work with: http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { lb.Proxy(w, r) })
func (this *LB) Proxy(w http.ResponseWriter, r *http.Request) {
	n := this.getNode()
	if n == nil {
		w.WriteHeader(404)
	} else {
		n.proxy.ServeHTTP(w, r)
	}
}
