package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"quickstep/backend/rest"
	"quickstep/backend/store"

	"gopkg.in/mgo.v2/bson"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Name            string  `yaml: "name"`
	Db              qdb.Qdb `yaml: "db"`
	RestPlugins     string  `yaml: "plugins"`
	MinPasswdLength int     `yaml: "min_super_passwd"`
}

//CheckOrCreateSuper - set super user or create new one
func CheckOrCreateSuper(password string, s *qdb.QSession, minPasswdLen int) error {
	if s == nil {
		return errors.New("db session error")
	}
	defer s.Close()
	_, err := s.FindUser("system", "")
	if err != nil {
		if qdb.EntryNotFound(err) {
			if len(password) < minPasswdLen {
				msg := fmt.Sprintf("Password too short. Should be at least %d characters long.\n", minPasswdLen)
				return errors.New(msg)
			}
			acl := qdb.CreateACL("", "crud")
			user := new(qdb.User)
			user.ID = bson.NewObjectId()
			user.Name = "system"
			user.ACL = append(user.ACL, *acl)
			h := sha1.New()
			h.Write([]byte(password))
			user.Password = hex.EncodeToString(h.Sum(nil))
			err = s.InsertUser(user)
			if err != nil {
				return err
			}
			log.Printf("system user created\n")
		}
		return err
	}
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
		var pLen int
		pLen = 16 // minimum default length
		if config.MinPasswdLength > pLen {
			pLen = config.MinPasswdLength
		}
		e := CheckOrCreateSuper(*superPassword, session.New(), pLen)
		if e != nil {
			log.Fatal("System user error : ", e)
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
