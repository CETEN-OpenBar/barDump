package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ExportData(outputDir string) error {
	sourceURI := os.Getenv("MONGODB_URI")
	sourceDBName := os.Getenv("MONGODB_DB_NAME")
	if sourceDBName == "" {
		sourceDBName = "bar"
	}

	if sourceURI == "" {
		return fmt.Errorf("Erreur : La variable d'environnement MONGODB_URI n'est pas définie")
	}

	log.Printf("--- Connexion à la source : %s ---", sourceDBName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(sourceURI))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	db := client.Database(sourceDBName)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	collections := []string{"accounts", "transactions"}
	for _, colName := range collections {
		log.Printf("Export de la collection : %s...", colName)
		collection := db.Collection(colName)

		cur, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			return err
		}

		var results []bson.M
		if err = cur.All(context.Background(), &results); err != nil {
			return err
		}

		filePath := filepath.Join(outputDir, fmt.Sprintf("%s.json", colName))
		tmpFilePath := filePath + ".tmp"

		file, err := os.Create(tmpFilePath)
		if err != nil {
			return err
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(results); err != nil {
			file.Close()
			return err
		}
		file.Close()

		if err := os.Rename(tmpFilePath, filePath); err != nil {
			return err
		}
	}

	log.Println("--- Export terminé avec succès ---")
	return nil
}
