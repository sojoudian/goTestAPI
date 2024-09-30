package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Book represents a simple book structure
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// In-memory book storage and a mutex for concurrent access
var (
	books = make(map[int]Book)
	mu    sync.Mutex
	nextID = 1
)

// Helper to send JSON responses
func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// CreateBook handles POST /books for creating a new book
func CreateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	book.ID = nextID
	books[book.ID] = book
	nextID++
	mu.Unlock()

	sendJSON(w, book, http.StatusCreated)
}

// GetBooks handles GET /books for listing all books
func GetBooks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var bookList []Book
	for _, book := range books {
		bookList = append(bookList, book)
	}

	sendJSON(w, bookList, http.StatusOK)
}

// GetBook handles GET /books/{id} for retrieving a specific book by ID
func GetBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/books/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	book, found := books[id]
	if !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	sendJSON(w, book, http.StatusOK)
}

// UpdateBook handles PUT /books/{id} for updating an existing book
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/books/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	var updatedBook Book
	if err := json.NewDecoder(r.Body).Decode(&updatedBook); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, found := books[id]
	if !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	updatedBook.ID = id
	books[id] = updatedBook

	sendJSON(w, updatedBook, http.StatusOK)
}

// DeleteBook handles DELETE /books/{id} for deleting a book
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/books/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, found := books[id]
	if !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(books, id)
	w.WriteHeader(http.StatusNoContent)
}

// Route handling for CRUD operations
func main() {
	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetBooks(w, r)
		case http.MethodPost:
			CreateBook(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetBook(w, r)
		case http.MethodPut:
			UpdateBook(w, r)
		case http.MethodDelete:
			DeleteBook(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
