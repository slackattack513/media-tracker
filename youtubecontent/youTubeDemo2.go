package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	RequestYouTubeAPIObjectFromId(string)
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
	genObj.RequestYouTubeAPIObjectFromId(id)
	genObj.CreateDataMap()
	genObj.PrepareRequestStatement()
	_, err := genObj.RunDBExecution()
	if err != nil {
		fmt.Println("Error in adding object to database")
		fmt.Println(err)
	}
}

type channelInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	// youtube.Channel
	youtubeChannel *youtube.Channel
}

func (cio *channelInsertObject) RequestYouTubeAPIObjectFromId(id string) {
	params := "contentDetails, statistics, snippet, topicDetails"
	call := cio.getServiceConnection().Channels.List(params)
	call.Id(id)
	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting channelResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	cio.youtubeChannel = response.Items[0]
}

func (cio *channelInsertObject) CreateDataMap() {
	var ret = make(map[string]interface{})

	title := cio.youtubeChannel.Snippet.Title
	URL := cio.youtubeChannel.Snippet.CustomUrl
	id := cio.youtubeChannel.Id
	desc := cio.youtubeChannel.Snippet.Description
	subscr := cio.youtubeChannel.Statistics.SubscriberCount
	uploadPlaylistId := cio.youtubeChannel.ContentDetails.RelatedPlaylists.Uploads
	country := cio.youtubeChannel.Snippet.Country

	ret["channelName"] = title
	ret["URL"] = URL
	ret["channelId"] = id
	ret["description"] = desc
	ret["subscriberCount"] = subscr
	ret["uploadPlaylistID"] = uploadPlaylistId
	ret["country"] = country

	cio.SetDataMap(ret)
}

func makeChannelInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, dbName string) *channelInsertObject {
	DBTO := databaseDriver.NewDataBaseTableObject(dbName, "channelData")
	return &channelInsertObject{GeneralAPIToDataBaseControllerObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn}, nil}
}

type videoInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	// youtube.Channel
	youtubeVideo *youtube.Video
}

func (vio *videoInsertObject) RequestYouTubeAPIObjectFromId(id string) {
	params := "contentDetails, statistics, snippet"
	call := vio.getServiceConnection().Videos.List(params)
	call.Id(id)
	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting videoResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	vio.youtubeVideo = response.Items[0]
}

func (vio *videoInsertObject) PrepareRequestStatement() {
	newData := vio.GetDataMap()

	if channelId, ok := newData["channelId"]; ok {
		channelIO := makeChannelInsertObject(vio.DataBaseConnectionObject, vio.ServiceConnectionObject, vio.GetDatabaseName())
		if !channelIO.CellExists("channelId", channelId) {
			ScrapeObjectThenAddToDatabase(channelIO, channelId.(string))
		}
		channelTableId := channelIO.GetColumnCellValues("TABLEID", map[string][]string{"channelId": []string{channelId.(string)}})
		res, err := strconv.ParseInt(string(channelTableId[0].([]uint8)), 10, 0)

		if err != nil {
			fmt.Println("Error getting channel table ID")
			fmt.Println(err)
			os.Exit(1)
		}

		newData["channelTableId"] = res

	}
	vio.SetDataMap(newData)
	vio.GeneralAPIToDataBaseControllerObject.DataInsertObject.PrepareRequestStatement()
}
func (vio *videoInsertObject) CreateDataMap() {
	var ret = make(map[string]interface{})

	id := vio.youtubeVideo.Id

	title := vio.youtubeVideo.Snippet.Title
	channelId := vio.youtubeVideo.Snippet.ChannelId
	desc := vio.youtubeVideo.Snippet.Description
	tags := strings.Join(vio.youtubeVideo.Snippet.Tags, ",")

	duration := ParseAPIDurationResponse(vio.youtubeVideo.ContentDetails.Duration)

	viewCount := vio.youtubeVideo.Statistics.ViewCount
	likeCount := vio.youtubeVideo.Statistics.LikeCount
	dislikeCount := vio.youtubeVideo.Statistics.DislikeCount

	ret["videoId"] = id

	ret["videoName"] = title
	ret["channelId"] = channelId
	ret["description"] = desc
	ret["tags"] = tags

	ret["duration"] = duration

	ret["viewCount"] = viewCount
	ret["likeCount"] = likeCount
	ret["dislikeCount"] = dislikeCount

	// vio.SetTableName("channelData")

	vio.SetDataMap(ret)
}

func makeVideoInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, dbName string) *videoInsertObject {
	DBTO := databaseDriver.NewDataBaseTableObject(dbName, "videoData")
	return &videoInsertObject{GeneralAPIToDataBaseControllerObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn}, nil}
}

// TODO

// func makePlaylistInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, dbName string, playlistName string) *channelInsertObject {
// 	DBTO := databaseDriver.NewDataBaseTableObject(dbName, "channelData")
// 	return &channelInsertObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn, nil}
//

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
		obj.RequestYouTubeAPIObjectFromId(objectID)
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
