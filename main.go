package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Book struct {
	ID string`json:"ID"`
	Title string`json:"Title"`
	Author string`json:"Author"`
}

func EnvLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var db *sql.DB
var err error

func main () {
	EnvLoad()

	db, err = sql.Open("mysql", os.Getenv("USER_NAME") + ":" + os.Getenv("PASSWORD") + "@/sample")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	serviceAccountKeyFilePath, err := filepath.Abs("./serviceAccountKey.json")
	if err != nil {
		panic("Unable to lead serviceAccountKey.json file")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

  //Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Firebase lead error")
	}

	//Firebase Auth
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	token, err := client.CustomToken(context.Background(), "some-uid")
	if err != nil {
					log.Fatalf("error minting custom token: %v\n", err)
	}

	log.Printf("Got custom token: %v\n", token)



	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/books", getAllBooks).Methods("GET")
	router.HandleFunc("/book", createBook).Methods("POST")
	router.HandleFunc("/books/{id}", getOneBook).Methods("GET")
	router.HandleFunc("/books/{id}", updateBook).Methods("PATCH")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getOneBook(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Invalid ID")
	}

	var book Book
	err = db.QueryRow("SELECT * FROM books where id = ?", bookID).Scan( &book.ID, &book.Title, &book.Author)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(book)
	response, _ := json.Marshal(book)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT INTO books(title, author) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	result, err := stmt.Exec( book.Title, book.Author)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}
	log.Printf("ID: %d was created", lastInsertID)
}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	rows , err := db.Query("SELECT * FROM books")
	if err != nil {
		panic(err.Error())
	}

	books := []Book{}
	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ID, &book.Title, &book.Author)
		if err != nil {
			panic(err.Error())
		}
		books = append(books, book)
	}

	fmt.Println(books)
	response, _ := json.Marshal(books)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Fatal(err)
	}

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])

	if err != nil {
		fmt.Println("Invalid ID")
	}

	stmtUpdate, err := db.Prepare("UPDATE books SET title=?, author=? WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmtUpdate.Exec(book.Title,book.Author, bookID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Successfully updated")
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Invalid ID")
	}

	stmtDelete, err := db.Prepare("DELETE FROM books WHERE id=?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmtDelete.Exec(bookID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Successfully deleted")
}