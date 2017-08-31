package qrouter

/* common function for rest
 */
import (
	"errors"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"strings"
	"time"

	"goji.io/pat"
	"gopkg.in/mgo.v2/bson"
)

//Qcontext Session information retrieved from context
type Qcontext struct {
	r         *http.Request
	dbSession *qdb.QSession
	user      string
	userID    string
	u         string
	o         string
}

//NewQContext - create new context
func NewQContext(r *http.Request, validate bool) (*Qcontext, error) {
	if r == nil {
		return nil, errors.New("empty request")
	}
	q := new(Qcontext)
	q.r = r
	session := q.r.Context().Value("dbsession")
	if session != nil {
		q.dbSession = session.(*qdb.QSession)
	}
	arg := q.r.Context().Value("user")
	if arg != nil {
		q.user = arg.(string)
	}
	arg = q.r.Context().Value("user_id")
	if arg != nil {
		q.userID = arg.(string)
	}
	uo := strings.SplitN(q.user, "#", 2)
	if len(uo) != 2 {
		if q.dbSession != nil {
			q.dbSession.Close()
		}
		return nil, errors.New("UserArg bad format")
	}
	if len(uo[1]) == 0 {
		uo[1] = "SYSTEM"
	}
	q.u = uo[0]
	q.o = uo[1]
	if validate {
		if q.dbSession == nil {
			return nil, errors.New("Session error")
		}
		if len(q.user) == 0 {
			if q.dbSession != nil {
				q.dbSession.Close()
			}
			return nil, errors.New("User error")
		}
	}
	return q, nil
}

//UserString get user string
func (q *Qcontext) UserString() string {
	return q.user
}

//UserID - get user id
func (q *Qcontext) UserID() string {
	return q.userID
}

//DBSession  - get database session
func (q *Qcontext) DBSession() *qdb.QSession {
	return q.dbSession
}

//Org - get org
func (q *Qcontext) Org() string {
	return q.o
}

//User - get user name
func (q *Qcontext) User() string {
	return q.u
}

//JSONError send Json error code
func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{error: %q}", message)
}

//JSONOk send Json string
func JSONOk(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{message: %q}", message)
}

//GetParamFromRequest get safe param from request
func GetParamFromRequest(r *http.Request, name string, suffix string) (string, error) {
	path := r.RequestURI
	if strings.HasSuffix(path, suffix) {
		// we should have more than suffix
		return "", errors.New("bad path")
	}
	value := pat.Param(r, name)
	return value, nil
}

//ParseRestTask parse task fill gaps when possible
func ParseRestTask(t *qdb.Task) error {
	if len(t.Name) == 0 {
		return errors.New("Name can't be zero length")
	}
	if t.CreationTime.IsZero() {
		t.CreationTime = time.Now()
	}
	if t.DeadLineTime.IsZero() || t.CreationTime.After(t.DeadLineTime) || t.CreationTime.Equal(t.DeadLineTime) {
		return errors.New("bad time value")
	}
	if len(t.ParentID) > 0 {
		if !bson.IsObjectIdHex(t.ParentID) {
			return errors.New("wrong ID")
		}
	}
	for _, childID := range t.ChildID {
		if !bson.IsObjectIdHex(childID) {
			return errors.New("wrong ID")
		}
	}
	t.ModificationTime = time.Now()
	return nil
}
