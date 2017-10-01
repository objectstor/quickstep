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
	"quickstep/backend/stats"
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
	stats := ctx.Statistics()
	stats.Inc(qstats.TaskGetCount)
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

func postTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := GetParamFromRequest(r, "id", "/task")
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	ctx, err := NewQContext(r, true)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	stats := ctx.Statistics()
	stats.Inc(qstats.TaskPostCount)
	session := ctx.DBSession()
	defer session.Close()
	// we need to check if ctx user can modify this Task

	currentTaskTable, err := session.FindUserTasks(ctx.UserID(), taskID)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(currentTaskTable) != 1 {
		msg := fmt.Sprintf("Bad number of tasks %d (%s,%s)", len(currentTaskTable), ctx.User(), taskID)
		JSONError(w, msg, http.StatusBadRequest)
		return
	}
	// check ACl for task
	//qdb.CheckACL(creator, acl.User, acl.Domain, "c")
	if ctx.User() != currentTaskTable[0].ACL.User {
		msg := fmt.Sprintf("Bad ACL for task %s", taskID)
		JSONError(w, msg, http.StatusBadRequest)
		return
	}
	if ctx.Org() != currentTaskTable[0].ACL.Domain {
		msg := fmt.Sprintf("Bad ACL for task %s", taskID)
		JSONError(w, msg, http.StatusBadRequest)
		return
	}
	if !currentTaskTable[0].ACL.Update {
		msg := fmt.Sprintf("Bad Access for task %s", taskID)
		JSONError(w, msg, http.StatusBadRequest)
		return
	}

	fmt.Println(currentTaskTable[0].ACL.Update)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"taskid\": %q}", taskID)
}

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
	stats := ctx.Statistics()
	stats.Inc(qstats.TaskPutCount)
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
		creator, dberr := session.FindUser(ctx.User(), ctx.Org())
		if dberr != nil {
			JSONError(w, dberr.Error(), http.StatusBadRequest)
			return
		}
		if !qdb.CheckACL(creator, acl.User, acl.Domain, "c") {
			JSONError(w, "bad access permissions", http.StatusBadRequest)
			return
		}
		// we need to check if UserCTX have access to user acl
		UserTaskID = nUser.ID.Hex()
	} else {
		// have full permission on my object
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
	uTask.ACL = acl
	err = session.InsertUserTask(uTask)
	if err != nil {
		session.DeleteTask(uTask.TaskID)
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// if len(task.Action) > 0 {
	// process action
	// jobID :=  queue.submitAction(task)
	//  w.Header().Set("X-Action-ID", jobID)
	//}
	//
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"taskid\": %q}", task.ID.Hex())
	return
}
