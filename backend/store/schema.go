package qdb

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//ACLPerm - acl's
type ACLPerm struct {
	User   string `json:"user,omitempty" bson:"user"`
	Domain string `json:"domain" bson:"domain"`
	Create bool   `json:"create" bson:"create"`
	Read   bool   `json:"read" bson:"read"`
	Update bool   `json:"update" bson:"update"`
	Delete bool   `json:"delete" bson:"delete"`
}

/*User - user schema */
type User struct {
	ID       bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Name     string        `json:"name" bson:"name"`
	Password string        `json:"password,omitempty" bson:"password"`
	Org      string        `json:"org" bson:"org"`
	ACL      []ACLPerm     `json:"acl,omitempty" bson:"acl"`
}

/*Task - token schema */
type Task struct {
	ID               bson.ObjectId `json:"id,omitempty" bson:"_id"`
	ParentID         string        `json:"parent_id,omitempty" bson:"parent_id"`
	ChildID          []string      `json:"child_id,omitempty" bson:"child_id"`
	Private          bool          `json:"private" bson:"private"`
	Status           string        `json:"status" bson:"status"`
	CreationTime     time.Time     `bson:"c_time" json:"c_time"`
	DeadLineTime     time.Time     `bson:"d_time" json:"d_time"`
	ModificationTime time.Time     `bson:"m_time" json:"m_time"`
	Name             string        `bson:"name" json:"name"`
	Description      []byte        `bson:"description" json:"description"`
	Comments         []byte        `bson:"comments" json:"comments"`
}

/*UserTask  - user task*/
type UserTask struct {
	TaskID       string    `json:"taskid" bson:"taskid"`
	UserID       string    `json: "userid" bson:"userid"`
	TaskName     string    `json: "name" bson: "name"`
	CreationTime time.Time `json:"c_time" bson:"c_time"`
	DeadLineTime time.Time `json:"d_time" bson:"d_time"`
	ACL          ACLPerm   `json:"acl" bson:"acl"`
}

/*SKeys - secure key schema */
type SKeys struct {
}

//CreateACL - create acl with specyfic org and acl string
func CreateACL(user string, domain string, perm string) *ACLPerm {
	acl := new(ACLPerm)
	if len(user) > 0 {
		acl.User = user
	}
	acl.Domain = domain
	acl.Create = false
	acl.Read = false
	acl.Update = false
	acl.Delete = false
	for _, sPerm := range strings.Split(strings.ToLower(perm), "") {
		switch sPerm {
		case "c":
			acl.Create = true
		case "r":
			acl.Read = true
		case "u":
			acl.Update = true
		case "d":
			acl.Delete = true
		}
	}
	return acl
}

//CheckACL - check acl permission
func CheckACL(u *User, user string, domain string, perm string) bool {
	var havePerm bool
	var proceed bool
	havePerm = false
	proceed = false
	for _, acl := range u.ACL {
		if len(user) > 0 {
			if strings.Compare(user, acl.User) == 0 && strings.Compare(domain, acl.Domain) == 0 {
				proceed = true
			}
		} else {
			if strings.HasSuffix(domain, acl.Domain) {
				proceed = true
			}
		}
		if proceed {
			for _, sPerm := range strings.Split(strings.ToLower(perm), "") {
				switch sPerm {
				case "c":
					havePerm = acl.Create
				case "r":
					havePerm = acl.Read
				case "u":
					havePerm = acl.Update
				case "d":
					havePerm = acl.Delete
				}
			}
			break
		}
	}
	return havePerm
}
