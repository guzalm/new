package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

// Product structure represents a product in the store
type Product struct {
	ID    int
	Name  string
	Size  string
	Price float64
}

// PurchaseRequest structure represents data for the POST request to buy a product
type PurchaseRequest struct {
	ProductID int `json:"product_id"`
	// Add other fields if necessary
}

var db *sql.DB

func initDB() *sql.DB {
	// Replace with your actual PostgreSQL connection details
	connStr := "user=postgres password=rayana2015 dbname=newdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error opening database connection:", err)
		panic(err)
	}

	// Ensure the database connection is successful
	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		panic(err)
	}

	fmt.Println("Connected to the database")

	return db
}

// fetchProductsFromDB retrieves products from the PostgreSQL database
func fetchProductsFromDB() ([]Product, error) {
	var products []Product

	rows, err := db.Query("SELECT id, name, size, price FROM products")
	if err != nil {
		fmt.Println("Error fetching products from the database:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Size, &p.Price); err != nil {
			fmt.Println("Error scanning product row:", err)
			continue
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating over product rows:", err)
		return nil, err
	}

	return products, nil
}

// IndexHandler handles the GET request on the main page of the store
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch products from the database
	products, err := fetchProductsFromDB()
	if err != nil {
		http.Error(w, "Error fetching products from the database", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Online Store</title>
		</head>
		<body>
			<h1>Welcome to the Online Store!</h1>
			<h2>Products:</h2>
			<table border="1">
				<tr>
					<th>ID</th>
					<th>Name</th>
					<th>Size</th>
					<th>Price</th>
					<th>Action</th>
				</tr>
				{{range .}}
					<tr>
						<td>{{.ID}}</td>
						<td>{{.Name}}</td>
						<td>{{.Size}}</td>
						<td>${{.Price}}</td>
						<td><form method="post" action="/buy/{{.ID}}"><input type="submit" value="Buy"></form></td>
					</tr>
				{{end}}
			</table>
			<a href="/add-product">Add Product</a>
		</body>
		</html>
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, products)
}

// BuyHandler handles the POST request to buy a product
func BuyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var purchaseRequest PurchaseRequest
	if err := decoder.Decode(&purchaseRequest); err != nil {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	// Process data and send a response
	// ...

	fmt.Fprintf(w, "Product with ID %d successfully purchased!", purchaseRequest.ProductID)
}

// AddProductHandler displays the page for adding a new product
func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("addProduct").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Add Product</title>
		</head>
		<body>
			<h1>Add a new product</h1>
			<form method="post" action="/add-product-post">
				<label for="name">Name:</label>
				<input type="text" name="name" required><br>
				<label for="size">Size:</label>
				<input type="text" name="size" required><br>
				<label for="price">Price:</label>
				<input type="number" name="price" step="0.01" required><br>
				<input type="submit" value="Add Product">
			</form>
			<a href="/">Back to Home</a>
		</body>
		</html>
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// AddProductPostHandler handles the POST request to add a new product
func AddProductPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Insert the new product into the PostgreSQL database
	_, err := db.Exec("INSERT INTO products (name, size, price) VALUES ($1, $2, $3)",
		r.FormValue("name"), r.FormValue("size"), r.FormValue("price"))
	if err != nil {
		fmt.Println("Error inserting into database:", err)
		http.Error(w, "Error inserting into database", http.StatusInternalServerError)
		return
	}

	// Log the addition of the new product
	fmt.Printf("New product added: Name=%s, Size=%s, Price=%s\n", r.FormValue("name"), r.FormValue("size"), r.FormValue("price"))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	db = initDB()
	defer db.Close()

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/buy/", BuyHandler)
	http.HandleFunc("/add-product", AddProductHandler)
	http.HandleFunc("/add-product-post", AddProductPostHandler)

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
