package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"quickstep/backend/store"

	"gopkg.in/mgo.v2/bson"
)

//TODO add etag checking
// TODO proces in bqserver which search and delete task witout owner
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
	var task qdb.Task
	var acl qdb.ACLPerm
	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	userID := GetIDFromContext(r)
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
	aclString := r.Header.Get("X-Task-ACL")
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&task)
	if err != nil {
		JSONError(w, "Content error", http.StatusBadRequest)
		return
	}
	err = ParseRestTask(&task)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return

	}
	if len(aclString) > 0 {
		//json.
		err := json.Unmarshal([]byte(aclString), &acl)
		if err != nil {
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(acl.User) == 0 {
			JSONError(w, "missing name", http.StatusBadRequest)
			return
		}
	} else {
		acl = *qdb.CreateACL(contextUser, contextOrg, "crud")
	}
	// always create task.ID  as this is put even when task name are the same
	// we can have 2 tasks with the same name
	task.ID = bson.NewObjectId()
	//TODO we should check parrentID abd check is if thet exists
	//TODO store in db
	// create UserTask entry
	//fmt.Fprintf(w, "%s %s %s acl:%v", contextUser, contextOrg, userID, acl)
	taskID, err := session.InsertTask(&task)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if task.ID.Hex() != taskID {
		JSONError(w, "task id error", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"taskid\": %q}", task.ID.Hex())
	return
}
