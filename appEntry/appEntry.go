package main

import (
	"fmt"
	"media-tracker/databaseDriver"
	"media-tracker/youtubecontent"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const db_driver string = "mysql"
const db_username string = "root"
const db_password string = "Onyixj@bs$"

// func dummy(i interface{}) {
// 	fmt.Println(reflect.TypeOf(i))
// }

// func dummy2() interface{} {
// 	return "hi"
// }

// func dummy3() interface{} {
// 	return true
// }

func main() {
	fmt.Println("starting")

	db, err := databaseDriver.OpenAndPingDBConnection(db_driver, db_username, db_password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("About to do defer\n")
		defer db.Close()
	}

	DBConn := databaseDriver.NewDataBaseConnectionObject(db)

	youtubecontent.InitializeDemo2(DBConn)

	// dummy(dummy2())
	// dummy(dummy3())

	// columnNames := []string{"channelID", "channelName", "tableId"}
	// var dataValues = make(map[string]string)

	// dataValues["channelName"] = "testChannelName"

	// returnedData := databaseDriver.GetDataFromTable(db, "youtubedata", "channelData", columnNames, dataValues)
	//
	// fmt.Println(returnedData)
	// fmt.Print("\n\n\n\n\n")

	// var num int64
	// num, _ = strconv.ParseInt(string(returnedData[0]["tableId"].([]uint8)), 10, 0)
	// fmt.Println(num)

	// Ready to communicate with DB
	// ret := databaseDriver.MakeYoutubePlaylistTable(db, "youtubedata", "relatedVideos")
	// fmt.Printf("Made table: %t\n", ret)

	// youtubecontent.Initialize(db)

}
