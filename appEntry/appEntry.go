package main

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const db_driver string = "mysql"
const db_username string = "root"
const db_password string = "Onyixj@bs$"

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

	// Ready to communicate with DB
	ret := databaseDriver.MakeYoutubePlaylistTable(db, "youtubedata", "relatedVideos")
	fmt.Printf("Made table: %t\n", ret)
	// youtubecontent.DemoMain()

}
