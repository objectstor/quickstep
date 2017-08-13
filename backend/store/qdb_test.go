package qdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDbOpen(t *testing.T) {
	db := new(Qdb)
	_, err := db.Open()
	assert.Error(t, err)
	db.Type = "Blah"
	_, err = db.Open()
	assert.Error(t, err)
	db.Url = "localhostBadName"
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
	db.Url = "localhost"
	session, err := db.Open()
	assert.Nil(t, err)
	defer session.Close()
	new_session := session.New()
	assert.NotNil(t, new_session)
	defer new_session.Close()
	err = db.Close()
	assert.Nil(t, err)

}
