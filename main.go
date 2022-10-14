package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	templatesList, err := filepath.Glob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(templatesList)
	templates := template.Must(template.ParseFiles(templatesList...))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			fmt.Println(r.PostForm)
		}
		if r.Method == "GET" {
			if err := templates.ExecuteTemplate(w, "main.html", nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			fmt.Println(r.PostForm)
			postFormValues := r.PostForm
			postFormValues["result"] = []string{
				"first line",
				"second line",
			}
			if err := templates.ExecuteTemplate(w, "search.html", postFormValues); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	fmt.Println("Listen http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
	
}
