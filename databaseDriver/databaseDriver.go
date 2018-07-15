package databaseDriver

import (
	"database/sql"
	"fmt"
	"os"
)

func OpenDBConnection(database, username, password string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", username, password, database)) //Onyixj@bs$@/youtubedata")
	return db, err
}

func OpenAndPingDBConnection(database, username, password string) (*sql.DB, error) {
	db, err := OpenDBConnection(database, username, password)
	if err != nil {
		fmt.Println("Error opening DB connection")
	} else {
		// Test DB connection with ping
		err = db.Ping()
		if err != nil {
			fmt.Println("Error pinging DB")
		}
	}

	return db, err

}

func changeDatabase(db *sql.DB, databasename string) {
	_, err := db.Exec("USE " + databasename)
	// fmt.Println(res)
	if err != nil {
		fmt.Printf("Error using database %s\n", databasename)
		os.Exit(1)
	}
}

func tableExists(db *sql.DB, databasename string, tableName string) bool {
	existenceCommand := "SELECT * FROM information_schema.tables WHERE table_schema = '" + databasename + "'AND table_name = '" + tableName + "'"
	rows, err := db.Query(existenceCommand)

	if err != nil {
		fmt.Println("Error checking table existence")
		os.Exit(1)
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false

}

// func getDataFromTable(tableName string) {
// 	var (
// 		id   int
// 		name string
// 	)
// 	rows, err := db.Query("select id, name from users where id = ?", 1)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		err := rows.Scan(&id, &name)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Println(id, name)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func MakeYoutubePlaylistTable(db *sql.DB, databaseName, tableName string) bool {

	changeDatabase(db, databaseName)
	if !tableExists(db, databaseName, tableName) {
		_, err := db.Exec("CREATE TABLE " + tableName + " (videoID VARCHAR(500) NOT NULL, videoTableID INT NOT NULL, " + tableName + "TableID INT NOT NULL AUTO_INCREMENT PRIMARY KEY);")
		if err != nil {
			fmt.Printf("Error making table", tableName)
			return false
			os.Exit(1)
		}
		fmt.Printf("Created table")
		return true
	} else {
		fmt.Printf("Table already exists")
		return false
	}
}

// if err != nil {
// 	fmt.Println("There was an error")
// } else {
// 	fmt.Println("Successful connection")
// }

// defer db.Close()

// Test connection
// err = db.Ping()
// if err != nil {
// 	fmt.Println("Error in ping")
// }

// Prepare statement for inserting data
// stmtIns, err := db.Prepare("INSERT INTO channelData (channelID, channelName) VALUES( ?, ? )") // ? = placeholder
// if err != nil {
// 	panic(err.Error()) // proper error handling instead of panic in your app
// }
// defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

// Preprare read of data

// stmtOut, err := db.Prepare("SELECT squareNumber FROM squarenum WHERE number = ?")
// if err != nil {
// 	panic(err.Error()) // proper error handling instead of panic in your app
// }
// defer stmtOut.Close()

// insertData
// _, err = stmtIns.Exec("testChannelID", "testChannelName")
// if err != nil {
// 	fmt.Println("Error on insertion") // proper error handling instead of panic in your app
// }
