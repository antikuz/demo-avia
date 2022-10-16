package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/antikuz/demo-avia/internal/db"
	"github.com/antikuz/demo-avia/internal/handlers"
	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/internal/processors"
	"github.com/antikuz/demo-avia/pkg/logging"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
)

var sessions = map[string]*models.Session{}

func main() {
	logger := logging.GetLogger()
	ctx, _ := context.WithCancel(context.Background())
	
	// For testing purpose leave cred here
	user := "postgres"
	password := "chunga"
	host := "localhost"
	port := "5432"
	dbname := "demo"
	
	connectstring := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)
	db_storage, err := pgxpool.Connect(ctx, connectstring)
	if err != nil {
		logger.Fatalln(err)
	}
	defer db_storage.Close()
	
	storage := db.NewStorage(db_storage, logger)
	processor := processors.NewStorageProcessor(storage, logger)

	templatesList, err := filepath.Glob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	logger.Debugln(templatesList)
	templates := template.Must(template.ParseFiles(templatesList...))
	handler := handlers.NewHandler(templates, processor, sessions, logger)
	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("static"))
	handler.Register(router)

	fmt.Println("Listen http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
