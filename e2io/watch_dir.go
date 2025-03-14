package e2io

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

var mu sync.Mutex

func WatchDir(path string, callback func(string, fsnotify.Event)) {
	logrus.Infof("Watching directory: %s", path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Errorf("Error creating watcher error=%v", err)
		return
	}
	defer watcher.Close()

	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("Error accessing path=%v error=%v", p, err)
			return nil
		}
		if info.IsDir() {
			err := watcher.Add(p)
			if err != nil {
				logrus.Errorf("Error adding directory to watcher, directory=%v, error=%v", p, err)
			}
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("Error walking directory error=%v", err)
		return
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			mu.Lock()
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				debounceEvent(event.Name, callback, event)
			}

			if event.Op&fsnotify.Create != 0 {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() {
					_ = watcher.Add(event.Name)
				}
			}

			mu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.Errorf("Watcher error=%v", err)
		}
	}
}

var eventTimers = make(map[string]*time.Timer)

var tu sync.Mutex

func debounceEvent(filename string, callback func(string, fsnotify.Event), event fsnotify.Event) {
	tu.Lock()
	defer tu.Unlock()
	if strings.HasSuffix(filename, "~") {
		return
	}
	eventKey := filename + event.Name

	if timer, exists := eventTimers[eventKey]; exists {
		timer.Reset(500 * time.Millisecond)
		return
	}

	eventTimers[eventKey] = time.AfterFunc(500*time.Millisecond, func() {
		tu.Lock()
		delete(eventTimers, eventKey)
		tu.Unlock()
		logrus.Infof("Callback by File event: %s, Op: %s\n", event.Name, event.Op)
		callback(eventKey, event)
	})
}
