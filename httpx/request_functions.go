package httpx

import (
	"net"
	"net/http"
	"sync"

	"github.com/smartystreets/pipeline/numeric"
)

func ClientIPAddress(request *http.Request) string {
	if origin := request.Header.Get("X-Forwarded-For"); len(origin) > 0 {
		return origin
	} else if address, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		return address
	} else {
		return request.RemoteAddr
	}
}

func ExtractUint64Header(request *http.Request, name string) uint64 {
	return numeric.StringToUint64(request.Header.Get(name))
}

func NewWaitGroup(workers int) *sync.WaitGroup {
	waiter := &sync.WaitGroup{}
	waiter.Add(workers)
	return waiter
}
