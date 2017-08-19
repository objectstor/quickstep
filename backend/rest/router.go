package qrouter

/* REST server*/

import (
	"quickstep/backend/store"
	"strings"

	"goji.io"
	"goji.io/pat"
)

// main rest

/*RServer - Rest Server struct */
type RServer struct {
	url     string
	s       *qdb.QSession
	Mux     *goji.Mux
	plugins string
}

/*New create new rest server , port and db connection must be provided */
func New(url string, s *qdb.QSession) (*RServer, error) {
	r := new(RServer)
	r.url = url
	r.s = s
	return r, nil
}

/*EnablePlugins *enable plugins string */
func (r *RServer) EnablePlugins(plugins string) error {
	r.plugins = plugins
	return nil
}

/*Enable prepare server for work */
func (r *RServer) Enable() error {
	var useTokenAuth bool
	r.Mux = goji.NewMux()
	r.Mux.HandleFunc(pat.Post("/login"), doLogin(r.s))
	r.Mux.HandleFunc(pat.Get("/stat"), getStat)
	r.Mux.HandleFunc(pat.Post("/user/:name"), createUser)
	r.Mux.HandleFunc(pat.Get("/user/:name"), getUser)

	r.Mux.HandleFunc(pat.Head("/task"), headTasks)
	r.Mux.HandleFunc(pat.Post("/task/:id"), postTask)
	r.Mux.HandleFunc(pat.Put("/task/:id"), putTask)

	//mux.HandleFunc(pat.Get("/task"), getAllTasks(r.s, false))
	//mux.HandleFunc(pat.Head("/task/:id"), getTaskById(r.s, true))
	//mux.HandleFunc(pat.Get("/task/:id"), getTaskById(r.s, false))
	//mux.HandleFunc(pat.Put("/task/:id"), storeTaskById(r.s))
	useTokenAuth = false
	for _, plugin := range strings.Split(r.plugins, ",") {
		switch strings.ToLower(plugin) {
		case "logging":
			r.Mux.Use(logging) //!Untested
		}
	}
	if !useTokenAuth {
		r.Mux.Use(TokenAuth(r.s)) //!untested
	}
	return nil
}

/*Stop stop RServer */
func (r *RServer) Stop() error {
	return nil
}
