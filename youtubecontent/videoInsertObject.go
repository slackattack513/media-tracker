package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"
	"strconv"
	"strings"

	youtube "google.golang.org/api/youtube/v3"
)

//videoInsertObject

type videoInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	youtubeVideo *youtube.Video
}

func (vio *videoInsertObject) RequestYouTubeAPIObjectFromId(id string, options map[string]interface{}) {
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

func (vio *videoInsertObject) GetID() string {
	return vio.GetDataMap()["videoId"].(string)
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
