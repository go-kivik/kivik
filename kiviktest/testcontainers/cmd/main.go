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
	"net/http"
	"os/signal"
	"syscall"

	tc "github.com/go-kivik/kivik/v4/kiviktest/testcontainers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	s := http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(startCouchDB),
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	<-ctx.Done()
	if err := s.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func startCouchDB(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "image query parameter is required", http.StatusBadRequest)
		return
	}
	dsn, err := tc.StartCouchDB(r.Context(), image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dsn))
	_, _ = w.Write([]byte("\n"))
}
