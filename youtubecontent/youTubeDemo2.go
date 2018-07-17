package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"

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
}

type GeneralYouTubeAPIObjectI interface {
	RequestYouTubeAPIObjectFromId(string)
}

type channelInsertObject struct {
	databaseDriver.DataInsertObject
	ServiceConnectionObject
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

func makeChannelInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, DBTO databaseDriver.DataBaseTableObject) *channelInsertObject {
	return &channelInsertObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn, nil}
}

func addObjectToDatabaseTable(db databaseDriver.DataBaseConnectionObject, conn ServiceConnectionObject, DBTO databaseDriver.DataBaseTableObject, objectID string, objectType interface{}) int64 {
	var obj APIToDataBaseControllerObjectI
	var columnName string

	switch objectType.(type) {
	case *youtube.Channel:
		obj = makeChannelInsertObject(db, conn, DBTO)
		columnName = "channelId"
	case *youtube.Video:
		obj = makeChannelInsertObject(db, conn, DBTO)
		columnName = "videoId"
	default:
		return -1
	}

	// obj = obj.(*channelInsertObject)

	// obj.CellExists()

	// if !StringDataInDatabase(conn, databaseName, tableName, "channelId", channelID) {

	if !obj.CellExists(columnName, objectID) {
		obj.RequestYouTubeAPIObjectFromId(objectID)
		obj.CreateDataMap()
		obj.PrepareRequestStatement()
		obj.RunDBExecution()
	}

	// TODO Create methods to get values from the database table. Mainly copy getDataTableColumnValue() from the original code
	return obj.GetCellValue()
}
