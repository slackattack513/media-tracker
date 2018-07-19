package youtubecontent

import (
	"fmt"
	"media-tracker/databaseDriver"
	"os"

	youtube "google.golang.org/api/youtube/v3"
)

// playlistInsertObject

// TODO

type playlistInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	maxResultsDefault int
	// youtube.Channel
	playlistID                       string
	youtubePlaylistItemListResponses []*youtube.PlaylistItemListResponse
}

func (pio *playlistInsertObject) getDefaultMaxResults() int64 {
	return 10
}

func makePlaylistInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, dbName string, playlistID string) *playlistInsertObject {
	// defaultMaxResults := 10
	DBTO := databaseDriver.NewDataBaseTableObject(dbName, "playlistData")
	return &playlistInsertObject{GeneralAPIToDataBaseControllerObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn}, 10, playlistID, nil}
}

func (pio *playlistInsertObject) RequestYouTubeAPIObjectFromId(playlistID string, options map[string]interface{}) {

	params := "contentDetails"
	call := pio.getServiceConnection().PlaylistItems.List(params)
	call.PlaylistId(playlistID)

	// Check options

	// pageToken
	var pageToken string
	if pToken, ok := options["pageToken"]; ok {
		if pTokenString, okk := pToken.(string); okk {
			if pTokenString == "*" {
				options["recurse"] = true
			} else {
				options["recurse"] = false
				pageToken = pTokenString
			}
		}
	}

	if pageToken != "" {
		call.PageToken(pageToken)
	}

	// maxResults
	var maxResultsInt64 int64 = -1
	if maxResults, ok := options["maxResults"]; ok {
		if maxResultsInt, ok := maxResults.(int64); ok {
			if maxResultsInt == 0 {
				maxResultsInt = pio.getDefaultMaxResults()
			} else if maxResultsInt > 50 {
				maxResultsInt = 50
			}
			maxResultsInt64 = maxResultsInt

		} else {
			maxResultsInt64 = pio.getDefaultMaxResults()
		}

	}
	if maxResultsInt64 >= 0 {
		call.MaxResults(maxResultsInt64)
	}

	// Make call, check errors
	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting playlistListResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	// Add response to struct
	pio.youtubePlaylistItemListResponses = append(pio.youtubePlaylistItemListResponses, response)
	// Check recursion for all pages
	if recurse, ok := options["recurse"]; ok {
		recurseBool, okk := recurse.(bool)
		if okk && recurseBool {
			nextPageToken := response.NextPageToken
			if nextPageToken != "" {
				pio.RequestYouTubeAPIObjectFromId(playlistID, map[string]interface{}{"recurse": true, "pageToken": nextPageToken, "maxResults": maxResultsInt64})
			}
		}
	}

	// pio.playlistName =

}

//
func (pio *playlistInsertObject) PrepareRequestStatement() {

	newData := pio.GetDataMap()
	// var newDataMap[string]interface{}
	// var playlistItemIDs

	for _, playlistItemInsertObj := range newData {
		if piio, ok := playlistItemInsertObj.(playlistItemInsertObject); ok {
			piio.PrepareAndExecute()
		} else {
			fmt.Println("Somethign seriously wrong. Cast to playlistItemInsertObject should work by construction")
		}
	}
	pio.SetRequestStatement("")
}

func (pio *playlistInsertObject) GetID() string {
	return pio.playlistID
}

//
func (pio *playlistInsertObject) CreateDataMap() {
	var ret map[string]interface{}

	for _, plItem := range pio.youtubePlaylistItemListResponses {
		for _, pItem := range plItem.Items {
			newPlaylistItemInsertObj := makePlaylistItemInsertObject(pio.DataBaseConnectionObject, pio.ServiceConnectionObject, pio.DataBaseTableObject)
			newPlaylistItemInsertObj.youtubePlaylistItem = pItem
			newPlaylistItemInsertObj.CreateDataMap()
			ret[newPlaylistItemInsertObj.GetID()] = newPlaylistItemInsertObj
		}

	}

	pio.SetDataMap(ret)
}

type playlistItemInsertObject struct {
	GeneralAPIToDataBaseControllerObject
	youtubePlaylistItem *youtube.PlaylistItem
}

func makePlaylistItemInsertObject(dataConn databaseDriver.DataBaseConnectionObject, serviceConn ServiceConnectionObject, DBTO databaseDriver.DataBaseTableObject) *playlistItemInsertObject {
	// DBTO := dbTabObj
	return &playlistItemInsertObject{GeneralAPIToDataBaseControllerObject{databaseDriver.NewDataInsertObject(databaseDriver.NewDataBaseTableRequestObject(dataConn, DBTO, ""), make(map[string]interface{})), serviceConn}, nil}
}

func (pio *playlistItemInsertObject) RequestYouTubeAPIObjectFromId(itemID string, options map[string]interface{}) {

	params := "snippet, contentDetails"
	call := pio.getServiceConnection().PlaylistItems.List(params)
	call.Id(itemID)

	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting playlistListResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	// By construction only a single return item
	pio.youtubePlaylistItem = response.Items[0]
}

func (piio *playlistItemInsertObject) GetID() string {

	return piio.GetDataMap()["playlistItemID"].(string)

}

func (vio *playlistItemInsertObject) CreateDataMap() {
	var ret = make(map[string]interface{})

	id := vio.youtubePlaylistItem.Id

	title := vio.youtubePlaylistItem.Snippet.Title
	// channelId := vio.youtubePlaylistItem.Snippet.ChannelId
	videoId := vio.youtubePlaylistItem.ContentDetails.VideoId
	channelTitle := vio.youtubePlaylistItem.Snippet.ChannelTitle

	position := vio.youtubePlaylistItem.Snippet.Position
	// desc := vio.youtubePlaylistItem.Snippet.Description

	// duration := ParseAPIDurationResponse(vio.youtubeVideo.ContentDetails.Duration)

	// viewCount := vio.youtubeVideo.Statistics.ViewCount
	// likeCount := vio.youtubeVideo.Statistics.LikeCount
	// dislikeCount := vio.youtubeVideo.Statistics.DislikeCount

	ret["playlistItemID"] = id

	ret["playlistItemTitle"] = title
	ret["ownerChannelName"] = channelTitle
	ret["videoID"] = videoId

	ret["playlistPosition"] = position

	vio.SetDataMap(ret)
}

func (vio *playlistItemInsertObject) PrepareRequestStatement() {
	newData := vio.GetDataMap()

	if videoId, ok := newData["videoID"]; ok {
		videoIO := makeVideoInsertObject(vio.DataBaseConnectionObject, vio.ServiceConnectionObject, vio.GetDatabaseName())
		if !videoIO.CellExists("videoID", videoId) {
			ScrapeObjectThenAddToDatabase(videoIO, videoId.(string))
		}
		// channelTableId := channelIO.GetColumnCellValues("TABLEID", map[string][]string{"channelId": []string{channelId.(string)}})
		// res, err := strconv.ParseInt(string(channelTableId[0].([]uint8)), 10, 0)

		// if err != nil {
		// 	fmt.Println("Error getting channel table ID")
		// 	fmt.Println(err)
		// 	os.Exit(1)
		// }

		// newData["channelTableId"] = res

	}
	// vio.SetDataMap(newData)
	vio.GeneralAPIToDataBaseControllerObject.DataInsertObject.PrepareRequestStatement()
}
