package main

import (
	"database/sql"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Supplier struct {
	SupplierID        int64  `json:supplierID`
	SupplierName      string `json:supplierName`
	SupplierContactNo string `json:supplierContactNo` // IF NULLL WILL BE VALUE N/A
	LeadTime          int64  `json:leadTime`
	MondayDeliver     bool   `json:mondayDeliver`
	TuesdayDeliver    bool   `json:tuesdayDeliver`
	WednesdayDeliver  bool   `json:wednesdayDeliver`
	ThursdayDeliver   bool   `json:thursdayDeliver`
	FridayDeliver     bool   `json:fridayDeliver`
	SaturdayDeliver   bool   `json:saturdayDeliver`
	SundayDeliver     bool   `json:sundayDeliver`
}
type Room struct {
	RoomId   int    `json:roomId`
	RoomName string `json:roomName`
}
type Stock struct {
	StockID       int     `json:stockID`
	ItemName      string  `json:itemName`
	Level         float64 `json:level`
	RoomID        int     `json:roomID`
	SupplierID    int     `json:supplierID`
	IncidentLevel float64 `json:incidentLevel`
	LastLogID     int     `json:lastLogID` // IF NONE WILL BE VALUE 0
}
type LogRow struct {
	LogID        int            `json:logID`
	StockID      int            `json:stockID`
	Differance   float64        `json:differance`
	TotalAfter   float64        `json:totalAfter`
	IncidentTime mysql.NullTime `json:incidentTime`
	Daily        bool           `json:daily`
}
type FullStock struct {
	StockID       int            `json:stockID`
	ItemName      string         `json:itemName`
	Level         float64        `json:level`
	RoomID        int            `json:roomID`
	Room          string         `json:room`
	SupplierID    int            `json:supplierID`
	Supplier      string         `json:supplier`
	IncidentLevel float64        `json:incidentLevel`
	LastLogID     int            `json:lastLogID`
	LastChanged   mysql.NullTime `json:lastChange`
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
	CREATE TABLE IF NOT EXISTS rooms(
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

var db *sql.DB

func main() {
	var err error

	godotenv.Load()
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_endpoint := os.Getenv("DB_ENDPOINT")
	port := ":5000"
	// CRETE A CONNECTION
	db, err = connection(db_user, db_pass, db_name, db_endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// _, err = db.Exec("DROP TABLE IF EXISTS logs, stock, rooms, suppliers;")
	err = initialiseTables()
	if err != nil {
		log.Fatal(err)
	}
	// err = addStock("cheese", 3, 2, 1, 4)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// CREATE server
	http.HandleFunc("/logs/", logs)
	http.HandleFunc("/suppliers/", suppliers)
	http.HandleFunc("/rooms/", rooms)
	http.HandleFunc("/stock/", stock)
	http.HandleFunc("/fullStock/", stockFull)

	http.HandleFunc("/", root)
	fmt.Printf("attempting to connect on port%v \n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
func root(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
	} else {

		fmt.Fprintf(w, "Welcome to the HomePage!")
		fmt.Println("Endpoint Hit: root")
	}
}
func logs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var res []LogRow
		var err error
		fmt.Println("Endpoint Hit: logs GET")
		res, err = getLogs()
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(res)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func suppliers(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		fmt.Println("Endpoint Hit: suppliers GET")
		res, err := getSuppliers()
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(res)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}
func rooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, PATCH, OPTIONS, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		fmt.Println("Endpoint Hit: rooms OPTIONS")
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		fmt.Println("Endpoint Hit: rooms GET")
		res, err := getRooms()
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(res)
	case http.MethodDelete:
		fmt.Println("Endpoint Hit: rooms DELETE")

		id := strings.TrimPrefix(r.URL.Path, "/rooms/")
		idnum, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}
		err = deleteRoom(idnum)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data deleated sucesfuly"))

	case http.MethodPost:
		fmt.Println("Endpoint Hit: rooms POST")
		var data Room
		data.RoomId = 0
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		err = addRoom(data.RoomName)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data written sucesfuly"))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}
func stock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		fmt.Println("Endpoint Hit: stock OPTIONS")
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		fmt.Println("Endpoint Hit: stock GET")

		res, err := getStock()

		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(res)
	case http.MethodDelete:
		fmt.Println("Endpoint Hit: stock DELETE")

		id := strings.TrimPrefix(r.URL.Path, "/stock/")
		idnum, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}
		err = deleteStock(idnum)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data deleated sucesfuly"))
	case http.MethodPost:
		fmt.Println("Endpoint Hit: stock POST")
		var data Stock
		data.SupplierID = 1
		data.LastLogID = 0
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(data)
		err = addStock(data.ItemName, data.Level, data.RoomID, data.SupplierID, data.IncidentLevel)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data added sucesfuly"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	}
}
func stockFull(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		fmt.Println("Endpoint Hit: stock_full OPTIONS")
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		var res []FullStock
		var err error

		fmt.Println("Endpoint Hit: stock_full GET")

		id := strings.TrimPrefix(r.URL.Path, "/fullStock/")
		idnum, err := strconv.Atoi(id)

		if id != "" && err == nil {
			res, err = getFullStockById(idnum)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			res, err = getStockFull()
			if err != nil {
				log.Fatal(err)
			}
		}
		json.NewEncoder(w).Encode(res)
	case http.MethodPatch:

		fmt.Println("Endpoint Hit: stock PATCH")
		var data FullStock
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		err = updateFullStockLevel(data)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data received successfully"))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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
func initialiseTables() (err error) {

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

// CREATE

func addStock(name string, level float64, roomID int, supplierID int, incident float64) (err error) {
	query := "INSERT INTO stock(itemName,level,roomID,supplierID,incidentLevel) VALUES (?,?,?,?,?)"

	_, err = db.Exec(query, name, level, roomID, supplierID, incident)
	if err != nil {
		return err
	}

	return nil
}
func addRoom(roomName string) (err error) {
	query := "INSERT INTO rooms(roomName) VALUES (?)"

	_, err = db.Exec(query, roomName)
	if err != nil {
		return err
	}
	return nil
}

// GET

func getSuppliers() (res []Supplier, err error) {
	row, err := db.Query("SELECT * FROM suppliers")
	if err != nil {
		return res, err
	}
	defer row.Close()

	var data Supplier
	var contactNo sql.NullString
	for row.Next() {

		err := row.Scan(&data.SupplierID, &data.SupplierName, &contactNo, &data.LeadTime,
			&data.MondayDeliver, &data.TuesdayDeliver, &data.WednesdayDeliver, &data.ThursdayDeliver, &data.FridayDeliver, &data.SaturdayDeliver, &data.SundayDeliver)
		if err != nil {
			return res, err
		}

		if contactNo.Valid {
			data.SupplierContactNo = contactNo.String
		} else {
			data.SupplierContactNo = "N/A"
		}
		res = append(res, data)
	}
	return res, nil
}
func getRooms() (res []Room, err error) {

	rows, err := db.Query("SELECT * FROM rooms")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data Room
	for rows.Next() {
		err = rows.Scan(&data.RoomId, &data.RoomName)
		if err != nil {
			return res, err
		}
		res = append(res, data)
	}
	return res, nil
}
func getStock() (res []Stock, err error) {
	rows, err := db.Query("SELECT * FROM stock")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data Stock
	for rows.Next() {
		var log sql.NullInt64
		rows.Scan(&data.StockID, &data.ItemName, &data.Level, &data.RoomID, &data.SupplierID, &data.IncidentLevel, &log)

		if log.Valid {
			data.LastLogID = int(log.Int64)
		} else {
			data.LastLogID = 0
		}

		res = append(res, data)
	}
	return res, nil
}
func getStockFull() (res []FullStock, err error) {
	rows, err := db.Query(`
		SELECT
		    stock.stockID,
		    stock.itemName,
		    stock.level,
			rooms.roomID,
		    rooms.roomName AS room,
			suppliers.supplierID,
		    suppliers.supplierName AS supplier,
		    stock.incidentLevel,
		    stock.lastLogID,
		    logs.incidentTime AS "last changed"
		FROM
		    stock
		JOIN
		    rooms ON stock.roomID = rooms.roomID
		JOIN
		    suppliers ON stock.supplierID = suppliers.supplierID
		LEFT JOIN
		    logs ON stock.lastLogID = logs.logID;`)

	defer rows.Close()
	if err != nil {
		return res, err
	}

	var data FullStock
	var log sql.NullInt64
	for rows.Next() {
		err = rows.Scan(&data.StockID, &data.ItemName, &data.Level, &data.RoomID, &data.Room, &data.SupplierID, &data.Supplier, &data.IncidentLevel, &log, &data.LastChanged)
		if err != nil {
			return res, err
		}
		if log.Valid {
			data.LastLogID = int(log.Int64)

		} else {
			data.LastLogID = 0
		}
		res = append(res, data)
	}

	return res, err
}
func getLogs() (res []LogRow, err error) {

	rows, err := db.Query("SELECT * FROM logs")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var data LogRow

	for rows.Next() {
		err = rows.Scan(&data.LogID, &data.StockID, &data.Differance, &data.TotalAfter, &data.IncidentTime, &data.Daily)
		if err != nil {
			return res, err
		}
		res = append(res, data)
	}
	return res, nil
}
func getFullStockById(id int) (res []FullStock, err error) {
	row, err := db.Query(`
		SELECT
		    stock.stockID,
		    stock.itemName,
		    stock.level,
			rooms.roomID,
		    rooms.roomName AS room,
			suppliers.supplierID,
		    suppliers.supplierName AS supplier,
		    stock.incidentLevel,
		    stock.lastLogID,
		    logs.incidentTime AS "last changed"
		FROM
		    stock
		JOIN
		    rooms ON stock.roomID = rooms.roomID
		JOIN
		    suppliers ON stock.supplierID = suppliers.supplierID
		LEFT JOIN
		    logs ON stock.lastLogID = logs.logID
		WHERE
		    stock.stockID = ?;
		`, id)
	defer row.Close()

	if err != nil {
		return res, err
	}
	var data FullStock
	var log sql.NullInt64
	if row.Next() {
		err = row.Scan(&data.StockID, &data.ItemName, &data.Level, &data.RoomID, &data.Room, &data.SupplierID, &data.Supplier, &data.IncidentLevel, &log, &data.LastChanged)
		if err != nil {
			return res, err
		}
		if log.Valid {
			data.LastLogID = int(log.Int64)
		} else {
			data.LastLogID = 0
		}
		res = append(res, data)
	}
	return res, nil
}

// UPDATE

func updateFullStockLevel(data FullStock) (err error) {
	const selectOldLevel = `SELECT level FROM stock WHERE stockID=? LIMIT 1 FOR UPDATE`
	const insertLog = `INSERT INTO logs(stockID,differance,totalAfter,incidentTime,daily) VALUES (?,?,?,NOW(),0);`
	const selectLog = `SELECT LAST_INSERT_ID();`
	const updateQuery = `UPDATE stock SET level=?, lastLogID=? WHERE stockID=?;`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stockId := data.StockID
	var oldlevel float64
	err = tx.QueryRow(selectOldLevel, stockId).Scan(&oldlevel)
	if err != nil {
		return err
	}
	stockLevel := data.Level
	differance := stockLevel - oldlevel

	_, err = tx.Exec(insertLog, stockId, differance, stockLevel)
	if err != nil {
		return err
	}

	var logID int
	err = tx.QueryRow(selectLog).Scan(&logID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(updateQuery, stockLevel, logID, stockId)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DELETE

func deleteStock(id int) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM logs WHERE stockID=?", id)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM stock WHERE stockID=?", id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}
func deleteRoom(id int) (err error) {
	_, err = db.Exec("DELETE FROM rooms WHERE roomID=?", id)
	if err != nil {
		return err
	}
	return nil
}
