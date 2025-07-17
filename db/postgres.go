package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Init() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=1234 dbname=go_example sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Veritabanı erişilemedi: %v", err)
	}

	log.Printf("Veritabanı bağlantısı başarılı: go_example")
	return db
}
