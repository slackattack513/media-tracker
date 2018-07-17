package youtubecontent

import (
	"customHelperPackages/prettyPrinting"
	"database/sql"
	"encoding/json"
	"fmt"
	"media-tracker/databaseDriver"
	"os"
	"strconv"
	"strings"

	"google.golang.org/api/youtube/v3"

	_ "github.com/go-sql-driver/mysql"
)

type connections struct {
	db      *sql.DB
	service *youtube.Service
}

func channelsListByUsername(service *youtube.Service, part string, forUsername string) {
	call := service.Channels.List(part)
	call = call.ForUsername(forUsername)
	response, err := call.Do()
	handleError(err, "")
	fmt.Println("%s", response.Items[0].Id)
	fmt.Println(fmt.Sprintf("This channel's ID is %s. Its title is '%s', "+
		"and it has %d views.",
		response.Items[0].Id,
		response.Items[0].Snippet.Title,
		response.Items[0].Statistics.ViewCount))
}

func callListPlaylistObjectAccessAllPages(service *youtube.Service, params string, playlistID string, maxResults int64) []*youtube.PlaylistItemListResponse {
	ret := []*youtube.PlaylistItemListResponse{}
	var nextPageToken string
	// videoIDSlices := []string{}
	for {
		response := callListPlaylistObject(service, params, playlistID, maxResults, nextPageToken)

		// printJSON(response)
		ret = append(ret, response)
		// videoIDSlices = append(videoIDSlices, getPlaylistVideosIdsSinglePlaylistItemListResponse(response)...)
		if response.NextPageToken != "" {
			nextPageToken = response.NextPageToken
			// fmt.Printf("%s\n", nextPageToken)
		} else {
			break
		}

	}

	return ret
}

func callListPlaylistObject(service *youtube.Service, params string, playlistID string, maxResults int64, pageToken string) *youtube.PlaylistItemListResponse {
	var maxResultsDefault int64 = 10
	if maxResults == 0 {
		maxResults = maxResultsDefault
	} else if maxResults > 50 {
		maxResults = 50
	}

	call := service.PlaylistItems.List(params)
	// call = call.ForUsername(forUsername)
	call = call.PlaylistId(playlistID)
	call = call.MaxResults(maxResults)

	if pageToken != "" {
		call.PageToken(pageToken)
	}
	response, err := call.Do()
	handleError(err, "")
	return response
}

func getPlaylistVideosIdsSinglePlaylistItemListResponse(response *youtube.PlaylistItemListResponse) []string {
	videoIDSlice := []string{}

	playlistItems := response.Items

	for _, item := range playlistItems {
		videoIDSlice = append(videoIDSlice, item.ContentDetails.VideoId)
	}

	return videoIDSlice
}

func getPlaylistVideosIdsMultiplePlaylistItemListResponse(responses []*youtube.PlaylistItemListResponse) []string {
	videoIDSlice := []string{}
	for _, response := range responses {
		playlistItems := response.Items

		for _, item := range playlistItems {
			videoIDSlice = append(videoIDSlice, item.ContentDetails.VideoId)
		}
	}
	return videoIDSlice
}

func getVideoListByIDsResult(service *youtube.Service, params string, ids []string) *youtube.VideoListResponse {
	call := service.Videos.List(params)
	call = call.Id(strings.Join(ids, ","))
	response, err := call.Do()
	if err != nil {
		fmt.Println("Error getting video info")
		os.Exit(1)
	}

	return response
}

func parseVideoListResultsToVideoSlice(results *youtube.VideoListResponse) []*youtube.Video {
	return results.Items
}

func addObjectsToDatabaseTable(conn connections, databaseName, tableName string, objects, objectType interface{}) []int64 {
	// db := conn.db
	returnTableIDs := []int64{}
	switch objects.(type) {
	// If slice of channel IDs is given
	case []string:
		for _, singleID := range objects.([]string) {
			returnTableIDs = append(returnTableIDs, addObjectToDatabaseTable(conn, databaseName, tableName, singleID, objectType))
		}
		// If single channel ID given
	case string:
		returnTableIDs = append(returnTableIDs, addObjectToDatabaseTable(conn, databaseName, tableName, objects.(string), objectType))
		// If other type given
	default:
		return []int64{-1}
	}
	return returnTableIDs
}
func addVideosToDatabaseTable(conn connections, databaseName, tableName string, videos interface{}) []int64 {

	returnTableIDs := []int64{}
	switch videos.(type) {
	// If slice of channel IDs is given
	case []string:
		for _, singleID := range videos.([]string) {
			returnTableIDs = append(returnTableIDs, addVideoToDatabaseTable(conn, databaseName, tableName, singleID))
		}
		// If single channel ID given
	case string:
		returnTableIDs = append(returnTableIDs, addVideoToDatabaseTable(conn, databaseName, tableName, videos.(string)))
		// If other type given
	default:
		return []int64{-1}
	}
	return returnTableIDs

	db := conn.db

	// ret := []bool{}
	videosDataMaps := []map[string]interface{}{}

	switch videos.(type) {
	case []*youtube.Video:
		for _, vid := range videos.([]*youtube.Video) {
			videosDataMaps = append(videosDataMaps, createVideoDataMap(vid))
		}
	case *youtube.Video:
		videosDataMaps = append(videosDataMaps, createVideoDataMap(videos.(*youtube.Video)))
	default:
		return nil

	}

	// Add to database
	for _, dataMap := range videosDataMaps {
		// Make sure the channel for the video is stored so that the video can be used to quickly point to its channel data

		channelTableId := getDataTableColumnValue(conn, databaseName, "channelData", "channelId", dataMap["channelId"].(string), "tableId")
		if channelTableId == -1 {
			channelTableId = addChannelToDatabaseTable(conn, databaseName, "channelData", dataMap["channelId"].(string))
		}

		// Update the video dataMap with its channel's ID in the channelTable
		dataMap["tableId"] = channelTableId

		// Add the video data to the table in the database
		databaseDriver.AddDataToTable(db, databaseName, tableName, dataMap)

	}

	return returnTableIDs
}

// Returns slice of channelTableID in same order as the data was given
func addChannelsToDatabaseTable(conn connections, databaseName, tableName string, channelIds interface{}) []int64 {
	// db := conn.db
	returnTableIDs := []int64{}
	switch channelIds.(type) {
	// If slice of channel IDs is given
	case []string:
		for _, singleID := range channelIds.([]string) {
			returnTableIDs = append(returnTableIDs, addChannelToDatabaseTable(conn, databaseName, tableName, singleID))
		}
		// If single channel ID given
	case string:
		returnTableIDs = append(returnTableIDs, addChannelToDatabaseTable(conn, databaseName, tableName, channelIds.(string)))
		// If other type given
	default:
		return []int64{-1}
	}
	return returnTableIDs
}
func addVideoToDatabaseTable(conn connections, databaseName, tableName string, videoID string) int64 {
	// db := conn.db
	if !StringDataInDatabase(conn, databaseName, tableName, "videoId", videoID) {
		obj := getVideoObject(conn, videoID)
		insertDataMapToTable(conn, databaseName, tableName, createObjectDataMapFactory(obj))
	}

	return getDataTableColumnValue(conn, databaseName, tableName, "videoId", videoID, "videoTableId")
}

func addChannelToDatabaseTable(conn connections, databaseName, tableName string, channelID string) int64 {
	// db := conn.db
	if !StringDataInDatabase(conn, databaseName, tableName, "channelId", channelID) {
		obj := getChannelObject(conn, channelID)
		insertDataMapToTable(conn, databaseName, tableName, createObjectDataMapFactory(obj))
	}

	return getDataTableColumnValue(conn, databaseName, tableName, "channelId", channelID, "tableId")
}

func getDataTableColumnValue(conn connections, databaseName, tableName, columnNameKnown, columnValueKnown, columnNameDesired string) int64 {

	db := conn.db
	var dataValuesMap = make(map[string]string)
	dataValuesMap[columnNameKnown] = columnValueKnown
	retrievedData := databaseDriver.GetDataFromTable(db, databaseName, tableName, []string{columnNameDesired}, dataValuesMap)

	if len(retrievedData) > 0 {
		ret, err := strconv.ParseInt(string(retrievedData[0][columnNameDesired].([]uint8)), 10, 0)
		if err != nil {

		}
		return ret
	} else {
		return -1
	}
}

func StringDataInDatabase(conn connections, databaseName, tableName, columnName, columnValue string) bool {

	db := conn.db
	var dataValuesMap = make(map[string]string)
	dataValuesMap[columnName] = columnValue
	retrievedData := databaseDriver.GetDataFromTable(db, databaseName, tableName, []string{}, dataValuesMap)
	ret := len(retrievedData) > 0
	fmt.Printf("Data is in database: %t\n", ret)
	return ret
}

func getObjectFromIDFactory(conn connections, id string, objectType interface{}) interface{} {
	switch objectType.(type) {
	case *youtube.Channel:
		// part = "contentDetails, statistics, snippet, topicDetails"
		// call = conn.service.Channels.List(part)
		// call.Id(id)
		return getChannelObject(conn, id)
	case *youtube.Video:
		// part = "contentDetails, statistics, snippet, topicDetails"
		// call = conn.service.Videos.List(part)
		return getVideoObject(conn, id)
	default:
		return nil

	}

	// call.Id(id)
	// response, err := call.Do()

	// if err != nil {
	// 	fmt.Println("Error getting channelResponse")
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// return response.Items[0]
}

func getVideoObject(conn connections, videoId string) *youtube.Video {
	params := "contentDetails, statistics, snippet, topicDetails"
	call := conn.service.Videos.List(params)
	call.Id(videoId)
	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting videoResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	return response.Items[0]
}

func getChannelObject(conn connections, channelId string) *youtube.Channel {

	params := "contentDetails, statistics, snippet, topicDetails"
	call := conn.service.Channels.List(params)
	call.Id(channelId)
	response, err := call.Do()

	if err != nil {
		fmt.Println("Error getting channelResponse")
		fmt.Println(err)
		os.Exit(1)
	}

	return response.Items[0]
}

// Takes a *youtube.Video object and constructs a map of key:val pairs specifying table_column_name: new_column_data
func createVideoDataMap(video *youtube.Video) map[string]interface{} {
	return nil
}

func createObjectDataMapFactory(object interface{}) map[string]interface{} {
	switch object.(type) {
	case *youtube.Channel:
		return createChannelDataMap(object.(*youtube.Channel))
	case *youtube.Video:
		return createVideoDataMap(object.(*youtube.Video))
	default:
		return make(map[string]interface{})
	}
}

func createChannelDataMap(channel *youtube.Channel) map[string]interface{} {
	prettyPrinting.PrintJSON(channel)
	var ret = make(map[string]interface{})

	title := channel.Snippet.Title
	URL := channel.Snippet.CustomUrl
	id := channel.Id
	desc := channel.Snippet.Description
	subscr := channel.Statistics.SubscriberCount
	uploadPlaylistId := channel.ContentDetails.RelatedPlaylists.Uploads
	country := channel.Snippet.Country

	ret["channelName"] = title
	ret["URL"] = URL
	ret["channelId"] = id
	ret["description"] = desc
	ret["subscriberCount"] = subscr
	ret["uploadPlaylistID"] = uploadPlaylistId
	ret["country"] = country

	return ret
}

func insertDataMapToTable(conn connections, databaseName, tableName string, channelDataMap map[string]interface{}) {

	db := conn.db
	if len(channelDataMap) > 0 {
		databaseDriver.AddDataToTable(db, databaseName, tableName, channelDataMap)
	}

}

func channelsListById(service *youtube.Service, part string, id string) {
	call := service.Channels.List(part)
	call = call.Id(id)
	response, err := call.Do()
	handleError(err, "")
	// fmt.Println("%s", response.Items[0].Id)
	for _, item := range response.Items {
		fmt.Println("Im here")
		// retJSON, _ := item.MarshalJSON()
		// fmt.Printf("%s\n", retJSON)
		// var buf bytes.Buffer
		retJSONObj, _ := json.MarshalIndent(item, "", "    ")
		fmt.Printf("%s\n", retJSONObj)
		// fmt.Fprintf(&buf, string(retJSONObj))
		// fmt.Println(item.Id, item.Snippet)
	}
	// fmt.Println(fmt.Sprintf("This channel's ID is %s. Its title is '%s', "+
	// 	"and it has %d views.",
	// 	response.Items[0].Id,
	// 	response.Items[0].Snippet.Title,
	// 	response.Items[0].Statistics.ViewCount))
}

func Initialize(db *sql.DB) {

	service := NewYouTubeConnection()

	conn := connections{db: db, service: service}

	// channelsListByUsername(service, "snippet,contentDetails,contentOwnerDetails,statistics", "GoogleDevelopers")
	// ifscID := "UC2MGuhIaOP6YLpUx106kTQw"
	// mrWolfId := "UCs0XZm6REAhwVwmyBPRAKMQ"
	// conanID := "UCi7GJNg51C3jgmYTUwqoUXA"
	// slackID := "UCyHI7IpiV3ogpKzzm0xxqnA"
	// channelsListById(service, "brandingSettings,snippet,contentDetails,contentOwnerDetails,statistics", slackID)
	// slackFavoritesPlaylist := "FLyHI7IpiV3ogpKzzm0xxqnA"
	// slackLikesPlaylist := "LLyHI7IpiV3ogpKzzm0xxqnA"
	// listVideosFromPlaylist(service, "snippet,contentDetails,id,status", slackLikesPlaylist)

	// likedVideosPlaylistItemListResponseObject := callListPlaylistObject(service, "contentDetails", slackLikesPlaylist, 50, "")
	// vids := getPlaylistVideosIdsSinglePlaylistItemListResponse(likedVideosPlaylistItemListResponseObject)
	// printJSON(vids)

	// allLikedVidsPlaylist := callListPlaylistObjectAccessAllPages(service, "contentDetails", slackLikesPlaylist, 50)
	// vids := getPlaylistVideosIdsMultiplePlaylistItemListResponse(allLikedVidsPlaylist)

	suicideSheepId := "UC5nc_ZtjKW1htCVZVRxlQAQ"
	cols := addChannelsToDatabaseTable(conn, "youtubedata", "channelData", suicideSheepId)

	prettyPrinting.PrintJSON(cols)

	// Get my liked videos
	// COMPLETE: Get the videos on my liked playlist - only need 'contentDetails'
	// COMPLETE: parse the JSON response to get each video's videoID
	// COMPLETE: for each videoID look up that videoInfo
	// Step 4: check if video in database   TODO
	// 	- if not, start process to add video to database:
	// 		* check if video channel in db TODO
	//			# if not, add channel to db TODO
	//		* add video to db PARTIALLY DONE
	//		* link video to channel ( so that video can point to channelTableID) COMPLETED
	// Step 5: add entry to likedPlaylist DB and link to video in video DB TODO

	// This process can be used for most playlists
	// To generalize this process further,
	//	* COMPLETE: write code to create a DB for a new playlist - Should not take much more space since video info is only saved once in that single DB

}
