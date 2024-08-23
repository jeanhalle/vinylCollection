package main

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Vinyl represents a vinyl record
type Vinyl struct {
	ID       string `bson:"_id,omitempty"`
	Artist   string `bson:"artist"`
	Album    string `bson:"album"`
	Year     int    `bson:"year"`
	AddedBy  string `bson:"added_by"`
	IsLocked bool   `bson:"locked"`
}

var client *mongo.Client
var collection *mongo.Collection

func main() {
	// MongoDB connection setup
	var err error
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI("mongodb+srv://jeanhalle:UMFSuxzi9ePuSjKd@cluster0.hugwz.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0").SetServerAPIOptions(serverAPI)

	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("vinylstore").Collection("vinyls")

	http.HandleFunc("/", vinylHandler)
	http.HandleFunc("/add", addVinylHandler)
	http.HandleFunc("/remove", removeVinylHandler)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		url := "https://curlmyip.org/"

		// Make the HTTP GET request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error making the request:", err)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the response:", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		// Print the IP address
		w.Write([]byte(string(body)))
		//w.Write([]byte("OK"))
	})

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func vinylHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filterQuery := r.URL.Query().Get("filter")
	filter := bson.D{}

	if filterQuery != "" {
		filterRegex := bson.D{{"$regex", filterQuery}, {"$options", "i"}}
		filter = bson.D{
			{"$or", bson.A{
				bson.D{{"artist", filterRegex}},
				bson.D{{"album", filterRegex}},
				bson.D{{"year", filterRegex}},
			}},
		}
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Unable to fetch records from database", http.StatusInternalServerError)
		return
	}

	var vinyls []Vinyl
	if err = cursor.All(ctx, &vinyls); err != nil {
		http.Error(w, "Unable to decode records from database", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, vinyls)
}

func addVinylHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		artist := r.FormValue("artist")
		album := r.FormValue("album")
		year, _ := strconv.Atoi(r.FormValue("year"))
		addedBy := r.FormValue("added_by")

		newVinyl := Vinyl{Artist: artist, Album: album, Year: year, AddedBy: addedBy, IsLocked: false}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, newVinyl)
		if err != nil {
			http.Error(w, "Unable to insert record into database", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func removeVinylHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, errDelete := collection.DeleteOne(ctx, bson.M{"_id": objID})
		if errDelete != nil {
			http.Error(w, "Unable to delete record from database", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
