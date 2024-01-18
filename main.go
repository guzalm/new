package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

// Product structure represents a product in the store
type Product struct {
	ID    int
	Name  string
	Size  string
	Price float64
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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
						<td>
							<form method="post" action="/delete/{{.ID}}">
								<input type="hidden" name="_method" value="DELETE">
								<button type="submit">Delete</button>
							</form>
						</td>
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

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/delete/"):]
	productID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		fmt.Println("Error deleting from database:", err)
		http.Error(w, "Error deleting from database", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Product deleted with ID: %d\n", productID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	db = initDB()
	defer db.Close()

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/delete/", DeleteHandler)

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
