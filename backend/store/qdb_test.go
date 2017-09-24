package qdb

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	//mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestDbOpen(t *testing.T) {
	db := new(Qdb)
	_, err := db.Open()
	assert.Error(t, err)
	db.Type = "Blah"
	_, err = db.Open()
	assert.Error(t, err)
	db.URL = "localhostBadName"
	// don't wait use short timeout
	db.Timeout = time.Second * 1
	_, err = db.Open()
	assert.Error(t, err)
	db.Type = "Mysql" // this should fail for now
	_, err = db.Open()
	assert.Error(t, err)
	db.Type = "mongodb"
	_, err = db.Open()
	assert.Error(t, err)
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	defer session.Close()
	newSession := session.New()
	assert.NotNil(t, newSession)
	defer newSession.Close()
	err = db.Close()
	assert.Nil(t, err)
}

func TestDBUserInsertFind(t *testing.T) {
	db := new(Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	defer db.Close()
	acl := CreateACL("", "org", "crud")
	user := new(User)
	user.ID = bson.NewObjectId()
	user.Name = "super"
	user.Secret = "secret"
	user.ACL = append(user.ACL, *acl)
	user.Org = "org"
	_, err = session.FindUser("super", "org")
	if err == nil {
		err = session.DeleteUser("super", "org")
		assert.Nil(t, err)
	}
	assert.Error(t, session.DeleteUser("", ""))
	err = session.InsertUser(user)
	assert.Nil(t, err)
	newUser, nerr := session.FindUser("super", "org")
	assert.NotNil(t, newUser)
	assert.Nil(t, nerr)
	assert.Equal(t, strings.Compare(user.Name, newUser.Name), 0)
	_, dberr := session.FindUser("super_one", "org")
	assert.Equal(t, EntryNotFound(dberr), true)
	defer session.Close()
	_, dberr = session.FindUser("", "org")
	assert.Error(t, dberr)
	emptyuser := new(User)
	err = session.InsertUser(emptyuser)
	assert.Error(t, err)
	err = session.DeleteUser("not_exists", "org")
	assert.Error(t, err)

	newUser, nerr = session.FindUser("super", "org")
	assert.NotNil(t, newUser)
	assert.Nil(t, nerr)
	newUser.Name = "new_name" // ID stays so it should failed
	err = session.InsertUser(newUser)
	assert.Error(t, err) // duplicate key - same is is used
	newUser.ID = bson.NewObjectId()
	_, nerr = session.FindUser("new_name", "org")
	err = session.InsertUser(newUser)
	if nerr != nil {
		assert.Nil(t, err)
	} else {
		assert.Error(t, err) // cannot change ID on existing document
	}
	newUser.ID = "BAD ID"
	err = session.InsertUser(newUser) // same story as above
	assert.Error(t, err)

	if session != nil {
		session.mgoSession = nil
		err = session.InsertUser(emptyuser)
		assert.Error(t, err)
	}
	err = session.DeleteUser("not_exists", "org")
	assert.Error(t, err)
}

func TestCreateTask(t *testing.T) {
	db := new(Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	task := new(Task)
	task.Private = false
	task.Status = "NEW"
	myTime := time.Now()
	task.CreationTime = myTime
	task.ModificationTime = myTime
	task.DeadLineTime = time.Now().Add(time.Hour * 24 * 7)
	task.Name = ""
	_, err = session.InsertTask(task)
	assert.Error(t, err)
	task.Name = "new_task"
    task.ItemType = TaskItem
	taskID, err := session.InsertTask(task)
	assert.Nil(t, err)
	assert.True(t, bson.IsObjectIdHex(taskID))

	sameTask, err := session.FindTask(taskID)
	assert.Nil(t, err)
	assert.Equal(t, taskID, sameTask.ID.Hex())
	notFoundID := bson.NewObjectId()
	notFoundTask, err := session.FindTask(notFoundID.Hex())
	assert.Error(t, err)
	assert.Nil(t, notFoundTask)
	defer db.Close()
}

func TestCreateUserTask(t *testing.T) {
	db := new(Qdb)
	db.Type = "mongodb"
	db.Timeout = time.Second * 10
	db.URL = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	uTask := new(UserTask)
	uTask.TaskID = "nothing"
	err = session.InsertUserTask(uTask)
	assert.Error(t, err)

	user, err := session.FindUser("super", "org")
	assert.Nil(t, err)

	// create new task
	task := new(Task)
	task.Private = false
	task.Status = "NEW"
    task.ItemType = TaskItem
	myTime := time.Now()
	task.CreationTime = myTime
	task.ModificationTime = myTime
	task.DeadLineTime = time.Now().Add(time.Hour * 24 * 7)

	task.Name = fmt.Sprintf("new_and_sweet_%s", time.Now().Format("20060102150405"))
	taskID, err := session.InsertTask(task)
	assert.Nil(t, err)

	ntask := new(Task)
	ntask = task
    ntask.ItemType = TaskItem
	ntask.Name = "C_new_name"
	ntask.ID = ""
	ntaskID, err := session.InsertTask(task)
	assert.Nil(t, err)

	//complete task
	uTask.TaskID = taskID
	err = session.InsertUserTask(uTask)
	assert.Error(t, err)
	uTask.UserID = user.ID.Hex()
	uTask.CreationTime = task.CreationTime
	uTask.DeadLineTime = task.DeadLineTime

	err = session.InsertUserTask(uTask)
	assert.Nil(t, err)

	nuTask := new(UserTask)
	nuTask = uTask
	nuTask.TaskID = ntaskID
	nuTask.CreationTime = ntask.CreationTime
	nuTask.DeadLineTime = ntask.DeadLineTime
	err = session.InsertUserTask(uTask)
	assert.Nil(t, err)

	ret, dberr := session.FindUserTasks(uTask.UserID, uTask.TaskID)
	assert.Nil(t, dberr)
	assert.Equal(t, 1, len(ret))

	ret, dberr = session.FindUserTasks(uTask.UserID, "")
	assert.Nil(t, dberr)
	assert.True(t, len(ret) > 1)

	err = session.DeleteTask(uTask.TaskID)
	assert.Nil(t, err)
	err = session.DeleteTask("AAA")
	assert.Error(t, err)

	session.Close()
	defer db.Close()
	err = session.DeleteTask(uTask.TaskID)
	assert.Error(t, err)

}
