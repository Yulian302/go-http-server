package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"context"

	"github.com/gofor-little/env"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ENV_FILE_PATH = ".env"

type Book struct {
	Title  string `bson:"title,omitempty"`
	Author string `bson:"author,omitempty"`
	Price  uint16 `bson:"price,omitempty"`
}

func main() {
	// loading env file
	if err := env.Load(ENV_FILE_PATH); err != nil {
		panic(err)
	}
	clusterName := env.Get("MONGO_CLUSTER", "")
	appName := env.Get("MONGO_APP_NAME", "")
	clusterPassword := env.Get("MONGO_PASSWORD", "")
	// mongodb
	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(fmt.Sprintf("mongodb+srv://%s:%s@%s.hfa617f.mongodb.net/?retryWrites=true&w=majority&appName=%s", clusterName, clusterPassword, strings.ToLower(clusterName), appName)).SetServerAPIOptions(serverApi)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	if err := client.Database("test").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// main router
	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("static"))
	// books
	bookrouter := router.PathPrefix("/books").Subrouter()
	bookrouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		var books []Book
		booksColl := client.Database("test").Collection("books")
		cursor, booksError := booksColl.Find(context.TODO(), bson.D{})
		if booksError != nil {
			panic(booksError)
		}
		if booksError = cursor.All(context.TODO(), &books); err != nil {
			panic(booksError)
		}
		jsonBooks, err := json.Marshal(books)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBooks)
	}).Methods("GET")

	// serving static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	http.ListenAndServe("localhost:8082", router)
}
