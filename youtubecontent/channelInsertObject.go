package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"

	youtube "google.golang.org/api/youtube/v3"
)

// channelInsertObject

type channelInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	// youtube.Channel
	youtubeChannel *youtube.Channel
}

func (cio *channelInsertObject) RequestYouTubeAPIObjectFromId(id string, options map[string]interface{}) {
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

func (cio *channelInsertObject) GetID() string {
	return cio.GetDataMap()["channelId"].(string)
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
