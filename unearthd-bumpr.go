package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/nurnurnur/unearthd-bumpr/confirm"
	"github.com/nurnurnur/unearthd-bumpr/term"
	"io/ioutil"
	flag "launchpad.net/gnuflag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var VERSION string
var MINVERSION string

var NAME string = "Unearthd Track Bumpr"

var helpFlag = flag.Bool("help", false, "Show this screen")
var tracksFlag = flag.String("tracks", "", "Comma separated list of track_ids eg. 123,231,122")
var fileFlag = flag.String("file", "", "A file of line separated track_ids")
var playlistFlag = flag.String("playlist", "", "The id of a JJJ Unearthed playlist")
var playedTracksCount = map[string]int{}
var tracks []TrackInfo

var HTTP_ERR_RETRY = time.Duration(10)
var HTTP_GET_ERROR = "HTTP GET for %s failed.\nWaiting 10s and trying again.."
var HTTP_ETAG_GET_ERROR = "ETAG cached HTTP GET for %s failed.\nWaiting 10s and trying again.."

func output_welcome() {
	fmt.Printf("\033]0;unearthd-bumpr v%s\007", VERSION)
	fmt.Println()
	fmt.Printf("%s - v%s\n", NAME, VERSION)
	fmt.Println("Created by NUR (Never Underestimate Reason)")
	fmt.Println("'It ain't no fun, if the homies can't have none'")
	fmt.Println()
}

func exit_and_output_stats() {
	fmt.Println()
	fmt.Printf(term.Red + "Exiting %s..." + term.Reset,NAME)
	fmt.Println()
	fmt.Println(term.Green + "Track stats:" + term.Reset)
	for tracknum, track := range tracks {
		play_count := playedTracksCount[track.ID]
		fmt.Printf("%d. %s - %s [%d]\n", tracknum+1, track.ArtistTitle, track.Title, play_count)
	}
	os.Exit(0)
}

func output_help() {
	fmt.Printf("Built on: %s\n", MINVERSION)
	flag.Usage()
}

func output_tracklist(tracks []TrackInfo) string {
	var output string
	for i, track := range tracks {
		if track.Duration == "" {
			track.Duration = "00:00:00"
		} else {
			track.Duration = strings.TrimSuffix(track.Duration, "\n")
		}
		output += fmt.Sprintf("%d. %s - %s [%s]\n", i+1, track.ArtistTitle, track.Title, track.Duration)
	}
	return output
}

func sleep_for_track_length(length string) {
	var sleep_dur time.Duration
	duration_arr := strings.Split(length, ":")

	fmt.Println("Waiting for " + length)

	hours, _ := strconv.ParseInt(duration_arr[0], 0, 32)
	mins, _ := strconv.ParseInt(duration_arr[1], 0, 32)
	secs, _ := strconv.ParseInt(duration_arr[2], 0, 32)

	sleep_dur += time.Duration(hours) * time.Hour
	sleep_dur += time.Duration(mins) * time.Minute
	sleep_dur += time.Duration(secs) * time.Second
	time.Sleep(sleep_dur)
}

func get_track_info(track_id int) *TrackInfoCollection {
	var track_url string
	var jukebox_url string

	track_url = build_track_url(track_id)
	jukebox_url = build_jukebox_url(track_id)

	output := http_get(track_url, jukebox_url)

	tic := new(TrackInfoCollection)

	fmt.Println(output)
	if err := tic.FromXML(output); err != nil {
		fmt.Printf(term.Red+"ERROR: %v"+term.Reset, err)
	}

	return tic
}

func get_playlist_info(playlist_id string) *TrackInfoCollection {
	var playlist_url string

	playlist_url = build_playlist_info_url(playlist_id)
	output := http_get(playlist_url, "")

	tic := new(TrackInfoCollection)

	if err := tic.FromXML(output); err != nil {
		fmt.Printf(term.Red+"ERROR: %v"+term.Reset, err)
	}

	return tic
}

func hit_jukebox(track_id int, artist_url string) string {
	var jukebox_url string

	jukebox_url = build_jukebox_url(track_id)

	http_get(artist_url, "")
	output := http_get(jukebox_url, artist_url)

	return output
}

func hit_track_play(track_id string) {
	fmt.Println("Hitting play URL...")
	url := build_play_url(track_id)
	http_get(url, "")
}

func build_playlist_info_url(playlist_id string) string {
	return fmt.Sprintf("https://www.triplejunearthed.com/api/jukebox/rest/views/playlist_tracks?args=%s", playlist_id)
}

func build_play_url(track_id string) string {
	return fmt.Sprintf("https://www.triplejunearthed.com/play/%s", track_id)
}

func build_jukebox_url(track_id int) string {
	return fmt.Sprintf("https://www.triplejunearthed.com/jukebox/play/track/%d", track_id)
}

func build_track_url(track_id int) string {
	return fmt.Sprintf("https://www.triplejunearthed.com/api/jukebox/rest/views/jukebox_track?args=%d", track_id)
}

func build_track_download_track_url(track_id int) string {
	return fmt.Sprintf("https://www.triplejunearthed.com/download/track/%d", track_id)
}

func full_artist_url(path string) string {
	return "https://www.triplejunearthed.com" + path
}

func hit_mp3_url(url string, etag string) {
	fmt.Println("Hitting mp3 URL...")
	http_etag_get(url, etag, "")
}

func hit_user_url() {
	http_get("https://www.triplejunearthed.com/api/jukebox/rest/views/user")
}


func track_ids_from_tracks_flag() []int {
	var track_ids []int

	for _, arg := range strings.Split(*tracksFlag, ",") {
		var int_track_id int
		int_track_id64, _ := strconv.ParseInt(arg, 0, 0)
		int_track_id = int(int_track_id64)
		track_ids = append(track_ids, int_track_id)
	}

	return track_ids
}

func track_ids_from_file_flag() []int {
	var track_ids []int
	file, err := os.Open(*fileFlag)
	if err != nil {
		log.Fatalln(term.Red + "Error opening file" + term.Reset)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		str := scanner.Text()
		var int_track_id int
		int_track_id64, _ := strconv.ParseInt(str, 0, 0)
		int_track_id = int(int_track_id64)
		track_ids = append(track_ids, int_track_id)
	}
	return track_ids
}

func track_ids_from_stdin() []int {
	var track_ids []int
	stats, _ := os.Stdin.Stat()
	if stats.Size() > 0 {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			str := scanner.Text()
			var int_track_id int
			int_track_id64, _ := strconv.ParseInt(str, 0, 0)
			int_track_id = int(int_track_id64)
			track_ids = append(track_ids, int_track_id)
		}
	}

	return track_ids
}

func main() {
	var track_ids []int
	etags := map[string]string{}

	// capture ctrl+c and stop
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGQUIT)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		exit_and_output_stats()
	}()

	ok := true

	output_welcome()

	flag.Parse(true)
	if *helpFlag {
		output_help()
	} else {

		// If track flag is present, append listed track_ids
		if *tracksFlag != "" {
			track_ids = append(track_ids, track_ids_from_tracks_flag()...)
		}

		// If file flag is present, append track_ids from file
		if *fileFlag != "" {
			track_ids = append(track_ids, track_ids_from_file_flag()...)
		}

		if ok {
			var track_id int

			fmt.Println("Fetching all the track details...")

			if *playlistFlag != "" {
				tracks = get_playlist_info(*playlistFlag).Tracks
			}

			if (len(track_ids) == 0) && (len(tracks) == 0) {
				for {
					fmt.Printf("Enter a track_id: ")
					fmt.Scanf("%d", &track_id)

					if track_id == 0 {
						track_id = 816296
					}
					track_ids = append(track_ids, track_id)
					fmt.Printf("Add more tracks? [y/n] ")
					if !confirm.AskForConfirmation() {
						break
					}
				}
			}

			if len(track_ids) > 0 {
				track_info_collection := new(TrackInfoCollection)
				track_info_collection.FromTrackIds(track_ids)
				tracks = append(tracks, track_info_collection.Tracks...)
			}

			fmt.Println("Track list built.")

			fmt.Println(output_tracklist(tracks))

			fmt.Printf("Is this correct? [y/n] ")
			if confirm.AskForConfirmation() {
				for {
					for _, track := range tracks {
						if track.Duration == "" {
							track.Duration = "00:04:42"
						}

						fmt.Printf("Playing %s-%s [%d]\n", track.ArtistTitle, track.Title, playedTracksCount[track.ID])

						if etags[track.URL128] == "" {
							url_headers := http_head(track.URL128)
							etags[track.URL128] = url_headers["Etag"][0]
						}

						// hit_user_url()
						hit_mp3_url(track.URL128, etags[track.URL128])
						hit_track_play(track.ID)
						playedTracksCount[track.ID]++
						sleep_for_track_length(track.Duration)
					}
				}
			} else {
				fmt.Println()
				fmt.Printf("Exiting %s...",NAME)
				fmt.Println()
				os.Exit(2)
			}
		}

	}
}

func build_http(url string, request string) *http.Request {
	req, err := http.NewRequest(request, url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "*/*")

	return req
}

func http_head(url string) http.Header {
	request := build_http(url, "HEAD")
	client := &http.Client{}

	resp, _ := client.Do(request)

	return resp.Header
}

func http_get(url string, referrer string) string {
	request := build_http(url, "GET")
	client := &http.Client{}

	if referrer != "" {
		request.Header.Set("Referrer", referrer)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Println(term.Red + fmt.Sprintf(HTTP_GET_ERROR, url) + term.Reset)
		time.Sleep(HTTP_ERR_RETRY)
		return http_get(url, referrer)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body)
	}
}

func http_etag_get(url string, etag string, referrer string) (string, http.Header) {
	request := build_http(url, "GET")
	client := &http.Client{}

	request.Header.Set("If-none-match", etag)

	if referrer != "" {
		request.Header.Set("Referrer", referrer)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Println(term.Red + fmt.Sprintf(HTTP_ETAG_GET_ERROR, url) + term.Reset)
		time.Sleep(HTTP_ERR_RETRY)
		return http_etag_get(url, etag, referrer)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body), resp.Header
	}
}

type TrackInfo struct {
	ArtistBio         string `xml:"artist_bio,omitempty" json:"artist_bio,omitempty"`
	ArtistFollowCount string `xml:"artist_follow_count,omitempty" json:"artist_follow_count,omitempty"`
	ArtistImageSmall  string `xml:"artist_image_small,omitempty" json:"artist_image_small,omitempty"`
	ArtistID          string `xml:"artist_id,omitempty" json:"artist_id,omitempty"`
	ArtistLocation    string `xml:"artist_location,omitempty" json:"artist_location,omitempty"`
	ArtistProfileURL  string `xml:"artist_profile_url,omitempty" json:"artist_profile_url,omitempty"`
	ArtistTitle       string `xml:"artist_title,omitempty" json:"artist_title,omitempty"`
	Approved          string `xml:"track_approved,omitempty" json:"track_approved,omitempty"`
	ChartPos          string `xml:"track_chart_pos,omitempty" json:"track_chart_pos,omitempty"`
	DownloadCount     string `xml:"track_download_count,omitempty" json:"track_download_count,omitempty"`
	Duration          string `xml:"track_duration,omitempty" json:"track_duration,omitempty"`
	Genres            string `xml:"track_genres,omitempty" json:"track_genres,omitempty"`
	ID                string `xml:"track_id,omitempty" json:"track_id,omitempty"`
	LoveCount         string `xml:"track_love_count,omitempty" json:"track_love_count,omitempty"`
	Mature            string `xml:"track_mature,omitempty" json:"track_mature,omitempty"`
	PlayCount         string `xml:"track_play_count,omitempty" json:"track_play_count,omitempty"`
	PlayedOn          string `xml:"track_played_on,omitempty" json:"track_played_on,omitempty"`
	PlaylistDesc      string `xml:"playlist_description,omitempty" json:"playlist_description,omitempty"`
	PlaylistTitle     string `xml:"playlist_title,omitempty" json:"playlist_title,omitempty"`
	Rating            string `xml:"track_rating,omitempty" json:"track_rating,omitempty"`
	ReviewCount       string `xml:"track_review_count,omitempty" json:"track_review_count,omitempty"`
	URL128            string `xml:"url_for_the_128k_media,omitempty" json:"url for the 128k media,omitempty"`
	URL               string `xml:"track_url,omitempty" json:"track_url,omitempty"`
	Title             string `xml:"track_title,omitempty" json:"track_title,omitempty"`
}

type TrackInfoCollection struct {
	XMLName xml.Name `xml:"result"`
	Tracks []TrackInfo `xml:"item"`
}

func (tic *TrackInfoCollection) FromJson(jsonStr string) error {
	var data = &tic.Tracks
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	return decoder.Decode(&data)
}

func (tic *TrackInfoCollection) FromXML(xmlStr string) error {
	return xml.Unmarshal([]byte(xmlStr),&tic)
}

func (tic *TrackInfoCollection) FromTrackIds(track_ids []int) {
	for _, element := range track_ids {
		tracks := get_track_info(element)
		tic.Tracks = append(tic.Tracks, tracks.Tracks...)
	}
}
