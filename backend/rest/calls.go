package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"quickstep/backend/store"
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
			fmt.Printf("AUTHENTICATION: owner %s\n", owner)
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
	session := GetDbSessionFromContext(r)
	user := GetUserFromContext(r)
	if !ValidUserAndSession(session, user, w) {
		return
	}

	fmt.Printf("create/update  new user: user from token %s\n", user)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	session := GetDbSessionFromContext(r)
	user := GetUserFromContext(r)
	if !ValidUserAndSession(session, user, w) {
		return
	}
	fmt.Printf("get new user - (password not seen) - token from user %s\n", user)
}
