package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type book struct {
	ID string`json:"ID"`
	Title string`json:"Title"`
	Author string`json:"Author"`
}

var books = []book{
	{
		ID: "1",
		Title: "Readable Code",
		Author: "Dustin Boswell",
	},
	{
		ID: "2",
		Title: "Clean Architecture",
		Author: "Robert C.Martin",
	},
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func main () {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/books", getAllBooks).Methods("GET")
	router.HandleFunc("/book", createBook).Methods("POST")
	router.HandleFunc("/books/{id}", getOneBook).Methods("GET")
	router.HandleFunc("/books/{id}", updateBook).Methods("PATCH")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var newBook book
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
	}

	json.Unmarshal(reqBody, &newBook)
	books = append(books, newBook)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newBook)
}

func getOneBook(w http.ResponseWriter, r *http.Request){
	bookID := mux.Vars(r)["id"]

	for _, singleBook := range books {
		if singleBook.ID == bookID {
			json.NewEncoder(w).Encode(singleBook)
		}
	}
}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(books)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]
	var updateBook book

	reqBody, err:= ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
	}
	json.Unmarshal(reqBody, &updateBook)

	for i , singleBook := range books {
		if singleBook.ID == bookID {
			singleBook.Title = updateBook.Title
			singleBook.Author = updateBook.Author
			books = append(books[:i], singleBook)
			json.NewEncoder(w).Encode(singleBook)
		}
	}
}
func deleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]

	for i , singleBook := range books {
		if singleBook.ID == bookID {
			books = append(books[:i], books[i+1:]...)
			fmt.Fprintf(w, "The book with Id %v has been deleted successfully", bookID)
		}
	}
}