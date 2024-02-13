package main

import (
	// "context"
	// "fmt"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	// "os"

	"github.com/gorilla/mux"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	MONGODB_URI := "mongodb://localhost:8005/"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MONGODB_URI))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	db := client.Database("db")
	opts := options.GridFSBucket().SetName("Test Bucket")
	bucket, err := gridfs.NewBucket(db, opts)
	if err != nil {
		panic(err)
	}

	file, err := os.Open("static/test-sc.png")
	if err != nil {
		panic(err)
	}

	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"metadata tag", "first"}})

	objectId, err := bucket.UploadFromStream("test-sc.png", io.Reader(file), uploadOpts)
	if err != nil {
		panic(err)
	}

	fmt.Println("New file Uploaded with ID: ", objectId)

	r := mux.NewRouter()
	r.HandleFunc("/images/{id}", GetImage).Methods("GET")

	log.Println("Server started on port 8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	fmt.Println(id)

	// Assuming you have initialized your MongoDB client somewhere in your code
	// Replace the client and database variables with your actual MongoDB client and database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:8005"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	database := client.Database("db")
	opts := options.GridFSBucket().SetName("Test Bucket")
	fs, err := gridfs.NewBucket(
		database, opts,
	)
	if err != nil {
		panic(err)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}

	stream, err := fs.OpenDownloadStream(objID)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "image/png")
	if _, err := io.Copy(w, stream); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
