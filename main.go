package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"time"
)
import "github.com/gorilla/schema"

type SegmentationRequest struct {
	TextId int64 `schema:textId`
}

type Text struct {
	Content      string
	LanguageCode string
}

type Error struct {
	Message string
}

func handleSegment(w http.ResponseWriter, r *http.Request) {
	var decoder = schema.NewDecoder()
	var segmentationRequest SegmentationRequest
	response := json.NewEncoder(w)
	err := decoder.Decode(&segmentationRequest, r.URL.Query())
	if err != nil {
		log.Println("Error in GET parameters : ", err)
	} else {
		log.Println("GET parameters : ", segmentationRequest)
	}

	datasource := os.Getenv("LWT_DATASOURCE")
	// "username:password@tcp(localhost:3306)/lwt"
	db, err := sql.Open("mysql", datasource)
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	query := fmt.Sprintf("select Content,LanguageCode from texts where id = %d;", segmentationRequest.TextId)
	rows, err := db.Query(query)
	if err != nil {
		response.Encode(Error{Message: "failed to query text."})
		return
	}
	if rows.Next() {
		var text = Text{}
		if err := rows.Scan(&text.Content, &text.LanguageCode); err != nil {
			response.Encode(Error{Message: "Failed to map result."})
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(text)
			return
		}
	} else {
		response.Encode(Error{Message: "Text not found."})
		return
	}
}

func main() {
	http.HandleFunc("/segments", handleSegment)
	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
