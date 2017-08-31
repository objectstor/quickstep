package qrouter

//TODO ! REDUCE number of quersied in PUT especially for
// customized ACL
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
func getTasksForUser(w http.ResponseWriter, r *http.Request) {
	ctx, err := NewQContext(r, true)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	session := ctx.DBSession()
	defer session.Close()
	// if header have data pick date if not pick all
	result, err := session.FindUserTasks(ctx.UserID(), "")
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	buff, err := json.Marshal(result)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buff)
}

/*
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
	ctx, err := NewQContext(r, true)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	session := ctx.DBSession()
	defer session.Close()
	// must have header with oner ACL otherwise
	// crud for current owner
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
	UserTaskID := ""
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
		nUser, dberr := session.FindUser(acl.User, acl.Domain)
		if dberr != nil {
			JSONError(w, "user not found", http.StatusBadRequest)
			return
		}
		UserTaskID = nUser.ID.Hex()
	} else {
		acl = *qdb.CreateACL(ctx.User(), ctx.Org(), "crud")
		UserTaskID = ctx.UserID()
	}
	// always create task.ID  as this is put even when task name are the same
	// we can have 2 tasks with the same name
	task.ID = bson.NewObjectId()
	// check user
	if bson.IsObjectIdHex(task.ParentID) {
		_, derr := session.FindTask(task.ParentID)
		if derr != nil {
			JSONError(w, derr.Error(), http.StatusBadRequest)
			return
		}
	}

	for _, childIDString := range task.ChildID {
		if !bson.IsObjectIdHex(childIDString) {
			JSONError(w, "bad ID", http.StatusBadRequest)
			return
		}
		_, derr := session.FindTask(childIDString)
		if derr != nil {
			JSONError(w, derr.Error(), http.StatusBadRequest)
			return
		}
	}

	taskID, err := session.InsertTask(&task)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if task.ID.Hex() != taskID {
		session.DeleteTask(task.ID.Hex())
		JSONError(w, "task id error", http.StatusBadRequest)
		return
	}
	// store user task
	uTask := new(qdb.UserTask)
	uTask.UserID = UserTaskID
	uTask.TaskID = taskID
	uTask.TaskName = task.Name
	uTask.CreationTime = task.CreationTime
	uTask.DeadLineTime = task.DeadLineTime
	err = session.InsertUserTask(uTask)
	if err != nil {
		session.DeleteTask(uTask.TaskID)
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
		// TODO delete task
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"taskid\": %q}", task.ID.Hex())
	return
}
