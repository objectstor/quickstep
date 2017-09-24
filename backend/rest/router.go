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
func (r *RServer) EnableRest() error {
	var useTokenAuth bool
	r.Mux = goji.NewMux()
	r.Mux.HandleFunc(pat.Post("/login"), doLogin(r.s))
	r.Mux.HandleFunc(pat.Get("/stat"), getStat)

	r.Mux.HandleFunc(pat.Post("/user/:name"), createUser)
	r.Mux.HandleFunc(pat.Get("/user/:name"), getUser)

	//r.Mux.Handle(pat.Get("/task/:id"), getTask)
	r.Mux.HandleFunc(pat.Put("/task"), putTask)
	r.Mux.HandleFunc(pat.Get("/task"), getTasksForUser)
	//r.Mux.HandleFunc(pat.Post("/task/:id"), postTask)

	//r.Mux.hanlerFunc(pat.Get("/action/:id"), getAction) // get action status if task gad action id
	//r.Mux.hanlerFunc(pat.Put("/action/:id"), putAction) // put action status if task gad action id
	//r.Mux.hanlerFunc(pat.Post("/action/:id"), postAction) // post action status if task gad action id

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
