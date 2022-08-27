package transport

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"seamless-api-wrapper/internal/rpc"
	"strings"
	"time"
)

type HttpServer struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewHttpTransport(addr string, readTimeout, writeTimeout time.Duration) *HttpServer {
	return &HttpServer{
		addr:         addr,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (s *HttpServer) Run(ctx context.Context, resolver rpc.Resolver) error {
	srv := http.Server{
		Addr: s.addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, err := validate(r)
			if err != nil {
				w.WriteHeader(status)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			resolver.Resolve(ctx, w, r.Body)
			r.Body.Close()
		}),
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error(err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func validate(r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, errors.New("rpc: POST method required, received " + r.Method)
	}

	contentType := r.Header.Get("Content-Type")

	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	if strings.ToLower(contentType) != "application/json" {
		return http.StatusUnsupportedMediaType, errors.New("rpc: unrecognized content-type: " + contentType)
	}

	return http.StatusOK, nil
}
