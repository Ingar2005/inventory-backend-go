package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Supplier struct {
	supplierID          int64
	supplier_name       string
	supplier_contact_no string
	lead_time           int64
	monday_deliver      bool
	tuesday_deliver     bool
	wednesday_deliver   bool
	thursday_deliver    bool
	friday_deliver      bool
	saturday_deliver    bool
	sunday_deliver      bool
}
type room struct {
	roomId   int
	roomName string
}

const (
	createSuppliers = `
	CREATE TABLE IF NOT EXISTS suppliers (supplierID int NOT NULL AUTO_INCREMENT,
		supplierName varchar(255) NOT NULL ,
	    supplierContact_no varchar(255) ,
	    leadTime int  NOT NULL  ,
	    mondayDeliver boolean  NOT NULL ,
	    tuesdayDeliver boolean  NOT NULL,
	    wednesdayDeliver boolean  NOT NULL,
	    thursdayDeliver boolean  NOT NULL,
	    fridayDeliver boolean  NOT NULL,
	    saturdayDeliver boolean  NOT NULL,
	    sundayDeliver boolean  NOT NULL,
	    PRIMARY KEY (supplierID));
					`
	genericSupplier = `
		INSERT IGNORE INTO suppliers
		(supplierID, supplierName, leadTime,
			mondayDeliver, tuesdayDeliver, wednesdayDeliver, thursdayDeliver, fridayDeliver, saturdayDeliver, sundayDeliver)
		VALUES (1, 'generic', 0, 1, 1, 1, 1, 1, 1, 1);
		`
	createRooms = `
	CREATE TABLE rooms(
	roomID int NOT NULL AUTO_INCREMENT,
	roomName varchar(255) NOT NULL,
	primary key(roomID));
	`

	genericRoom = `
	INSERT IGNORE INTO rooms(roomID,roomName) VALUES (1,'generic');
	`
)

func main() {
	var err error

	godotenv.Load()
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_endpoint := os.Getenv("DB_ENDPOINT")

	// CRETE A CONNECTION
	db, err := connection(db_user, db_pass, db_name, db_endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// CREATE TABLES AND INSERT GENERIC VALUES
	err = initialiseTables(db)

	//PRINT TABLES
	printSuppliers(db)
	printRooms(db)
}
func connection(db_user string, db_pass string, db_name string, db_endpoint string) (*sql.DB, error) {
	var dsn string = fmt.Sprintf("%s:%s@tcp(%s)/%s", db_user, db_pass, db_endpoint, db_name)
	db, err := sql.Open("mysql", dsn)
	fmt.Println("attempting to connect to the database")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to the database")

	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database version: ", version)

	return db, nil
}
func initialiseTables(db *sql.DB) (err error) {

	_, err = db.Exec(createSuppliers)
	if err != nil {
		return (err)
	}

	_, err = db.Exec(genericSupplier)
	if err != nil {
		return (err)
	}

	_, err = db.Exec(createRooms)
	if err != nil {
		return (err)
	}

	_, err = db.Exec(genericRoom)
	if err != nil {
		return (err)
	}

	return nil
}
func printSuppliers(db *sql.DB) (err error) {
	row, err := db.Query("SELECT * FROM suppliers")
	if err != nil {
		return err
	}
	defer row.Close()

	for row.Next() {
		var contactNo sql.NullString
		data := Supplier{}

		err := row.Scan(&data.supplierID, &data.supplier_name, &contactNo, &data.lead_time,
			&data.monday_deliver, &data.tuesday_deliver, &data.wednesday_deliver, &data.thursday_deliver,
			&data.friday_deliver, &data.saturday_deliver, &data.sunday_deliver)
		if err != nil {
			return err
		}

		if contactNo.Valid {
			data.supplier_contact_no = contactNo.String
		} else {
			data.supplier_contact_no = "N/A"
		}
		fmt.Println(data)
	}
	return nil
}
func printRooms(db *sql.DB) (err error) {

	rows, err := db.Query("SELECT * FROM rooms")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var data room = room{}
		err = rows.Scan(&data.roomId, &data.roomName)
		if err != nil {
			return err
		}

		fmt.Println(data)
	}
	return nil
}
