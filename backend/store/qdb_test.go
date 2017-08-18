package qdb

import (
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
	acl := CreateACL("org", "crud")
	user := new(User)
	user.ID = bson.NewObjectId()
	user.Name = "super"
	user.Password = "password"
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
	err = session.InsertUser(newUser)
	assert.Error(t, err) // cannot change ID on existing document
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
