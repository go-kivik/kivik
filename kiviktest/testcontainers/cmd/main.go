// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js

// Package main provides a simple HTTP server to start CouchDB using testcontainers.
// It is meant to be used with GopherJS, or older versions of Go, that cannot
// use modern testcontainers directly.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	tc "github.com/go-kivik/kivik/v4/kiviktest/testcontainers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Set a timeout for the context to avoid hanging indefinitely
	ctx, timeoutCancel := context.WithTimeout(ctx, 20*time.Minute)
	defer timeoutCancel()

	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening on %s\n", l.Addr())
	s := http.Server{
		Handler: http.HandlerFunc(startCouchDB(ctx)),
	}
	go func() {
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	fmt.Println("Shutting down server...")
	if err := s.Shutdown(shutdownCtx); err != nil {
		panic(err)
	}
}

func startCouchDB(ctx context.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		image := r.URL.Query().Get("image")
		if image == "" {
			http.Error(w, "image query parameter is required", http.StatusBadRequest)
			return
		}
		dsn, err := tc.StartCouchDB(ctx, image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(dsn))
		_, _ = w.Write([]byte("\n"))
	}
}
