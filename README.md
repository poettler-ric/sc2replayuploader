# Usage

```
$ sc2replayuploader --help
Usage of sc2replayuploader:
  -allowUnknownFlags
    	Don't terminate the app if ini file contains unknown flags.
  -config string
    	Path to ini config for using in go flags. May be relative to the current executable path.
  -configUpdateInterval duration
    	Update interval for re-reading config file set via -config flag. Zero disables config file re-reading.
  -dir string
    	root folder for replays
  -dumpflags
    	Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app.
  -hash string
    	hash of the account for the replays
  -token string
    	authentication token to use
```

# Example

```
$ sc2replayuploader -dumpflags >~/.sc2replayuploader.conf
$ vi ~/.sc2replayuploader.conf
$ sc2replayuploader
```

