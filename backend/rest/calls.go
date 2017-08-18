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
	"time"

	"goji.io/pat"

	"github.com/dgrijalva/jwt-go"
)

//QuickStepUserClaims used by token
type QuickStepUserClaims struct {
	Owner string `json:"owner"`
	jwt.StandardClaims
}

/*UserAuth - user auth info */
type UserAuth struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Org      string `json: "org"` // additional org string ex. shop.com, research_group
}

func doLogin(s *qdb.QSession) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s == nil {
			JSONError(w, "database not connected ", http.StatusInternalServerError)
			return
		}
		session := s.New()
		defer session.Close()
		var userAuth UserAuth
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&userAuth)
		if err != nil {
			JSONError(w, "Auth error", http.StatusForbidden)
			return
		}
		if len(userAuth.Name) == 0 || len(userAuth.Password) == 0 {
			JSONError(w, "Auth error", http.StatusForbidden)
			return
		}
		// get user
		user, err := session.FindUser(userAuth.Name, userAuth.Org)
		if err != nil {
			JSONError(w, "Auth error", http.StatusForbidden)
			return
		}
		if user.Name == userAuth.Name && user.Password == userAuth.Password {
			// create user string
			owner := fmt.Sprintf("%s#%s", user.Name, user.Org)
			// Create the Claims
			claims := QuickStepUserClaims{
				owner, // this chould be user id
				jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
					Issuer:    "quickStep", // this should be host most likelly
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(s.SigningKey)
			if err != nil {
				JSONError(w, "Auth error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "{\"token\": \"%s\"}", tokenString)
		} else {
			JSONError(w, "Auth error", http.StatusForbidden)
			return
		}
	}
}

func getStat(w http.ResponseWriter, r *http.Request) {
	//temporary change
	session := GetDbSessionFromContext(r)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", r.URL.Path)
	if session != nil {
		fmt.Fprintf(w, "change:  key_not_empty\n")
	}
	user := GetUserFromContext(r)
	if len(user) > 0 {
		fmt.Fprintf(w, "change:  user_not_empty\n")
	}
	// end of temporary change
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var qUser qdb.User
	httpUser := pat.Param(r, "name")
	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	if !ValidUserAndSession(session, contextUserString, w) {
		return
	}
	contextUser, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Context error ", http.StatusBadRequest)
		return
	}

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
	creator, dberr := session.FindUser(contextUser, contextOrg)
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
			if qdb.CheckACL(creator, qUser.Org, "c") {
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
	if qdb.CheckACL(creator, qUser.Org, "u") {
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
	session := GetDbSessionFromContext(r)
	contextUserString := GetUserFromContext(r)
	httpUser := pat.Param(r, "name")
	if !ValidUserAndSession(session, contextUserString, w) {
		return
	}
	contextUser, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Syntax error", http.StatusBadRequest)
		return
	}
	//body zero we can get info about ourself
	if r.ContentLength == 0 {
		if httpUser == contextUser {
			nUser, err := session.FindUser(contextUser, contextOrg)
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
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&qUser)
	if err != nil {
		JSONError(w, "Syntax error", http.StatusBadRequest)
		return
	}

	creator, dberr := session.FindUser(contextUser, contextOrg)
	if dberr != nil {
		JSONError(w, dberr.Error(), http.StatusBadRequest)
		return
	}
	if qdb.CheckACL(creator, qUser.Org, "r") {
		nUser, dberr := session.FindUser(qUser.Name, qUser.Org)
		if dberr != nil {
			if qdb.EntryNotFound(err) {
				dberr = errors.New("Access error")
			}
			JSONError(w, dberr.Error(), http.StatusBadRequest)
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
