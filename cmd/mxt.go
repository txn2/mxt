package main

import (
	"flag"
	"os"

	"github.com/txn2/micro"
	"github.com/txn2/mxt"
	"go.uber.org/zap"
)

var (
	configFileEnv = getEnv("CONFIG", "./cfg/simple.yml")
)

func main() {
	configFile := flag.String("config", configFileEnv, "Config file")

	serverCfg, _ := micro.NewServerCfg("mxt")
	server := micro.NewServer(serverCfg)

	epCfg, err := mxt.CfgFromFile(*configFile)
	if err != nil {
		server.Logger.Fatal("Config file error", zap.Error(err))
	}

	pApi, err := mxt.NewProxy(&mxt.ProxyCfg{
		EpConfig:   epCfg,
		Logger:     server.Logger,
		HttpClient: server.Client,
	})
	if err != nil {
		server.Logger.Fatal("Unable to create proxy.", zap.Error(err))
	}

	// handle endpoints
	server.Router.GET("/get/:ep", pApi.EpHandler)

	// run provisioning server
	server.Run()
}

// getEnv gets an environment variable or sets a default if
// one does not exist.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}
