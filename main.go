package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Vinyl represents a vinyl record
type Vinyl struct {
	Artist string `bson:"artist"`
	Album  string `bson:"album"`
	Year   int    `bson:"year"`
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

		newVinyl := Vinyl{Artist: artist, Album: album, Year: year}

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

