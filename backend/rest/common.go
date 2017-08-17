package qrouter

/* common function for rest
 */
import (
	"errors"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"strings"
)

func JsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{error: %q}", message)
}

/*GetDbSessionFromContext retrive db session from context */
func GetDbSessionFromContext(r *http.Request) *qdb.QSession {
	session := r.Context().Value("dbsession")
	if session != nil {
		return session.(*qdb.QSession)
	}
	return nil
}

/*GetUserFromContext retrive user from context */
func GetUserFromContext(r *http.Request) string {
	session := r.Context().Value("user")
	if session != nil {
		return session.(string)
	}
	return ""
}

//ValidUserAndSession - validat database and user ( true ok)
func ValidUserAndSession(s *qdb.QSession, u string, w http.ResponseWriter) bool {
	if s == nil {
		JsonError(w, "database error", http.StatusNotAcceptable)
		return false
	}
	if len(u) == 0 {
		JsonError(w, "auth context error", http.StatusForbidden)
		return false
	}
	return true
}

//GetUserAndOrg - get user and org parts from user
func GetUserAndOrg(u string) (string, string, error) {
	uo := strings.SplitN(u, "#", 2)
	if len(uo) != 2 {
		return "", "", errors.New("UserArg bad format")
	}
	if len(uo[1]) == 0 {
		uo[1] = "SYSTEM"
	}
	return uo[0], uo[1], nil
}
