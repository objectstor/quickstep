package qrouter

/* middlewares for rest */

import (
	"fmt"
	"net/http"
)

func logging(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("<LOGGING> Received request: %v\n", r.URL)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
