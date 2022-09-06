package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NgMongoEntry struct {
	RelativePath string `bson:"relative_path"`
	Content      []byte `bson:"content"`
}

type NgMongoRouter struct {
	router         *regexRouter
	client         *mongo.Client
	collectionMaps map[string]*mongo.Collection
}

func (router *NgMongoRouter) getCollectionHandler(collectionName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		prefix := fmt.Sprintf("%s/%s/", pathName, collectionName)
		replacedPath := strings.Replace(req.URL.Path, prefix, "", 1)

		result := NgMongoEntry{}
		filter := bson.D{primitive.E{Key: "relative_path", Value: replacedPath}}

		collection, found := router.collectionMaps[collectionName]
		if !found {
			errorText := fmt.Sprintf("Cannot find collection with name %s", collectionName)
			http.Error(w, errorText, http.StatusNotFound)
			return
		}

		err := collection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				errorText := fmt.Sprintf("Cannot find document with relative_path %s", replacedPath)
				http.Error(w, errorText, http.StatusNotFound)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(result.Content)
	}
}

func (router *NgMongoRouter) Init() {
	fmt.Printf("Connecting mongo instance at %s\n", mongoPath)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoPath))
	if err != nil {
		panic(err)
	}
	router.client = client
	router.collectionMaps = make(map[string]*mongo.Collection)
	router.router = &regexRouter{}
	database := client.Database("chunk_store")

	collectionList, err := database.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		panic(err)
	}

	for _, collectionName := range collectionList {
		route := fmt.Sprintf("%s/%s", pathName, collectionName)
		fmt.Printf("Adding route for collection %s as %s\n", collectionName, route)
		router.collectionMaps[collectionName] = database.Collection(collectionName)

		router.router.AddRegExpRoute(route, http.HandlerFunc(router.getCollectionHandler(collectionName)))

	}
}

func (router *NgMongoRouter) Cleanup() {
	if err := router.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (router *NgMongoRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	router.router.ServeHTTP(w, r)
}

func main() {
	router := &NgMongoRouter{}
	router.Init()

	http.ListenAndServe(":7999", router)
	fmt.Printf("ping")
}
