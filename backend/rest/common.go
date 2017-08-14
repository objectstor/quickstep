package qrouter

/* common function for rest
 */
import (
	"fmt"
	"net/http"
	"quickstep/backend/store"
)

func JsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{error: %q}", message)
}

/*GetDbSession retrive db session from context */
func GetDbSession(r *http.Request) *qdb.QSession {
	session := r.Context().Value("dbsession")
	if session != nil {
		return session.(*qdb.QSession)
	}
	return nil
}
