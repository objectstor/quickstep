package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"expvar"
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
	/* allow only for top level user */
	contextUserString := GetUserFromContext(r)
	_, contextOrg, err := GetUserAndOrg(contextUserString)
	if err != nil {
		JSONError(w, "Auth error ", http.StatusForbidden)
		return
	}
	if len(strings.Split(contextOrg, ".")) > 1 {
		JSONError(w, "Auth error ", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")

	//user := GetUserFromContext(r)
	//if len(user) > 0 {
	//	fmt.Fprintf(w, "change:  user_not_empty\n")
	//	}
	// end of temporary change
}
