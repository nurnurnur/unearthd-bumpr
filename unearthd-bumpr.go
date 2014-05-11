package main

import (
  "fmt"
  "strings"
  "time"
  "encoding/json"
  "strconv"
  "bufio"
  "os"
  "log"
  "net/http"
  "io/ioutil"
  "./confirm"
  flag "launchpad.net/gnuflag"
)

const VERSION = "0.3"

var helpFlag = flag.Bool("help", false, "Show this screen")
var tracksFlag = flag.String("tracks", "", "Comma separated list of track_ids eg. 123,231,122")
var fileFlag = flag.String("file", "", "A file of line separated track_ids")

func output_welcome() {
  fmt.Println()
  fmt.Printf("Unearthd Track Bumpr - Version %s\n", VERSION)
  fmt.Println("Created by NUR (Never Underestimate Reason)")
  fmt.Println()
}

func output_help() {
  flag.Usage()
}

func output_tracklist(tracks []TrackInfo) (string) {
  var output string
  for i,track := range tracks {
    if track.Duration == "" {
      track.Duration = "00:00:00"
    } else {
      track.Duration = strings.TrimSuffix(track.Duration, "\n")
    }
    output += fmt.Sprintf("%d. %s - %s [%s]\n", i+1, track.ArtistTitle, track.Title, track.Duration)
  }
  return output
}

func sleep_for_track_length(length string) () {
  var sleep_dur time.Duration
  duration_arr := strings.Split(length, ":")

  fmt.Println("Waiting for "+length)

  hours,_ := strconv.ParseInt(duration_arr[0],0,32)
  mins,_ := strconv.ParseInt(duration_arr[1],0,32)
  secs,_ := strconv.ParseInt(duration_arr[2],0,32)

  sleep_dur += time.Duration(hours)*time.Hour
  sleep_dur += time.Duration(mins)*time.Minute
  sleep_dur += time.Duration(secs)*time.Second
  time.Sleep(sleep_dur)
}

func get_track_info(track_id int) (*TrackInfoCollection) {
  var track_url string
  var jukebox_url string

  track_url = build_track_url(track_id)
  jukebox_url = build_jukebox_url(track_id)

  output := http_get(track_url,jukebox_url)

  tic := new(TrackInfoCollection)

  if err := tic.FromJson(output); err != nil {
    fmt.Printf("ERROR: %v", err)
  }

  return tic
}

func hit_jukebox(track_id int, artist_url string) (string) {
  var jukebox_url string

  jukebox_url = build_jukebox_url(track_id)

  http_get(artist_url,"")
  output := http_get(jukebox_url, artist_url)

  return output
}

func hit_track_play(track_id string) {
  url := build_play_url(track_id)
  http_get(url,"")
}

func build_play_url(track_id string) (string) {
  return fmt.Sprintf("https://www.triplejunearthed.com/play/%s",track_id)
}

func build_jukebox_url(track_id int) (string) {
  return fmt.Sprintf("https://www.triplejunearthed.com/jukebox/play/track/%d",track_id)
}

func build_track_url(track_id int) (string) {
  return fmt.Sprintf("https://www.triplejunearthed.com/api/jukebox/rest/views/jukebox_track?args=%d",track_id)
}

func full_artist_url(path string) (string) {
  return "https://www.triplejunearthed.com"+path
}

func hit_mp3_url(url string, etag string) {
  fmt.Println("Hitting mp3 URL..")
  http_etag_get(url,etag,"")
}

func track_ids_from_tracks_flag() ([]int) {
  var track_ids []int

  for _,arg := range strings.Split(*tracksFlag,",") {
    var int_track_id int
    int_track_id64,_ := strconv.ParseInt(arg,0,0)
    int_track_id = int(int_track_id64)
    track_ids = append(track_ids, int_track_id)
  }

  return track_ids
}

func track_ids_from_file_flag() ([]int) {
  var track_ids []int
  file,err := os.Open(*fileFlag)
  if err != nil {
    log.Fatalln("Error opening file")
  }
  defer file.Close()
  scanner := bufio.NewScanner(file)
  scanner.Split(bufio.ScanLines)
  for scanner.Scan() {
      str := scanner.Text()
      var int_track_id int
      int_track_id64,_ := strconv.ParseInt(str,0,0)
      int_track_id = int(int_track_id64)
      track_ids = append(track_ids, int_track_id)
  }
  return track_ids
}

func track_ids_from_stdin() ([]int) {
  var track_ids []int
  stats,_ := os.Stdin.Stat()
  if stats.Size() > 0 {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        str := scanner.Text()
        var int_track_id int
        int_track_id64,_ := strconv.ParseInt(str,0,0)
        int_track_id = int(int_track_id64)
        track_ids = append(track_ids, int_track_id)
    }
  }

  return track_ids
}

func main() {
  var track_ids []int
  track_etags := map[string]string{}

  ok := true

  output_welcome()

  flag.Parse(true)
  if(*helpFlag) {
    output_help()
  } else {

    // If track flag is present, append listed track_ids
    if(*tracksFlag != "") { track_ids = append(track_ids, track_ids_from_tracks_flag()...) }

    // If file flag is present, append track_ids from file
    if(*fileFlag != "") { track_ids = append(track_ids, track_ids_from_file_flag()...) }

    if ok {
      var track_id int
      var tracks []TrackInfo

      if len(track_ids) == 0 {
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

      fmt.Println("Fetching all the track details...")

      for _,element := range track_ids {
        tracks = append(tracks, get_track_info(element).Tracks[0])
      }

      fmt.Println("Track list built.")

      fmt.Println(output_tracklist(tracks))

      fmt.Printf("Is this correct? [y/n] ")
      if confirm.AskForConfirmation() {
        for {
          for _,track := range tracks {
            if track.Duration == "" {
              track.Duration = "00:04:42"
            }

            fmt.Printf("Playing %s-%s\n",track.ArtistTitle, track.Title)

            if track_etags[track.URL] == "" {
              url_headers := http_head(track.URL)
              track_etags[track.URL] = url_headers["Etag"][0]
            }

            hit_mp3_url(track.URL,track_etags[track.URL])
            hit_track_play(track.ID)
            sleep_for_track_length(track.Duration)
          }
        }
      } else {
        fmt.Println()
        fmt.Println("Exiting unearthd-bumpr...")
        fmt.Println()
        os.Exit(2)
      }
    }

  }
}

func build_http(url string,request string) (*http.Request) {
  req, err := http.NewRequest(request, url, nil)
  if err != nil {
    log.Fatalln(err)
  }

  req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.131 Safari/537.36")
  req.Header.Add("X-Requested-With", "XMLHttpRequest")
  req.Header.Set("Accept", "application/json, text/javascript, */*")

  return req
}

func http_head(url string) (http.Header) {
  request := build_http(url, "HEAD")
  client := &http.Client{}

  resp,_ := client.Do(request)

  return resp.Header
}

func http_get(url string, referrer string) (string) {
  request := build_http(url, "GET")
  client := &http.Client{}

  if (referrer != "") {
    request.Header.Set("Referrer",referrer)
  }

  resp,_ := client.Do(request)
  body,_ := ioutil.ReadAll(resp.Body)

  return string(body)
}

func http_etag_get(url string, etag string, referrer string) (string,http.Header) {
  request := build_http(url, "GET")
  client := &http.Client{}

  request.Header.Set("If-none-match",etag)

  if (referrer != "") {
    request.Header.Set("Referrer",referrer)
  }

  resp,_ := client.Do(request)
  body,_ := ioutil.ReadAll(resp.Body)

  return string(body),resp.Header
}

type TrackInfo struct {
  ArtistBio string `json:"artist_bio,omitempty"`
  ArtistFollowCount string `json:"artist_follow_count,omitempty"`
  ArtistImageSmall string `json:"artist_image_small,omitempty"`
  ArtistID string `json:"artist_id,omitempty"`
  ArtistLocation string `json:"artist_location,omitempty"`
  ArtistProfileURL string `json:"artist_profile_url,omitempty"`
  ArtistTitle string `json:"artist_title,omitempty"`
  Approved string `json:"track_approved,omitempty"`
  ChartPos string `json:"track_chart_pos,omitempty"`
  DownloadCount string `json:"track_download_count,omitempty"`
  Duration string `json:"track_duration,omitempty"`
  Genres string `json:"track_genres,omitempty"`
  ID string `json:"track_id,omitempty"`
  LoveCount string `json:"track_love_count,omitempty"`
  Mature string `json:"track_mature,omitempty"`
  PlayCount string `json:"track_play_count,omitempty"`
  PlayedOn string `json:"track_played_on,omitempty"`
  Rating string `json:"track_rating,omitempty"`
  ReviewCount string `json:"track_review_count,omitempty"`
  URL128 string `json:"url_for_the_128k_media,omitempty"`
  URL string `json:"track_url,omitempty"`
  Title string `json:"track_title,omitempty"`
}

type TrackInfoCollection struct {
  Tracks []TrackInfo
}

func (tic *TrackInfoCollection) FromJson(jsonStr string) error {
    var data = &tic.Tracks
    decoder := json.NewDecoder(strings.NewReader(jsonStr))
    return decoder.Decode(&data)
}