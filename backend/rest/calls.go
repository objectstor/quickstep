package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"strings"
	"time"

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
			JsonError(w, "database not connected ", http.StatusInternalServerError)
			return
		}
		session := s.New()
		defer session.Close()
		var userAuth UserAuth
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&userAuth)
		if err != nil {
			JsonError(w, "Auth error", http.StatusForbidden)
			return
		}
		if len(userAuth.Name) == 0 || len(userAuth.Password) == 0 {
			JsonError(w, "Auth error", http.StatusForbidden)
			return
		}
		// get user
		user, err := session.FindUser(userAuth.Name)
		if err != nil {
			JsonError(w, "Auth error", http.StatusForbidden)
			return
		}
		if user.Name == userAuth.Name && user.Password == userAuth.Password {
			// create user string
			owner := fmt.Sprintf("%s.%s", user.Name, user.Org)
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
				JsonError(w, "Auth error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "{\"token\": \"%s\"}", tokenString)
		} else {
			JsonError(w, "Auth error", http.StatusForbidden)
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
	var dbUser qdb.User
	session := GetDbSessionFromContext(r)
	restuser := GetUserFromContext(r)
	if !ValidUserAndSession(session, restuser, w) {
		return
	}
	acl := strings.SplitN(restuser, ".", 2)
	if len(acl[1]) == 0 {
		acl[1] = "SYSTEM"
	}
	// let's decode body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&dbUser)
	if err != nil {
		JsonError(w, "Auth error", http.StatusForbidden)
		return
	}
	fmt.Printf("createUser() User.Name: %s User.Org: %s, token# user: \"%s\" org: %s\n", dbUser.Name, dbUser.Org, acl[0], acl[1])
	if len(dbUser.Org) == 0 {
		fmt.Println("failed\n")
	}
	// get antry from db and check if we can do something within domain
}

func getUser(w http.ResponseWriter, r *http.Request) {
	session := GetDbSessionFromContext(r)
	restuser := GetUserFromContext(r)
	if !ValidUserAndSession(session, restuser, w) {
		return
	}
	acl := strings.SplitN(restuser, ".", 2)
	if len(acl[1]) == 0 {
		acl[1] = "SYSTEM"
	}
	fmt.Printf("getUser() token user: \"%s\" org: %s\n", acl[0], acl[1])
	// we should check if in body other domain is not specified
	// if so we should check if we have access to that domain
}
