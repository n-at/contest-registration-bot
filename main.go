package main

import (
	"contest-registration-bot/bot"
	"contest-registration-bot/storage"
	"contest-registration-bot/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var (
	webConfiguration web.Configuration
	botConfiguration bot.Configuration
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
	if err := viper.UnmarshalKey("web", &webConfiguration); err != nil {
		log.Fatalf("Unable to read web configuration: %s", err)
	}
	if err := viper.UnmarshalKey("bot", &botConfiguration); err != nil {
		log.Fatalf("Unable to read bot configuration: %s", err)
	}
}

func main() {
	if err := storage.Open("data/bolt.db"); err != nil {
		log.Fatalf("unable to open storage: %s", err)
	}
	defer storage.Close()

	registrationBot, err := bot.New(botConfiguration)
	if err != nil {
		log.Fatalf("unable to create bot: %s", err)
	}
	registrationBot.Start()

	server := web.NewServer(webConfiguration)
	log.Fatal(server.Start(webConfiguration.Listen))
}
