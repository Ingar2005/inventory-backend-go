package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func testConnection(db_user string, db_pass string, db_name string, db_endpoint string) {
	var dsn string = fmt.Sprintf("%s:%s@tcp(%s)/%s", db_user, db_pass, db_endpoint, db_name)
	db, err := sql.Open("mysql", dsn)
	fmt.Println("attempting to connect to the database")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Successfully connected to the database")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database version: ", version)
}

func main() {
	godotenv.Load()
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_endpoint := os.Getenv("DB_ENDPOINT")

	testConnection(db_user, db_pass, db_name, db_endpoint)

}
