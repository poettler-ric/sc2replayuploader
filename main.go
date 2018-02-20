package main

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"github.com/poettler-ric/sc2replayuploader/uploader"
	"github.com/vharitonsky/iniflags"
	"log"
	"path/filepath"
)

var (
	token     string
	rootDir   string
	hash      string
	uploadAll bool
)

func init() {
	flag.StringVar(&token, "token", "", "authentication token to use")
	flag.StringVar(&rootDir, "dir", "", "root folder for replays")
	flag.StringVar(&hash, "hash", "", "hash of the account for the replays")
	flag.BoolVar(&uploadAll, "all", false,
		"ulpload all replays (not just the newest)")
}

func main() {
	defaultConfigFile, err := homedir.Expand("~/.sc2replayuploader.conf")
	if err != nil {
		log.Fatalf("error while assembling default configfile: %v", err)
	}
	iniflags.SetConfigFile(defaultConfigFile)
	iniflags.Parse()
	if hash == "" || rootDir == "" || token == "" {
		flag.Usage()
		log.Fatalln("dir, hash and token must be set")
	}
	rootDir, err = homedir.Expand(rootDir)
	if err != nil {
		log.Fatalf("error while expanding homedir: %v", err)
	}

	var files []string
	if uploadAll {
		files, err = uploader.GetAllReplayFiles(rootDir)
		if err != nil {
			log.Fatalf("error when getting replay files: %v", err)
		}
	} else {
		lastReplay, err := uploader.GetLastReplay(token)
		if err != nil {
			log.Fatalf("error while getting last replay: %v", err)
		}
		files, err = uploader.GetNewerReplayFiles(rootDir, lastReplay)
		if err != nil {
			log.Fatalf("error when getting replay files: %v", err)
		}
	}
	for _, path := range files {
		log.Printf("uploading %v", filepath.Base(path))
		result, err := uploader.UploadReplay(hash, token, path)
		if err != nil {
			log.Fatalf("error while uploading replay: %v", err)
		}
		log.Printf("queued uploaded file with id %v", result.QueueID)
	}
}
