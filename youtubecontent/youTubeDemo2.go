package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"regexp"
	"strconv"

	"google.golang.org/api/youtube/v3"
)

// Contains *youtube.Service object and implements getServiceConnection() to return this object
type ServiceConnectionObject struct {
	service *youtube.Service
}

func (sco *ServiceConnectionObject) getServiceConnection() *youtube.Service {
	return sco.service
}

type APIToDataBaseControllerObjectI interface {
	GeneralYouTubeAPIObjectI
	databaseDriver.DataBaseTableRequestObjectI
	databaseDriver.DataInsertObjectI
	databaseDriver.DataBaseTableObjectI
	databaseDriver.DataBaseConnectionObjectI
}

type GeneralYouTubeAPIObjectI interface {
	RequestYouTubeAPIObjectFromId(string, map[string]interface{})
	// ScrapeObjectThenAddToDatabase(string)
}

type GeneralAPIToDataBaseControllerObject struct {
	databaseDriver.DataInsertObject
	ServiceConnectionObject
}

func (genObj *GeneralAPIToDataBaseControllerObject) RequestYouTubeAPIObjectFromId(id string) {
	// Do nothing
}

func ScrapeObjectThenAddToDatabase(genObj APIToDataBaseControllerObjectI, id string) {
	genObj.RequestYouTubeAPIObjectFromId(id, map[string]interface{}{})
	genObj.CreateDataMap()
	_, err := genObj.PrepareAndExecute()
	if err != nil {
		fmt.Println("Error in adding object to database")
		fmt.Println(err)
	}
}

func addObjectToDatabaseTableFromYouTubeID(db databaseDriver.DataBaseConnectionObject, conn ServiceConnectionObject, dbName, objectID string, objectType interface{}) int64 {
	var obj APIToDataBaseControllerObjectI
	var columnName string

	switch objectType.(type) {
	case *youtube.Channel:
		obj = makeChannelInsertObject(db, conn, dbName)
		columnName = "CHANNELID"
	case *youtube.Video:
		obj = makeVideoInsertObject(db, conn, dbName)
		columnName = "VIDEOID"
	default:
		return -1
	}
	databaseDriver.ChangeDatabase(obj.GetDBConnection(), obj.GetDatabaseName())

	// obj = obj.(*channelInsertObject)

	// obj.CellExists()

	// if !StringDataInDatabase(conn, databaseName, tableName, "channelId", channelID) {

	if !obj.CellExists(columnName, objectID) {
		fmt.Println("Cell doesnt exist, about to connect to API")
		obj.RequestYouTubeAPIObjectFromId(objectID, map[string]interface{}{})
		obj.CreateDataMap()
		obj.PrepareRequestStatement()
		_, err := obj.RunDBExecution()
		if err != nil {
			fmt.Println("Error in adding object to database")
			fmt.Println(err)
		}
	}

	ret := obj.GetColumnCellValues("tableId", map[string][]string{columnName: []string{objectID}})
	// by construction ret should be [int64]
	fmt.Println(ret)
	if len(ret) != 1 || (len(ret) == 1 && ret[0] == nil) {
		fmt.Println(ret)
		fmt.Println("Problem getting unique tableId")
		return -1
	} else {
		retInt, _ := strconv.ParseInt(string(ret[0].([]uint8)), 10, 0)
		return retInt
	}

}

func addObjectsToDatabaseTable(db databaseDriver.DataBaseConnectionObject, conn ServiceConnectionObject, dbName string, objectTypeToIds map[interface{}]interface{}) map[interface{}][]int64 {
	// db := conn.db
	var returnTableIdsMap map[interface{}][]int64

	for objectType, objectIDs := range objectTypeToIds {
		returnTableIDs := []int64{}
		switch objectIDs.(type) {
		// If slice of channel IDs is given
		case []string:
			for _, singleID := range objectIDs.([]string) {
				returnTableIDs = append(returnTableIDs, addObjectToDatabaseTableFromYouTubeID(db, conn, dbName, singleID, objectType))
			}
			// If single channel ID given
		case string:
			returnTableIDs = append(returnTableIDs, addObjectToDatabaseTableFromYouTubeID(db, conn, dbName, objectIDs.(string), objectType))
			// If other type given
		default:
			returnTableIDs = []int64{-1}
		}

		// End for
		returnTableIdsMap[objectType] = returnTableIDs
	}

	return returnTableIdsMap
}

func ParseAPIDurationResponse(durationRes string) int64 {
	regex := regexp.MustCompile("P(?:(\\d*)D)?T(?:(\\d*)H)?(\\d*)M(\\d*)S")

	matches := regex.FindStringSubmatch(durationRes)

	if len(matches) > 0 {
		days := ComplexParseInt(matches[1], 10, 0)
		hours := ComplexParseInt(matches[2], 10, 0)
		minutes := ComplexParseInt(matches[3], 10, 0)
		seconds := ComplexParseInt(matches[4], 10, 0)

		return ((days*24+hours)*60+minutes)*60 + seconds

	}

	return 0

}

func ComplexParseInt(stringInt string, base, bitSize int) int64 {

	asInt, err := strconv.ParseInt(stringInt, base, bitSize)

	if err != nil {
		return 0
	}

	return asInt

}

func InitializeDemo2(DBCO databaseDriver.DataBaseConnectionObject) {

	ytService := NewYouTubeConnection()

	service := ServiceConnectionObject{ytService}

	// channelDBTable := databaseDriver.NewDataBaseTableObject("youtubecontent", "channelData")
	// videoDBTable := databaseDriver.NewDataBaseTableObject("youtubecontent", "videoData")

	// CIO := makeChannelInsertObject(DBCO, service, channelDBTable)

	// CIO.

	// conanID := "UCi7GJNg51C3jgmYTUwqoUXA"
	// colNum := addObjectToDatabaseTableFromYouTubeID(DBCO, service, "youtubedata", "UCi7GJNg51C3jgmYTUwqoUXA", &youtube.Channel{})

	colNum := addObjectToDatabaseTableFromYouTubeID(DBCO, service, "youtubedata", "3JX8TivvGyc", &youtube.Video{})

	fmt.Printf("tableId is %v\n", colNum)
}

//TODO: UPDATE DATABASE TO SUPPORT UTF-8 ENCODING
