package qrouter

/* middlewares for rest */

import (
	"context"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type tokenAuth struct {
	h       http.Handler
	key     []byte
	owner   string
	session *qdb.QSession
}

func (t tokenAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user string
	var status bool
	if strings.Compare(r.URL.String(), "/login") != 0 { // for all url except login
		user, status = t.authenticate(r)
		if status == false {
			JsonError(w, "Auth Error", http.StatusForbidden)
			return
		}
	}
	if t.session != nil {
		dbsession := t.session.New()
		defer dbsession.Close() // clean up
		ctx := context.WithValue(r.Context(), "dbsession", dbsession)
		if len(user) > 0 {
			ctx = context.WithValue(ctx, "user", user)
		}
		t.h.ServeHTTP(w, r.WithContext(ctx))
	} else {
		t.h.ServeHTTP(w, r)
	}
}

// authenticate - actual authentication
func (t *tokenAuth) authenticate(r *http.Request) (string, bool) {
	tokenString := r.Header.Get("X-Golden-Ticket")
	if len(tokenString) == 0 {
		return "", false
	}
	token, err := jwt.ParseWithClaims(tokenString, &QuickStepUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return t.key, nil
	})
	if err == nil {
		if claims, ok := token.Claims.(*QuickStepUserClaims); ok && token.Valid {
			if claims.VerifyIssuer("quickStep", true) {
				if claims.VerifyExpiresAt(time.Now().Unix(), true) {
					// we have valid token
					return claims.Owner, true
				}
			}
		}
	}
	return "", false
}

// TokenAuth -  token authentication
func TokenAuth(session *qdb.QSession) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		var skey []byte
		if session == nil {
			skey = make([]byte, 0)
		} else {
			skey = session.SigningKey
		}
		return tokenAuth{h, skey, "", session}
	}
	return fn
}
