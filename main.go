package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"extprocserver/extproc"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":5441")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcServer, extproc.NewExtprocV3Server())

	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	headerMatcher := func(key string) (string, bool) {
		// Pass http headers through as metadata unchanged.
		// See https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#mapping-from-http-request-headers-to-grpc-client-metadata

		// http1 client can send connection header in the request, which will result in error when http1 is mapped to http/2
		// http/2 standard doesn't allow :connection header
		// skip connection header to make it compatible with http1 clients
		if key == "Connection" {
			return "", false
		}
		return key, true
	}
	mux := gwruntime.NewServeMux(gwruntime.WithIncomingHeaderMatcher(headerMatcher))
	err = setupReverseProxy(
		mux,
		fmt.Sprintf("http://127.0.0.1:%d", 5441),
		[]string{"/envoy.service.ext_proc.v3.ExternalProcessor/**"},
		true)

	h2cHandler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	}), &http2.Server{})

	httpServer := &http.Server{
		Handler: h2cHandler,
	}

	restLis, err := net.Listen("tcp", ":5442")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("h2c server listening on %s", restLis.Addr())
	err = httpServer.Serve(restLis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func setupReverseProxy(mux *gwruntime.ServeMux, target string, routes []string, useHttp2 bool) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return err
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	// set the error handler
	rp.ErrorHandler = rpErrorHandler()
	// useHttp2 is needed for grpc endpoints.
	if useHttp2 {
		rp.Transport = &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				ta, err := net.ResolveTCPAddr(network, addr)
				if err != nil {
					return nil, err
				}
				return net.DialTCP(network, nil, ta)
			},
		}
	}

	for _, route := range routes {
		mux.HandlePath("POST", route, rpHandler(rp))
	}
	return nil
}

func rpErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, r *http.Request, err error) {
		if _, ok := err.(http2.StreamError); ok {
			if err.(http2.StreamError).Code == http2.ErrCodeCancel {
				rw.WriteHeader(http.StatusBadGateway)
				return
			}
		}

		if err != nil && err != context.Canceled {
			log.Fatalf("reverseproxy http error: %v\n", err)
		}
		rw.WriteHeader(http.StatusBadGateway)
	}
}

func rpHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request, map[string]string) {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		p.ServeHTTP(w, r)
	}
}
