package reverseproxy

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ReverseProxy struct {
	listeners []Listener
	proxy     *httputil.ReverseProxy
	servers   []*http.Server
	targets   []*Target
}

type Target struct {
	router   *mux.Router
	upstream *url.URL
}

// AddTarget adds an upstream server to use for a request that matches
// a given gorilla/mux Router. These are matched via Director function.
func (r *ReverseProxy) AddTarget(upstream string, router *mux.Router) error {
	url, err := url.Parse(upstream)

	if err != nil {
		return err
	}

	if router == nil {
		router = mux.NewRouter()
		router.PathPrefix("/")
	}

	r.targets = append(r.targets, &Target{
		router:   router,
		upstream: url,
	})

	return nil
}

// AddListener adds a listener for non-TLS connections on the given address
func (r *ReverseProxy) AddListener(address string) {
	l := Listener{
		Addr: address,
	}

	r.listeners = append(r.listeners, l)
}

// AddListenerTLS adds a listener for TLS connections on the given address
func (r *ReverseProxy) AddListenerTLS(address, tlsCert, tlsKey string) {
	l := Listener{
		Addr:    address,
		TLSCert: tlsCert,
		TLSKey:  tlsKey,
	}

	r.listeners = append(r.listeners, l)
}

// Start will listen on configured listeners
func (r *ReverseProxy) Start() error {
	r.proxy = &httputil.ReverseProxy{
		Director: r.Director(),
	}

	// This breaks connections to non-http2 backends
	r.proxy.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			ta, err := net.ResolveTCPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			return net.DialTCP(network, nil, ta)
		},
	}

	for _, l := range r.listeners {
		listener, err := l.Make()
		if err != nil {
			// todo: Close any listeners that
			//       were created successfully
			return err
		}

		// This accepts h2c connections but doesn't seem to
		// pass that on through to the backend server until we over-ride the Transport in a way that allows h2c
		// However doing that breaks other "normal" connection types
		// Tested with: docker run --rm -it -p 8000:8000 -v $(pwd)/default.conf:/etc/nginx/conf.d/default.conf nginx:latest
		//                  ~/Code/Fideloper/pproxy-util/nginx-h2c/default.conf
		// Direct to docker: curl -v --http2-prior-knowledge http://localhost:8000
		// Through our proxy: curl -v --http2-prior-knowledge http://localhost
		// undo this: `go mod tidy` to remove golang.org/x/net/http2/h2c
		h2s := &http2.Server{}

		srv := &http.Server{Handler: h2c.NewHandler(r.proxy, h2s)}

		r.servers = append(r.servers, srv)

		// TODO: Handle unexpected errors from our servers
		if l.ServesTLS() {
			go func() {
				if err := srv.ServeTLS(listener, l.TLSCert, l.TLSKey); !errors.Is(err, http.ErrServerClosed) {
					log.Println(err)
				}
			}()
		} else {
			go func() {
				if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
					log.Println(err)
				}
			}()
		}
	}

	return nil
}

// Stop will gracefully shut down all listening servers
func (r *ReverseProxy) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup

	for _, srv := range r.servers {
		srv := srv
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := srv.Shutdown(ctx); err != nil {
				log.Println(err)
			}
			log.Println("A listener was shutdown successfully")
		}()
	}

	// Wait for all servers to shut down
	wg.Wait()
	log.Println("Server shut down")
}

// Director returns a function for use in http.ReverseProxy.Director.
// The function matches the incoming request to a specific target and
// sets the request object to be sent to the matched upstream server.
func (r *ReverseProxy) Director() func(req *http.Request) {
	return func(req *http.Request) {
		for _, t := range r.targets {
			match := &mux.RouteMatch{}
			if t.router.Match(req, match) {
				targetQuery := t.upstream.RawQuery

				req.URL.Scheme = t.upstream.Scheme
				req.URL.Host = t.upstream.Host
				req.URL.Path, req.URL.RawPath = joinURLPath(t.upstream, req.URL)
				if targetQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = targetQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
				}
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
				break
			}
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
