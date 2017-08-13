package qrouter

/* common function for rest
 */
import (
	"fmt"
	"net/http"
)

func JsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{error: %q}", message)
}
