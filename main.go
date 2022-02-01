package main

import (
	"contest-registration-bot/storage"
	"contest-registration-bot/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}
}

func main() {
	err := storage.Open("data/bolt.db")
	if err != nil {
		log.Fatalf("unable to open storage: %s", err)
	}
	defer storage.Close()

	server := web.NewServer()
	log.Fatal(server.Start(":3000"))
}
