package uploader

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// UploaderIdentifyer is used to identify ourselves on the api
	UploaderIdentifyer = "https://github.com/poettler-ric/sc2replayuploader"
	// APIURL is the root url for all api endpoints
	APIURL = "http://api.sc2replaystats.com"
	// LastReplayEndpoint points to the url to get the last endpoint
	LastReplayEndpoint = APIURL + "/account/last-replay"
	// ReplayEndpoint points to the replay url
	ReplayEndpoint = APIURL + "/replay"
	// ReplayBufferTime takes differences between local clock and times on
	// replays from sc2replaystats
	ReplayBufferTime = time.Minute * 5
	// ReplaySuffix is the filesuffix of replay files
	ReplaySuffix = ".SC2Replay"
)

// SC2Replay replresents a single replay retrieved from sc2replaystats
type SC2Replay struct {
	// ReplayTime is the parsed ReplayTimeString
	ReplayTime time.Time
	// ReplayTimeString is the time the replay happened as string
	ReplayTimeString string `json:"replay_date"`
}

// UploadResponse holds information on uploads
type UploadResponse struct {
	// StatusCode stores the http status code of the response
	StatusCode int
	// QueueIDString is the id of the job handling the upload
	QueueIDString string `json:"replay_queue_id"`
	// QueueID is the id of the job handling the upload parsed to int
	QueueID int
}

// GetLastReplay retrieves the last uploaded replay.
//
// token is the authorization to use.
func GetLastReplay(token string) (replay SC2Replay) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, LastReplayEndpoint, nil)
	if err != nil {
		log.Fatalf("error while creating get request (%v): %v",
			LastReplayEndpoint, err)
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error while doing the request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("error while reading body: %v", err)
		}
		err = json.Unmarshal(bodyBytes, &replay)
		if err != nil {
			log.Fatalf("error while reading json: %v", err)
		}
	}

	replay.ReplayTime, err = time.Parse(time.RFC3339,
		replay.ReplayTimeString)
	if err != nil {
		log.Fatalf("error while parsing time (%v): %v",
			replay.ReplayTimeString, err)
	}

	return
}

// GetNewerReplayFiles searches for all replays newer than the given one.
//
// rootFolder is the folder to search for replays. lastReplay is the last
// uploaded replay.
func GetNewerReplayFiles(rootFolder string,
	lastReplay SC2Replay) (files []string) {
	maxAge := lastReplay.ReplayTime.Add(-ReplayBufferTime)

	filepath.Walk(rootFolder,
		func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() &&
				strings.HasSuffix(path, ReplaySuffix) &&
				info.ModTime().After(maxAge) {
				files = append(files, path)
			}
			return nil
		})
	return
}

// UploadReplay uploads a replayfile.
//
// hash is the user to which to add the replays. token is the auth token to use.
// path to the file to upload.
func UploadReplay(hash, token, path string) (result UploadResponse) {
	log.Printf("uploading %v", filepath.Base(path))

	b := &bytes.Buffer{}
	mp := multipart.NewWriter(b)
	mp.WriteField("upload_method", UploaderIdentifyer)
	mp.WriteField("hashkey", hash)
	replayPart, err := mp.CreateFormFile("replay_file", filepath.Base(path))
	if err != nil {
		log.Fatalf("error while creating multipart for file (%v): %v",
			path,
			err)
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("error while opening file (%v): %v", path, err)
	}
	defer file.Close()

	_, err = io.Copy(replayPart, file)
	if err != nil {
		log.Fatalf("error wile copying filecontent (%v): %v", path, err)
	}
	mp.Close()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, ReplayEndpoint, b)
	if err != nil {
		log.Fatalf("error while creating get request (%v): %v",
			ReplayEndpoint, err)
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", mp.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error while doing the request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error while reading body: %v", err)
	}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Fatalf("error while reading json: %v", err)
	}
	result.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusOK {
		result.QueueID, err = strconv.Atoi(result.QueueIDString)
		if err != nil {
			log.Fatalf("error while parsing queuid (%v): %v",
				result.QueueIDString, err)
		}
		log.Printf("queued uploaded file with id %v", result.QueueID)
	} else {
		log.Fatalf("error while uploading: %v", string(bodyBytes))
	}

	return
}
