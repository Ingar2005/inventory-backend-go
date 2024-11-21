package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type supplier struct {
	supplierID          int64
	supplier_name       string
	supplier_contact_no string // IF NULLL WILL BE VALUE N/A
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
type stock struct {
	stockID       int
	itemName      string
	level         float64
	roomID        int
	supplierID    int
	incidentLevel float64
	lastLogID     int // IF NONE WILL BE VALUE 0
}
type logRow struct {
	logID        int
	stockID      int
	differance   float64
	totalAfter   float64
	incidentTime mysql.NullTime
	daily        bool
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

	createStock = `
	CREATE TABLE IF NOT EXISTS stock(
		stockID int NOT NULL AUTO_INCREMENT,
		itemName varchar(255) NOT NULL,
		level float NOT NULL,
		roomID int NOT NULL,
		supplierID int NOT NULL,
		incidentLevel float,
		lastLogID int,

		PRIMARY KEY (stockID),
		FOREIGN KEY (roomID) REFERENCES rooms(roomID),
		FOREIGN KEY (supplierID) REFERENCES suppliers(supplierID));
	`

	createLogs = `
	CREATE TABLE IF NOT EXISTS logs (
    logID int NOT NULL AUTO_INCREMENT,
    stockID int NOT NULL,
    differance float NOT NULL,
    totalAfter float NOT NULL,
    incidentTime datetime NOT NULL,
    daily boolean NOT NULL,
    PRIMARY KEY (logID),
    FOREIGN KEY (stockID) REFERENCES stock(stockID));
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
	// _, err = db.Exec("DROP TABLE IF EXISTS logs, stock, rooms, suppliers;")
	err = initialiseTables(db)

	//PRINT TABLES
	err = printSuppliers(db)
	if err != nil {
		log.Fatal(err)
	}
	err = printRooms(db)
	if err != nil {
		log.Fatal(err)
	}
	res, err := getStock(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	err = printLogs(db)
	if err != nil {
		log.Fatal(err)
	}

	// TEST VALUES
	// _, err = db.Exec("INSERT INTO stock(itemName, level, roomID, supplierID, incidentLevel) VALUES (?, ?, ?, ?, ?)", "test", 10, 1, 1, 0)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = db.Exec("INSERT INTO logs(stockID, differance, totalAfter, incidentTime, daily) VALUES (?, ?, ?, NOW(), ?)", 1, -1, 9, 0)
	// if err != nil {
	// 	log.Fatal(err)
	// }
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

	_, err = db.Exec(createStock)
	if err != nil {
		return err
	}

	_, err = db.Exec(createLogs)
	if err != nil {
		return err
	}

	return nil
}
func printSuppliers(db *sql.DB) (err error) {
	row, err := db.Query("SELECT * FROM suppliers")
	if err != nil {
		return err
	}
	defer row.Close()

	fmt.Println("Suppliers: ")

	var data supplier
	var contactNo sql.NullString
	for row.Next() {

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
func getRooms(db *sql.DB) (res []room, err error) {

	rows, err := db.Query("SELECT * FROM rooms")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data room
	for rows.Next() {
		err = rows.Scan(&data.roomId, &data.roomName)
		if err != nil {
			return res, err
		}

	}
	return res, nil
}
func getStock(db *sql.DB) (res []stock, err error) {
	rows, err := db.Query("SELECT * FROM stock")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data stock
	for rows.Next() {
		var log sql.NullInt64
		rows.Scan(&data.stockID, &data.itemName, &data.level, &data.roomID, &data.supplierID, &data.incidentLevel, &log)

		if log.Valid {
			data.lastLogID = int(log.Int64)
		} else {
			data.lastLogID = 0
		}
		res = append(res, data)
	}
	return res, nil
}
func getLogs(db *sql.DB) (res []logRow, err error) {

	rows, err := db.Query("SELECT * FROM logs")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data logRow
	for rows.Next() {
		err = rows.Scan(&data.logID, &data.stockID, &data.differance, &data.totalAfter, &data.incidentTime, &data.daily)
		if err != nil {
			return res, err
		}

		res = append(res, data)
	}
	return res, nil
}
