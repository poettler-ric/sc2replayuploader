package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/poettler-ric/sc2replayuploader/uploader"
	"github.com/vharitonsky/iniflags"
	"log"
	"os"
	"path/filepath"
	"sort"
)

var (
	token     string
	rootDir   string
	hash      string
	uploadAll bool
	watchDir  bool
)

func init() {
	flag.StringVar(&token, "token", "", "authentication token to use")
	flag.StringVar(&rootDir, "dir", "", "root folder for replays")
	flag.StringVar(&hash, "hash", "", "hash of the account for the replays")
	flag.BoolVar(&uploadAll, "all", false,
		"ulpload all replays (not just the newest)")
	flag.BoolVar(&watchDir, "watch", false,
		"watch directory for new replays")
}

func getDirectories(root string) (dirs []string, err error) {
	err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("incoming error for %v: %v",
					path, err)
			}
			if info.IsDir() {
				dirs = append(dirs, path)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("error when walking %v: %v",
			root, err)
	}
	return dirs, nil
}

func uploadFile(path, hash, token string) {
	log.Printf("uploading %v", filepath.Base(path))
	result, err := uploader.UploadReplay(hash, token, path)
	if err != nil {
		log.Fatalf("error while uploading replay: %v", err)
	}
	log.Printf("queued uploaded file with id %v", result.QueueID)
}

func handleWatch(watcher *fsnotify.Watcher, hash, token string) {
	fileByteCount := make(map[string]int64)
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				isReplay, err := uploader.IsReplayFile(event.Name)
				if err != nil {
					log.Fatalf("error while checking for replay(%v): %v",
						event.Name, err)
				}
				if isReplay {
					info, err := os.Stat(event.Name)
					if err != nil {
						log.Fatalf("error while getting size of file (%v): %v",
							event.Name, err)
					}
					size := info.Size()
					if fileByteCount[event.Name] != size {
						fileByteCount[event.Name] = size
						uploadFile(event.Name, hash, token)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Fatal(err)
		}
	}
}

func main() {
	// read configuration and setup
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

	// setup directory watcher
	var watcher *fsnotify.Watcher
	if watchDir {
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			log.Fatal("error while creating watcher")
		}
		defer watcher.Close()

		dirs, err := getDirectories(rootDir)
		if err != nil {
			log.Fatalf("error while gathering dirs (%v): %v",
				rootDir, err)
		}
		for _, dir := range dirs {
			log.Printf("watching %v", dir)
			watcher.Add(dir)
		}
	}

	// gather backlog files
	var files []*uploader.ReplayFile
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
	sort.Sort(uploader.ByDate(files))

	// upload backlog files
	for _, f := range files {
		uploadFile(f.Path, hash, token)
	}

	// upload new files
	if watchDir {
		go handleWatch(watcher, hash, token)
		fmt.Println("press any key to quit")
		os.Stdin.Read(make([]byte, 1))
	}
}
