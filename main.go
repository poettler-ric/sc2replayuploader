package main

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"github.com/poettler-ric/sc2replayuploader/uploader"
	"github.com/vharitonsky/iniflags"
	"log"
)

var (
	token   string
	rootDir string
	hash    string
)

func init() {
	flag.StringVar(&token, "token", "", "authentication token to use")
	flag.StringVar(&rootDir, "dir", "", "root folder for replays")
	flag.StringVar(&hash, "hash", "", "hash of the account for the replays")
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

	lastReplay := uploader.GetLastReplay(token)
	files := uploader.GetNewerReplayFiles(rootDir, lastReplay)
	for _, path := range files {
		uploader.UploadReplay(hash, token, path)
	}
}
