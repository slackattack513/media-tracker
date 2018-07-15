package youtubecontent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
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

func printJSON(obj interface{}) {
	retJSONObj, _ := json.MarshalIndent(obj, "", "    ")
	fmt.Printf("%s\n", retJSONObj)
}

func DemoMain() {
	fmt.Println("in")
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)
	service, err := youtube.New(client)

	handleError(err, "Error creating YouTube client")
	fmt.Println("About to do command")
	// channelsListByUsername(service, "snippet,contentDetails,contentOwnerDetails,statistics", "GoogleDevelopers")
	// ifscID := "UC2MGuhIaOP6YLpUx106kTQw"
	// mrWolfId := "UCs0XZm6REAhwVwmyBPRAKMQ"
	// conanID := "UCi7GJNg51C3jgmYTUwqoUXA"
	// slackID := "UCyHI7IpiV3ogpKzzm0xxqnA"
	// channelsListById(service, "brandingSettings,snippet,contentDetails,contentOwnerDetails,statistics", slackID)
	// slackFavoritesPlaylist := "FLyHI7IpiV3ogpKzzm0xxqnA"
	slackLikesPlaylist := "LLyHI7IpiV3ogpKzzm0xxqnA"
	// listVideosFromPlaylist(service, "snippet,contentDetails,id,status", slackLikesPlaylist)

	// likedVideosPlaylistItemListResponseObject := callListPlaylistObject(service, "contentDetails", slackLikesPlaylist, 50, "")
	// vids := getPlaylistVideosIdsSinglePlaylistItemListResponse(likedVideosPlaylistItemListResponseObject)
	// printJSON(vids)

	allLikedVidsPlaylist := callListPlaylistObjectAccessAllPages(service, "contentDetails", slackLikesPlaylist, 50)
	vids := getPlaylistVideosIdsMultiplePlaylistItemListResponse(allLikedVidsPlaylist)
	printJSON(vids)

	// Get my liked videos
	// COMPLETE: Get the videos on my liked playlist - only need 'contentDetails'
	// COMPLETE: parse the JSON response to get each video's videoID
	// Step 3: for each videoID look up that videoInfo
	// Step 4: check if video in database
	// 	- if not, start process to add video to database:
	// 		* check if video channel in db
	//			# if not, add channel to db
	//		* add video to db
	//		* link video to channel ( so that video can point to channelTableID)
	// Step 5: add entry to likedPlaylist DB and link to video in video DB

	// This process can be used for most playlists
	// To generalize this process further,
	//	* COMPLETE: write code to create a DB for a new playlist - Should not take much more space since video info is only saved once in that single DB

}
