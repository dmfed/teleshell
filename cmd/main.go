package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/dmfed/conf"
	"github.com/dmfed/teleshell"
)

var (
	varConfigFilename    = ".teleshell.conf"
	varDefaultScriptName = ".teleshell_setup.sh"
)

var (
	varEnvToken    = os.Getenv("TELESHELL_TOKEN")
	varEnvUsername = os.Getenv("TELESHELL_USERNAME")
)

func parseConfig(filename string) (token, username string) {
	cfg, err := conf.ParseFile(filename)
	if err != nil {
		log.Println("could not read config file:", filename)
		return
	}
	token = cfg.Get("token").String()
	username = cfg.Get("username").String()
	return
}

func main() {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "."
	}
	confFile := filepath.Join(dir, varConfigFilename)
	scriptFile := filepath.Join(dir, varDefaultScriptName)

	flag.StringVar(&scriptFile, "onstart", scriptFile, "script file to source when shell starts")
	flag.StringVar(&confFile, "c", confFile, "configuration file to use")
	flag.Parse()

	var token, username string
	// Env variables override config
	if varEnvToken != "" && varEnvUsername != "" {
		token, username = varEnvToken, varEnvUsername
	} else {
		token, username = parseConfig(confFile)
	}
	if token == "" || username == "" {
		log.Println("teleshell: can not proceed without username and token. exiting...")
		return
	}
	shell, err := teleshell.New(token, username, scriptFile)
	if err != nil {
		log.Println("teleshell: error starting bot:", err)
		os.Exit(2)
	} else {
		shell.Start()
		os.Exit(0)
	}
}
