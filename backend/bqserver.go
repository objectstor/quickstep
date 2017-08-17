package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"quickstep/backend/rest"
	"quickstep/backend/store"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Name        string  `yaml: "name"`
	Db          qdb.Qdb `yaml: "db"`
	RestPlugins string  `yaml: "plugins"`
}

//CheckOrCreateSuper - set super user or create new one
func CheckOrCreateSuper(password string) error {
	log.Printf("checking passwd %s", password)
	return nil
}

func main() {
	var configName = flag.String("config", ".qstepserver.conf", "bqstepserver config")
	var verbose = flag.Bool("verbose", true, "set to true to verbose mode")
	var restURL = flag.String("url", "localhost:9090", "web rest url")
	var superPassword = flag.String("passwd", "", "initial password")
	var config Config

	flag.Parse()
	if !*verbose {
		logf, err := os.OpenFile("qstep.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			panic(err)
		}
		defer logf.Close()
		log.SetOutput(logf)
	}

	source, err := ioutil.ReadFile(*configName)
	if err != nil {
		log.Fatal("Config file read failed : ", err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Fatal("Config file content error : ", err)
	}

	session, err := config.Db.Open()
	if err != nil {
		log.Fatal("Database connection failed : ", err)
	}
	if len(*superPassword) > 0 {
		e := CheckOrCreateSuper(*superPassword)
		if e != nil {
			log.Fatal("Super user access error : ", err)
		}
	}
	defer config.Db.Close()
	router, err := qrouter.New(*restURL, session)
	if err != nil {
		log.Fatal("Router create failed : ", err)
	}
	if len(config.RestPlugins) > 0 {
		err = router.EnablePlugins(config.RestPlugins)
		if err != nil {
			log.Fatal("Can't bring up plugins : ", err)
		}
	}
	err = router.Enable()
	if err != nil {
		log.Fatal("Router init failed : ", err)
	}
	err = http.ListenAndServe("localhost:8000", router.Mux)
	if err != nil {
		log.Fatal("Router failed to start : ", err)
	}
	defer router.Stop()
}
