package qrouter

/* common function for rest
 */
import (
	"errors"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"strings"

	"goji.io/pat"
)

//JSONError send Json error code
func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{error: %q}", message)
}

//JSONOk send Json string
func JSONOk(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{message: %q}", message)
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

/*GetIdFromContext retrive user id from context */
func GetIdFromContext(r *http.Request) string {
	session := r.Context().Value("user_id")
	if session != nil {
		return session.(string)
	}
	return ""
}

func GetParamFromRequest(r *http.Request, name string, suffix string) (string, error) {
	path := r.RequestURI
	if strings.HasSuffix(path, suffix) {
		// we should have more than suffix
		return "", errors.New("bad path")
	}
	value := pat.Param(r, name)
	return value, nil
}

//ValidUserAndSession - validat database and user ( true ok)
func ValidUserAndSession(s *qdb.QSession, u string, w http.ResponseWriter) bool {
	if s == nil {
		JSONError(w, "database error", http.StatusNotAcceptable)
		return false
	}
	if len(u) == 0 {
		JSONError(w, "auth context error", http.StatusForbidden)
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
