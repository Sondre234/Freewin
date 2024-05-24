package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

import (
	"html/template"
)

type Comment struct {
	ID      string    `bson:"_id"`
	Name    string    `bson:"name"`
	Email   string    `bson:"email"`
	MovieID string    `bson:"movie_id"`
	Text    string    `bson:"text"`
	Date    time.Time `bson:"date"`
}

func lol() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client().ApplyURI("mongodb+srv://sondre:passord@easywin.ihpovx4.mongodb.net/?retryWrites=true&w=majority&appName=Easywin").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	log.Println("Connected to MongoDB!")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		collection := client.Database("sample_mflix").Collection("comments")
		var comments []Comment

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal("Error finding documents: ", err)
		}

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var comment Comment
			if err := cursor.Decode(&comment); err != nil {
				log.Fatal("Error decoding document: ", err)
			}
			comments = append(comments, comment)
		}

		if err := cursor.Err(); err != nil {
			log.Fatal("Cursor error: ", err)
		}

		tmpl := template.Must(template.ParseFiles("index.html"))
		if err := tmpl.Execute(w, comments); err != nil {
			log.Fatal("Error executing template: ", err)
		}
	})

	log.Println("Starting server on :8000")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", nil))
}
