package qdb

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//ACLPerm - acl's
type ACLPerm struct {
	Domain string `json: "domain" bson: "domain"`
	Create bool   `json: "create" bson: "create"`
	Read   bool   `json: "read" bson: "create"`
	Update bool   `json: "update" bson: "create"`
	Delete bool   `json "delete bson:"create"`
}

/*User - user schema */
type User struct {
	ID       bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Name     string        `json:"name" bson:"name"`
	Password string        `json:"password,omitempty" bson:"password"`
	Org      string        `json: "org" bson:"org"`
	ACL      []ACLPerm     `json: "acl,omitempty" bson: "acl"`
}

/*Task - token schema */
type Task struct {
	ID               bson.ObjectId   `json:"id,omitempty" bson:"_id"`
	ParentID         bson.ObjectId   `json:"parent_id,omitempty" bson:"parent_id"`
	ChildID          []bson.ObjectId `json:"child_id,omitempty" bson:"child_id"`
	Private          bool            `json:"private" bson:"private"`
	Status           string          `json:"status" bson:"status"`
	CreationTime     time.Time       `bson:"c_time" json:"c_time"`
	DeadLineTime     time.Time       `bson:"d_time" json:"d_time"`
	ModificationTime time.Time       `bson:"m_time" json:"m_time"`
	Name             string          `bson:"name" json:"name"`
	Description      []byte          `bson:"description" json:"description"`
	Comments         []byte          `bson:"comments" json:"comments"`
}

/*UserTask  - user task*/
type UserTask struct {
	ID     bson.ObjectId `json:"id,omitempty" bson:"_id"`
	UserID bson.ObjectId `json:"user_id" bson:"user_id"`
	TaskID bson.ObjectId `json:"task_id" bson:"task_id"`
	ACL    ACLPerm       `json: "acl" bson: "acl"`
}

/*SKeys - secure key schema */
type SKeys struct {
}

//CreateACL - create acl with specyfic org and acl string
func CreateACL(domain string, perm string) *ACLPerm {
	acl := new(ACLPerm)
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
func CheckACL(u *User, domain string, perm string) bool {
	var havePerm bool
	havePerm = false
	for _, acl := range u.ACL {
		if strings.HasSuffix(domain, acl.Domain) {
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
