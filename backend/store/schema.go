package qdb

import (
	"strings"

	"gopkg.in/mgo.v2/bson"
)

/*Tasks - task schema */
type Tasks struct {
}

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

/*Tokens - token schema */
type Tokens struct {
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
		}
		break
	}
	return havePerm
}
