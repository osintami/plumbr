// Copyright © 2022 Sloan Childers
package sink

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type IFileWatcher interface {
	Add(file string, refresh func()) error
	Listen()
}

type FileWatcher struct {
	watcher *fsnotify.Watcher
	watched map[string]func()
}

func NewFileWatcher() *FileWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error().Err(err).Str("component", "watcher").Msg("create new watcher")
		return nil
	}
	return &FileWatcher{
		watcher: watcher,
		watched: make(map[string]func()),
	}
}

func (x *FileWatcher) Add(file string, refresh func()) error {
	err := x.watcher.Add(file)
	if err != nil {
		log.Error().Err(err).Str("component", "watcher").Msg("add watch")
		return err
	}

	x.watched[file] = refresh
	return nil
}

func (x *FileWatcher) Listen() {
	go func() {
		done := make(chan bool)
		go func() {
			defer close(done)
			for {
				select {
				case event, ok := <-x.watcher.Events:
					if !ok {
						return
					}
					if event.Op == fsnotify.Chmod { // || event.Op == fsnotify.Write {
						log.Info().Str("component", "watcher").Str("name", event.Name).Str("op", event.Op.String()).Msg("file event")
						refresh := x.watched[event.Name]
						if refresh != nil {
							refresh()
						}
					}
				case err, ok := <-x.watcher.Errors:
					if !ok {
						return
					}
					log.Error().Err(err).Str("component", "watcher").Msg("errors")
				}
			}
		}()
		<-done
	}()
}
