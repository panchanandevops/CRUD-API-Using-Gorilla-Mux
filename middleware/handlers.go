// internal/middleware/middleware.go
package middleware

import (
	"crud-api/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

// response is a struct for JSON responses
type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

// createConnection establishes a connection to the PostgreSQL database
func createConnection() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

// CreateStock handles the creation of a new stock in the database
func CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock models.Stock

	// Decode the JSON request body into the stock struct
	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable to decode the request body. %v", err)
	}

	// Call the insertStock function to insert the stock into the database
	insertID := insertStock(stock)

	// Prepare the response object
	res := response{
		ID:      insertID,
		Message: "Stock created successfully",
	}

	// Send the response as JSON
	json.NewEncoder(w).Encode(res)
}

// GetStock retrieves a single stock by its ID from the database
func GetStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert the string into int. %v", err)
	}

	// Call the getStock function to retrieve a single stock from the database
	stock, err := getStock(int64(id))
	if err != nil {
		log.Fatalf("Unable to get stock. %v", err)
	}

	// Send the stock as JSON response
	json.NewEncoder(w).Encode(stock)
}

// GetAllStock retrieves all stocks from the database
func GetAllStock(w http.ResponseWriter, r *http.Request) {
	// Call the getAllStocks function to retrieve all stocks from the database
	stocks, err := getAllStocks()
	if err != nil {
		log.Fatalf("Unable to get all stock. %v", err)
	}

	// Send all the stocks as a JSON response
	json.NewEncoder(w).Encode(stocks)
}

// UpdateStock updates the details of a stock in the database
func UpdateStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert the string into int. %v", err)
	}

	var stock models.Stock

	// Decode the JSON request body into the stock struct
	err = json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable to decode the request body. %v", err)
	}

	// Call the updateStock function to update the stock in the database
	updatedRows := updateStock(int64(id), stock)

	// Prepare the message string
	msg := fmt.Sprintf("Stock updated successfully. Total rows/record affected %v", updatedRows)

	// Prepare the response object
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// Send the response as JSON
	json.NewEncoder(w).Encode(res)
}

// DeleteStock deletes a stock's detail from the database
func DeleteStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert the string into int. %v", err)
	}

	// Call the deleteStock function to delete the stock from the database
	deletedRows := deleteStock(int64(id))

	// Prepare the message string
	msg := fmt.Sprintf("Stock updated successfully. Total rows/record affected %v", deletedRows)

	// Prepare the response object
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// Send the response as JSON
	json.NewEncoder(w).Encode(res)
}

// insertStock inserts one stock into the DB
func insertStock(stock models.Stock) int64 {
	db := createConnection()
	defer db.Close()

	// SQL statement for inserting a stock and returning its ID
	sqlStatement := `INSERT INTO stocks (name, price, company) VALUES ($1, $2, $3) RETURNING stockid`

	var id int64

	// Execute the SQL statement and scan the inserted ID
	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)
	return id
}

// getStock retrieves one stock from the DB by its stockid
func getStock(id int64) (models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stock models.Stock

	// SQL statement for selecting a stock by its ID
	sqlStatement := `SELECT * FROM stocks WHERE stockid=$1`

	// Execute the SQL statement and unmarshal the row into the stock struct
	row := db.QueryRow(sqlStatement, id)
	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return stock, nil
	case nil:
		return stock, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// Return an empty stock on error
	return stock, err
}

// getAllStocks retrieves all stocks from the DB
func getAllStocks() ([]models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stocks []models.Stock

	// SQL statement for selecting all stocks
	sqlStatement := `SELECT * FROM stocks`

	// Execute the SQL statement and iterate over the rows, unmarshalling each into a stock struct
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)
		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}
		stocks = append(stocks, stock)
	}

	// Return empty stock on error
	return stocks, err
}

// updateStock updates a stock in the DB
func updateStock(id int64, stock models.Stock) int64 {
	db := createConnection()
	defer db.Close()

	// SQL statement for updating a stock by its ID
	sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`

	// Execute the SQL statement and get the number of affected rows
	res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// Check how many rows were affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)
	return rowsAffected
}

// deleteStock deletes a stock from the DB
func deleteStock(id int64) int64 {
	db := createConnection()
	defer db.Close()

	// SQL statement for deleting a stock by its ID
	sqlStatement := `DELETE FROM stocks WHERE stockid=$1`

	// Execute the SQL statement and get the number of affected rows
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// Check how many rows were affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)
	return rowsAffected
}
