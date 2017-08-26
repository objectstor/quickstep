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
	"gopkg.in/mgo.v2/bson"
)

/*Qdb - dtabase abstraction class */
type Qdb struct {
	Type    string `yaml:"type"`
	URL     string `yaml:"url"`
	Timeout time.Duration
}

/*QSession database session abtraction struct */
type QSession struct {
	mgoSession *mgo.Session
	SigningKey []byte
}

//EntryNotFound check is erro == Not found
func EntryNotFound(err error) bool {
	if err == mgo.ErrNotFound {
		return true
	}
	//TODO add support for mysql
	return false
}

/*Close close session */
func (s *QSession) Close() {

	if s != nil && s.mgoSession != nil {
		s.mgoSession.Close()
	}
}

/*New - create new session*/
func (s *QSession) New() *QSession {
	// create abstraction session
	var c *QSession
	if s != nil && s.mgoSession != nil {
		c = new(QSession)
		c.mgoSession = s.mgoSession.Copy()
		c.SigningKey = s.SigningKey
	}
	return c
}

//FindUser - user with specific Name
func (s *QSession) FindUser(name string, org string) (*User, error) {
	if s != nil && s.mgoSession != nil && len(name) > 0 {
		c := s.mgoSession.DB("store").C("users")
		result := User{}
		err := c.Find(bson.M{"name": name, "org": org}).One(&result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	// do mysql when supported
	return nil, &mgo.QueryError{0, "Bad parameters or db not defined", false}
}

//InsertUser - user with specific Name
func (s *QSession) InsertUser(user *User) error {
	if len(user.Name) == 0 {
		return errors.New("user name can't be empty")
	}
	if s != nil && s.mgoSession != nil {
		c := s.mgoSession.DB("store").C("users")
		if !user.ID.Valid() {
			user.ID = bson.NewObjectId()
		}
		err := c.Update(bson.M{"name": user.Name, "org": user.Org}, user)
		if err != nil {
			if EntryNotFound(err) {
				err := c.Insert(user)
				if err != nil {
					return err
				}
				return nil
			}
			return err
		}
		return nil
	}
	// do mysql when supported
	return errors.New("session empty or unsupported engine")
}

//DeleteUser - delete user with specific name
func (s *QSession) DeleteUser(userName string, org string) error {

	if len(userName) == 0 {
		return errors.New("User name can't be null")
	}
	if s != nil && s.mgoSession != nil {
		c := s.mgoSession.DB("store").C("users")
		err := c.Remove(bson.M{"name": userName, "org": org})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("nil session")
}

//InsertTask - insert new task
func (s *QSession) InsertTask(task *Task) (string, error) {
	if s != nil && s.mgoSession != nil {
		if len(task.Name) == 0 {
			return "", errors.New("name empty")
		}
		c := s.mgoSession.DB("store").C("tasks")
		if !task.ID.Valid() {
			task.ID = bson.NewObjectId()
		}
		err := c.Insert(task)
		if err != nil {
			return "", err
		}
		return task.ID.Hex(), err
	}
	return "", errors.New("session empty or unsupported engine")
}

//FindTask with specific id
func (s *QSession) FindTask(ID string) (*Task, error) {
	if s != nil && s.mgoSession != nil && len(ID) > 0 {
		c := s.mgoSession.DB("store").C("tasks")
		result := Task{}
		bID := bson.ObjectIdHex(ID)
		err := c.Find(bson.M{"_id": bID}).One(&result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	// do mysql when supported
	return nil, &mgo.QueryError{0, "Bad parameters or db not defined", false}

}

//InsertUserTask - insert User task permissions
func (s *QSession) InsertUserTask(ut *UserTask) error {
	if s != nil && s.mgoSession != nil {
		if len(ut.TaskID) == 0 {
			return errors.New("TaskId empty")
		}
		if len(ut.UserID) == 0 {
			return errors.New("UserId empty")
		}

		if !bson.IsObjectIdHex(ut.TaskID) {
			return errors.New("bad taskID")
		}
		if !bson.IsObjectIdHex(ut.UserID) {
			return errors.New("bad userID")
		}
		// I'm not checking if taskId and UserId exists
		// this should be done on higher level
		c := s.mgoSession.DB("store").C("permissions")
		err := c.Insert(ut)
		return err
	}
	return &mgo.QueryError{0, "Bad parameters or db not defined", false}
}

//FindUserTasks return single or multiple tasks task
// based on arguments  if only user is specified
// return list of tasks
// if user and task return single task
func (s *QSession) FindUserTasks(UserID string, TaskID string) ([]UserTask, error) {
	var results []UserTask
	if s != nil && s.mgoSession != nil {
		if len(UserID) == 0 && !bson.IsObjectIdHex(UserID) {
			return results, errors.New("UserId empty")
		}
		c := s.mgoSession.DB("store").C("permissions")
		ut := new(UserTask)
		ut.UserID = UserID

		if len(TaskID) > 0 {
			//return single
			//f body
			if !bson.IsObjectIdHex(TaskID) {
				return results, errors.New("TaskId incorrect")
			}

			ut.TaskID = TaskID
			err := c.Find(ut).All(&results)
			if err != nil {
				return results, err
			}
			return results, nil
		}
		// return multi
		err := c.Find(bson.M{"userid": UserID}).All(&results)
		if err != nil {
			return results, err
		}
		return results, nil

	}
	return results, &mgo.QueryError{0, "Bad parameters or db not defined", false}
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
	permSession := s.Copy()

	defer taskSession.Close()
	defer userSession.Close()
	defer permSession.Close()
	c := taskSession.DB("store").C("tasks")
	indexTasks := mgo.Index{
		Key:        []string{"name"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(indexTasks)
	if err != nil {
		return err
	}
	u := userSession.DB("store").C("users")
	indexUsers := mgo.Index{
		Key:        []string{"name"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = u.EnsureIndex(indexUsers)
	if err != nil {
		return err
	}
	t := permSession.DB("store").C("permissions")
	indexPermUsers := mgo.Index{
		Key:        []string{"user_id"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = t.EnsureIndex(indexPermUsers)
	if err != nil {
		return err
	}

	indexPermTask := mgo.Index{
		Key:        []string{"task_id"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = t.EnsureIndex(indexPermTask)
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
