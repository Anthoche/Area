package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Album struct {
	Title  string  `bson:"title"`
	Artist string  `bson:"artist"`
	Price  float32 `bson:"price"`
}

var db *mongo.Database
var collection *mongo.Collection
var ctx context.Context

func main() {
	// URI de connexion MongoDB (modifie selon ton setup)
	uri := "mongodb://localhost:27017"

	// Options du client
	clientOptions := options.Client().ApplyURI(uri)

	// Timeout pour la connexion
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connexion au serveur MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Erreur lors de la connexion à MongoDB: %v", err)
	}

	// Vérification de la connexion
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Impossible de ping MongoDB: %v", err)
	}
	fmt.Println("Connected!")

	db = client.Database("musicdb")
	collection = db.Collection("albums")

	err = initDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database initialized!")

	albums, err := getAlbums()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums: %v\n", albums)

	alb, err := getAlbumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	albumsByArtist, err := getAlbumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albumsByArtist)

	err = addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Added album\n")

	// À la fin : déconnexion propre
	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("Erreur lors de la déconnexion: %v", err)
	}
}

func initDB() error {
	albums := []interface{}{
		Album{"Blue Train", "John Coltrane", 56.99},
		Album{"Giant Steps", "John Coltrane", 63.99},
		Album{"Jeru", "Gerry Mulligan", 17.99},
		Album{"Sarah Vaughan", "Sarah Vaughan", 34.98},
	}

	_, err := collection.InsertMany(ctx, albums)
	if err != nil {
		log.Fatalf("Erreur lors de l'insertion: %v", err)
		return err
	}
	return nil
}

func getAlbums() ([]Album, error) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("getAlbums: %v", err)
	}
	defer cur.Close(ctx)

	var albums []Album
	for cur.Next(ctx) {
		var alb Album
		if err := cur.Decode(&alb); err != nil {
			return nil, fmt.Errorf("getAlbums decode: %v", err)
		}
		albums = append(albums, alb)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("getAlbums cursor: %v", err)
	}
	return albums, nil
}

func getAlbumsByArtist(name string) ([]Album, error) {
	filter := bson.M{"artist": name}

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("getAlbumsByArtist: %v", err)
	}
	defer cur.Close(ctx)

	var albums []Album
	for cur.Next(ctx) {
		var alb Album
		if err := cur.Decode(&alb); err != nil {
			return nil, fmt.Errorf("getAlbumsByArtist decode: %v", err)
		}
		albums = append(albums, alb)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("getAlbumsByArtist cursor: %v", err)
	}
	return albums, nil
}

func getAlbumByID(id int64) (Album, error) {
	var alb Album

	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&alb)
	if err == mongo.ErrNoDocuments {
		return alb, fmt.Errorf("getAlbumByID %d: no such album", id)
	}
	if err != nil {
		return alb, fmt.Errorf("getAlbumByID %d: %v", id, err)
	}
	return alb, nil
}

func addAlbum(alb Album) error {
	// Trouver le max ID existant
	opts := options.FindOne().SetSort(bson.D{{"_id", -1}})

	var last Album
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&last)

	if err != nil {
		return fmt.Errorf("addAlbum: %v", err)
	}

	_, err = collection.InsertOne(ctx, alb)
	if err != nil {
		return fmt.Errorf("addAlbum insert: %v", err)
	}
	return nil
}
