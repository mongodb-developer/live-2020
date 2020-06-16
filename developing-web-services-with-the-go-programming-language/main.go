package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Podcast struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Title  string             `bson:"name,omitempty"`
	Author string             `bson:"author,omitempty"`
	Tags   []string           `bson:"tags,omitempty"`
}

type Episode struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Podcast     primitive.ObjectID `bson:"podcast,omitempty"`
	Title       string             `bson:"title,omitempty"`
	Description string             `bson:"description,omitempty"`
	Duration    int32              `bson:"duration,omitempty"`
}

type Connection struct {
	Podcasts *mongo.Collection
	Episodes *mongo.Collection
}

func (connection Connection) CreatePodcastEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var podcast Podcast
	if err := json.NewDecoder(request.Body).Decode(&podcast); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := connection.Podcasts.InsertOne(ctx, podcast)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func (connection Connection) GetPodcastsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var podcasts []Podcast
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := connection.Podcasts.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if err = cursor.All(ctx, &podcasts); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(podcasts)
}

func (connection Connection) UpdatePodcastEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var podcast Podcast
	json.NewDecoder(request.Body).Decode(&podcast)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := connection.Podcasts.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$set", podcast},
		},
	)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func (connection Connection) DeletePodcastEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := connection.Podcasts.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("ATLAS_URI")))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	connection := Connection{
		Podcasts: client.Database("quickstart").Collection("podcasts"),
		Episodes: client.Database("quickstart").Collection("episodes"),
	}

	router := mux.NewRouter()
	router.HandleFunc("/podcast", connection.CreatePodcastEndpoint).Methods("POST")
	router.HandleFunc("/podcasts", connection.GetPodcastsEndpoint).Methods("GET")
	router.HandleFunc("/podcast/{id}", connection.UpdatePodcastEndpoint).Methods("PUT")
	router.HandleFunc("/podcast/{id}", connection.DeletePodcastEndpoint).Methods("DELETE")
	http.ListenAndServe(":12345", router)
}
