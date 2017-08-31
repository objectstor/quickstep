package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"quickstep/backend/store"
)

func createUser(w http.ResponseWriter, r *http.Request) {
	var qUser qdb.User

	httpUser, err := GetParamFromRequest(r, "name", "/user")
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	ctx, err := NewQContext(r, true)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	session := ctx.DBSession()
	defer session.Close()

	// let's decode body
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&qUser)
	if err != nil {
		JSONError(w, "Content error", http.StatusBadRequest)
		return
	}
	if qUser.Name != httpUser {
		JSONError(w, "Syntax error", http.StatusBadRequest)
		return
	}
	if len(qUser.Org) == 0 {
		JSONError(w, "org can't be empty", http.StatusBadRequest)
		return
	}
	creator, dberr := session.FindUser(ctx.User(), ctx.Org())
	if dberr != nil {
		JSONError(w, dberr.Error(), http.StatusBadRequest)
		return
	}
	// we need to know if user exists and is hould be modified
	// or we should create new One
	nUser, dberr := session.FindUser(qUser.Name, qUser.Org)
	if dberr != nil {
		if qdb.EntryNotFound(dberr) {
			dberr = errors.New("bad permissions")
			if qdb.CheckACL(creator, "", qUser.Org, "c") {
				dberr = session.InsertUser(&qUser)
				if dberr == nil {
					JSONOk(w, "Created")
					return
				}
			}
		}
		JSONError(w, dberr.Error(), http.StatusBadRequest)
		return
	}
	// create
	dberr = errors.New("bad permissions")
	if qdb.CheckACL(creator, "", qUser.Org, "u") {
		qUser.ID = nUser.ID
		dberr = session.InsertUser(&qUser)
		if dberr == nil {
			JSONOk(w, "Updated")
			return
		}
	}
	JSONError(w, dberr.Error(), http.StatusBadRequest)
	return
	// get antry from db and check if we can do something within domain
}

func getUser(w http.ResponseWriter, r *http.Request) {
	var qUser qdb.User
	ctx, err := NewQContext(r, true)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}
	session := ctx.DBSession()
	defer session.Close()
	httpUser, err := GetParamFromRequest(r, "name", "/user")
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}

	//body zero we can get info about ourself
	if r.ContentLength == 0 {
		if httpUser == ctx.User() {
			nUser, err := session.FindUser(ctx.User(), ctx.Org())
			if err != nil {
				if qdb.EntryNotFound(err) {
					err = errors.New("Access error")
				}
				JSONError(w, err.Error(), http.StatusForbidden)
				return
			}
			nUser.Password = "" // remove password
			nUser.ID = ""       // remove ID
			j, err := json.Marshal(nUser)
			if err != nil {
				JSONError(w, err.Error(), http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Location", r.URL.Path)
			fmt.Fprintf(w, "%s", string(j))
			return
		}
		JSONError(w, "Access error", http.StatusForbidden)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&qUser)
	if err != nil {
		JSONError(w, "Syntax error", http.StatusBadRequest)
		return
	}

	creator, dberr := session.FindUser(ctx.User(), ctx.Org())
	if dberr != nil {
		JSONError(w, dberr.Error(), http.StatusBadRequest)
		return
	}
	if qdb.CheckACL(creator, "", qUser.Org, "r") {
		nUser, dberr := session.FindUser(qUser.Name, qUser.Org)
		if dberr != nil {
			if qdb.EntryNotFound(dberr) {
				dberr = errors.New("Access error")
			}
			JSONError(w, dberr.Error(), http.StatusBadRequest)
			return
		}
		nUser.ID = ""
		nUser.Password = ""
		j, err := json.Marshal(nUser)
		if err != nil {
			JSONError(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path)
		fmt.Fprintf(w, "%s", string(j))
		return
	}
	JSONError(w, "Access error", http.StatusForbidden)
}
