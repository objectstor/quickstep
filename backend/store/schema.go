package qdb

import (
	"gopkg.in/mgo.v2/bson"
)

/*Tasks - task schema */
type Tasks struct {
}

/*User - user schema */
type User struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Name     string        `json:"name" bson:"name"`
	Password string        `json:"password" bson:"password"`
	Org      string        `json: "org" bson:"org"`
}

/*Tokens - token schema */
type Tokens struct {
}

/*SKeys - secure key schema */
type SKeys struct {
}
