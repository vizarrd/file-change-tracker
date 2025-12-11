package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchList is the list of top-level paths you want to monitor.
var watchList = []string{
	"/etc",
	"/bin",
	"/sbin",
	"/usr/bin",
	"/usr/sbin",
	"/boot",
	"/var/www",
        "/home/viz/Desktop",
}

// watched keeps track of directories already added to the watcher
// to avoid duplicate Add() calls. Protected by a mutex for concurrency.
type watchedSet struct {
	mu sync.Mutex
	m  map[string]struct{}
}
func newWatchedSet() *watchedSet { return &watchedSet{m: make(map[string]struct{})} }
func (ws *watchedSet) has(path string) bool {
	ws.mu.Lock(); defer ws.mu.Unlock()
	_, ok := ws.m[path]
	return ok
}
func (ws *watchedSet) add(path string) {
	ws.mu.Lock(); defer ws.mu.Unlock()
	ws.m[path] = struct{}{}
}
func (ws *watchedSet) remove(path string) {
	ws.mu.Lock(); defer ws.mu.Unlock()
	delete(ws.m, path)
}

// addAllDirs walks root and adds every directory to watcher (best-effort).
// It uses watched to avoid duplicating Add calls.
func addAllDirs(w *fsnotify.Watcher, root string, watched *watchedSet) error {
	// If the root path doesn't exist or is not accessible, skip it gracefully.
	info, err := os.Stat(root)
	if err != nil {
		log.Printf("skip path (stat error): %s -> %v\n", root, err)
		return nil
	}
	if !info.IsDir() {
		log.Printf("skip path (not a directory): %s\n", root)
		return nil
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// permission or other error - skip that path but log it
			log.Printf("walk error: %v (path=%s)\n", err, path)
			return nil
		}
		if d.IsDir() {
			// Avoid duplicate adds
			if watched.has(path) {
				return nil
			}
			if err := w.Add(path); err != nil {
				// log error but continue
				log.Printf("failed to add watch on %s: %v\n", path, err)
			} else {
				watched.add(path)
				log.Printf("watching: %s\n", path)
			}
		}
		return nil
	})
}

func main() {
	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// trackedDirs prevents duplicate watcher.Add and helps remove on delete.
	tracked := newWatchedSet()

	// Add each configured root path recursively (best-effort).
	for _, root := range watchList {
		if err := addAllDirs(watcher, root, tracked); err != nil {
			log.Printf("error adding dirs for %s: %v", root, err)
		}
	}

	done := make(chan bool)

	go func() {
		// Simple debounce to reduce duplicate prints for quick events
		lastEvent := make(map[string]time.Time)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// debounce same path events within 200ms
				if t, found := lastEvent[event.Name]; found {
					if time.Since(t) < 200*time.Millisecond {
						continue
					}
				}
				lastEvent[event.Name] = time.Now()

				log.Printf("event: %s %s\n", event.Op.String(), event.Name)

				// If a new directory is created, add it (and its subtree) to watcher
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						// add new directory and its subdirectories
						if err := addAllDirs(watcher, event.Name, tracked); err != nil {
							log.Printf("error adding new dir %s: %v\n", event.Name, err)
						} else {
							log.Printf("added new directory to watcher: %s\n", event.Name)
						}
					}
				}

				// If a directory removed or renamed, attempt to remove it from watcher
				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					// Removing a watch is best-effort.
					if tracked.has(event.Name) {
						if err := watcher.Remove(event.Name); err != nil {
							// On many renames/removes the path may no longer exist -> Remove may error.
							log.Printf("failed to remove watcher for %s: %v\n", event.Name, err)
						} else {
							tracked.remove(event.Name)
							log.Printf("removed watcher for: %s\n", event.Name)
						}
					}
				}

				// Example reactions for file events (replace with hashing + auth logic later)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("[MODIFIED] %s\n", event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("[CREATED] %s\n", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Printf("[DELETED] %s\n", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v\n", err)
			}
		}
	}()

	log.Printf("Watching configured paths recursively: %v\n", watchList)
	<-done
}
