package qrouter

/*
all rest calls
*/
import (
	"fmt"
	"net/http"
)

/*
func headTasks(w http.ResponseWriter, r *http.Request) {
	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	if !ValidUserAndSession(session, contextUserString, w) {
		return
	}
	contextUser, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%s %s", contextUser, contextOrg)
}

func postTask(w http.ResponseWriter, r *http.Request) {
	taskID := pat.Param(r, "id")
	taskID, err := GetParamFromRequest(r, "id", "/task")
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}

	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	if !ValidUserAndSession(session, contextUserString, w) {
		return
	}
	contextUser, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "%s %s %s", contextUser, contextOrg, taskID)

}
*/
func putTask(w http.ResponseWriter, r *http.Request) {
	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	userID := GetIdFromContext(r)
	if !ValidUserAndSession(session, contextUserString, w) {
		return
	}
	if len(userID) == 0 {
		JSONError(w, "Context error ", http.StatusBadRequest)
	}
	// must have header with oner ACL otherwise
	// crud for current owner
	contextUser, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	acl := r.Header.Get("X-Task-ACL")
	// check for acl
	fmt.Fprintf(w, "%s %s %s acl:%s", contextUser, contextOrg, userID, acl)

}
