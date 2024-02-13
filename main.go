// package main

// import (
// 	// "context"
// 	// "fmt"
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"

// 	// "os"

// 	"github.com/gorilla/mux"
// 	// "go.mongodb.org/mongo-driver/bson"
// 	// "go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/gridfs"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// 	// "go.mongodb.org/mongo-driver/mongo/options"
// )

// func main() {

// 	MONGODB_URI := "mongodb://localhost:8005/"
// 	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MONGODB_URI))
// 	if err != nil {
// 		log.Fatal(err)
// 		panic(err)
// 	}

// 	db := client.Database("db")
// 	opts := options.GridFSBucket().SetName("Test Bucket")
// 	bucket, err := gridfs.NewBucket(db, opts)
// 	if err != nil {
// 		panic(err)
// 	}

// 	file, err := os.Open("static/test-sc.png")
// 	if err != nil {
// 		panic(err)
// 	}

// 	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"metadata tag", "first"}})

// 	objectId, err := bucket.UploadFromStream("test-sc.png", io.Reader(file), uploadOpts)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("New file Uploaded with ID: ", objectId)

// 	r := mux.NewRouter()
// 	r.HandleFunc("/images/{id}", GetImage).Methods("GET")

// 	log.Println("Server started on port 8083")
// 	log.Fatal(http.ListenAndServe(":8083", r))
// }

// func GetImage(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	id := params["id"]

// 	fmt.Println(id)

// 	// Assuming you have initialized your MongoDB client somewhere in your code
// 	// Replace the client and database variables with your actual MongoDB client and database
// 	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:8005"))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer client.Disconnect(context.Background())

// 	database := client.Database("db")
// 	opts := options.GridFSBucket().SetName("Test Bucket")
// 	fs, err := gridfs.NewBucket(
// 		database, opts,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	objID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		panic(err)
// 	}

// 	stream, err := fs.OpenDownloadStream(objID)
// 	if err != nil {
// 		panic(err)
// 	}

// 	w.Header().Set("Content-Type", "image/png")
// 	if _, err := io.Copy(w, stream); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

package main

import (
	"context"
	// "fmt"
	"io"
	"log"
	"net/http"

	// "os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	MONGODB_URI := "mongodb://localhost:8005/"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MONGODB_URI))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("db")
	opts := options.GridFSBucket().SetName("Test Bucket")
	bucket, err := gridfs.NewBucket(db, opts)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// Handle POST requests to upload images
	r.POST("/images", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image not found in form data"})
			return
		}
		defer file.Close()

		uploadStream, err := bucket.OpenUploadStream(header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open GridFS upload stream"})
			return
		}
		defer uploadStream.Close()

		obj, err := io.Copy(uploadStream, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Image uploaded successfully",
			"obj": obj})
	})

	// Handle GET requests to retrieve images
	r.GET("/images/:id", func(c *gin.Context) {
		id := c.Param("id")

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
			return
		}

		stream, err := bucket.OpenDownloadStream(objID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
			return
		}
		defer stream.Close()

		c.Writer.Header().Set("Content-Type", "image/png")
		_, err = io.Copy(c.Writer, stream)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send image"})
			return
		}
	})

	log.Println("Server started on port 8083")
	log.Fatal(r.Run(":8083"))
}
