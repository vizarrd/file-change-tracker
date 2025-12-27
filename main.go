package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchList defines directories to monitor
var watchList = []string{
	"/etc",
	"/bin",
	"/sbin",
	"/usr/bin",
	"/usr/sbin",
	"/boot",
	"/var/www",
	"/etc/apache2",
	"/home",
}

// watchedSet tracks watched directories
type watchedSet struct {
	mu sync.Mutex
	m  map[string]struct{}
}

func newWatchedSet() *watchedSet {
	return &watchedSet{m: make(map[string]struct{})}
}

func (ws *watchedSet) has(path string) bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	_, ok := ws.m[path]
	return ok
}

func (ws *watchedSet) add(path string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.m[path] = struct{}{}
}

func (ws *watchedSet) remove(path string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	delete(ws.m, path)
}

// addAllDirs walks root and adds every directory to watcher
func addAllDirs(w *fsnotify.Watcher, root string, watched *watchedSet) error {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && !watched.has(path) {
			if err := w.Add(path); err == nil {
				watched.add(path)
			}
		}
		return nil
	})
}

func main() {
	// Ensure log directory exists
	_ = os.MkdirAll("/var/log/filetracker", 0700)

	// Audit log setup
	logFile, err := os.OpenFile(
		"/var/log/filetracker/filetracker.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)

	approvePath := flag.String("approve", "", "Approve file change and update baseline")
	flag.Parse()

	// Approval mode
	if *approvePath != "" {
		if os.Geteuid() != 0 {
			log.Fatal("Approval requires root privileges")
		}
		loadBaseline()
		approveFile(*approvePath)
		return
	}

	// Normal service mode
	loadBaseline()
	if isBaselineEmpty() {
		populateBaseline(watchList)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	tracked := newWatchedSet()
	for _, root := range watchList {
		addAllDirs(watcher, root, tracked)
	}

	done := make(chan bool)

	go func() {
		lastEvent := make(map[string]time.Time)
		for {
			select {
			case event := <-watcher.Events:
				if t, found := lastEvent[event.Name]; found {
					if time.Since(t) < 200*time.Millisecond {
						continue
					}
				}
				lastEvent[event.Name] = time.Now()

				if event.Op&fsnotify.Write == fsnotify.Write {
					handleFileChange(event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						addAllDirs(watcher, event.Name, tracked)
					}
				}
				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					if tracked.has(event.Name) {
						watcher.Remove(event.Name)
						tracked.remove(event.Name)
					}
				}
			case err := <-watcher.Errors:
				_ = err // intentionally ignored
			}
		}
	}()

	<-done
}
