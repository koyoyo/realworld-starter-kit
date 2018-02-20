package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"

	"github.com/koyoyo/realworld-starter-kit/handlers"
	"github.com/koyoyo/realworld-starter-kit/models"
)

func main() {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
		viper.AutomaticEnv()
	}

	db, err := gorm.Open("postgres", viper.Get("POSTGRES_URL"))
	if err != nil {
		panic(fmt.Errorf("Fatal db connect: %s \n", err))
	}
	defer db.Close()

	// Initial Schema
	db.AutoMigrate(&models.User{})

	fmt.Println("Hello World!!")

	app := handlers.App{
		DB: models.DB{
			db,
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/users", app.RegisterHandler)
	http.Handle("/", r)
	http.ListenAndServe(viper.GetString("GO_PORT"), nil)
}
