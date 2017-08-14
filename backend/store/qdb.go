package qdb

// DATABASE abstraction
/* TODO: go routine for monitoring
use user login and database instead pure Session
all mongodb admin crap
probably run indexing in thread will make more sense check it
*/

import (
	"errors"
	"fmt"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
)

/*Qdb - dtabase abstraction class */
type Qdb struct {
	Type    string `yaml: "type"`
	URL     string `yaml: "url"`
	Timeout time.Duration
}

/*QSession database session abtraction struct */
type QSession struct {
	mgoSession *mgo.Session
	SigningKey []byte
}

/*Close close session */
func (s *QSession) Close() {
	if s.mgoSession != nil {
		s.mgoSession.Close()
	}
}

/*New - create new session*/
func (s *QSession) New() *QSession {
	// create abstraction session
	var c *QSession
	if s.mgoSession != nil {
		c = new(QSession)
		c.mgoSession = s.mgoSession.Copy()
		c.SigningKey = s.SigningKey
	}
	return c
}

func (q *Qdb) openMongo() (*QSession, error) {
	// open mongo db
	var err error
	s := new(QSession)
	s.mgoSession, err = mgo.DialWithTimeout(q.URL, q.Timeout)
	if err != nil {
		return nil, err
	}
	s.mgoSession.SetMode(mgo.Monotonic, true)
	s.SigningKey = []byte("FIXME_AND_GENERATE_AND_STORE_TO_DB_INSTEAD")
	err = indexMongo(s.mgoSession)
	return s, err
}

func indexMongo(s *mgo.Session) error {
	// indexing
	taskSession := s.Copy()
	userSession := s.Copy()
	tokenSession := s.Copy()

	defer taskSession.Close()
	defer userSession.Close()
	defer tokenSession.Close()
	c := taskSession.DB("store").C("tasks")
	indexTasks := mgo.Index{
		Key:        []string{"idx"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(indexTasks)
	if err != nil {
		return err
	}
	u := userSession.DB("store").C("users")
	indexUsers := mgo.Index{
		Key:        []string{"hash"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = u.EnsureIndex(indexUsers)
	if err != nil {
		return err
	}
	t := tokenSession.DB("auth").C("token")
	indexTokens := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = t.EnsureIndex(indexTokens)
	if err != nil {
		return err
	}
	return nil
}

/*Open database */
func (q *Qdb) Open() (*QSession, error) {
	var err error
	var session *QSession

	if len(q.Type) == 0 {
		return session, errors.New("Type can't be empty")
	}
	if len(q.URL) == 0 {
		return session, errors.New("Url can't be empty")
	}
	if q.Timeout == (time.Second * 0) {
		q.Timeout = time.Second * 60
	}
	switch strings.ToLower(q.Type) {
	case "mongodb":
		session, err = q.openMongo()
	case "mysql":
		fmt.Println("Mysql - NO IMPLEMENTED !!!!")
		err = errors.New("Not yet supported")
	default:
		err = errors.New("unsupported database type")
	}
	return session, err
}

/*Close close database */
func (q *Qdb) Close() error {
	return nil
}
