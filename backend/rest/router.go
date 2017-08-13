package qrouter

/* REST server*/

import (
	"quickstep/backend/store"
	"strings"

	"goji.io"
	"goji.io/pat"
)

// main rest

type rServer struct {
	url      string
	s        *qdb.QSession
	Mux      *goji.Mux
	pluggins string
}

/* create new rest server , port and db connection must be provided */
func New(url string, s *qdb.QSession) (*rServer, error) {
	r := new(rServer)
	r.url = url
	r.s = s
	return r, nil
}

func (r *rServer) Enable() error {
	r.Mux = goji.NewMux()
	r.Mux.HandleFunc(pat.Post("/login"), doLogin(r.s))
	//mux.HandleFunc(pat.Head("/task"), getAllTasks(r.s, true))
	//mux.HandleFunc(pat.Get("/task"), getAllTasks(r.s, false))
	//mux.HandleFunc(pat.Head("/task/:id"), getTaskById(r.s, true))
	//mux.HandleFunc(pat.Get("/task/:id"), getTaskById(r.s, false))
	//mux.HandleFunc(pat.Put("/task/:id"), storeTaskById(r.s))
	//mux.HandleFunc(pat.Put("/user"),adduser(r.s))
	//mux.HandleFunc(pat.Get("/user"), getUser(r.s))
	for _, plugin := range strings.Split(r.pluggins, ",") {
		switch strings.ToLower(plugin) {
		case "logging":
			r.Mux.Use(logging) //!Untested
		case "rauth":
			r.Mux.Use(rauth) //!untested
		}
	}
	return nil
}

func (r *rServer) Stop() error {
	return nil
}
