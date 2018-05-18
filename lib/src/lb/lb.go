// lb project lb.go
package lb

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"sync"
)

type node struct {
	activeTime int64 //time.Now().Unix()
	proxy      *httputil.ReverseProxy
	checkHB    bool
	valid      bool
}

type LoadBalance struct {
	nodes  []*node
	rwLock *sync.RWMutex
}

const maxNodeLen = 10

func NewLB() *LoadBalance {
	return &LoadBalance{
		nodes:  make([]*node, 0, maxNodeLen),
		rwLock: new(sync.RWMutex),
	}
}

func (this *LoadBalance) Register(scheme, host string, checkHB bool) {
	director := func(req *http.Request) {
		req.URL.Scheme = scheme
		req.URL.Host = host
	}

	this.nodes = append(this.nodes, &node{
		proxy: &httputil.ReverseProxy{Director: director},
		valid: true,
	})

}

func (this *LoadBalance) Proxy(w http.ResponseWriter, r *http.Request) {
	n := this.nodes[rand.Intn(len(this.nodes))]
	n.proxy.ServeHTTP(w, r)
}
