/*************************************************************************
 * Copyright 2020 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package base

import (
	"errors"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	signalOps fsnotify.Op = (fsnotify.Create | fsnotify.Rename | fsnotify.Remove | fsnotify.Write)

	debounceTime = time.Second // we will signal at most once per second
)

type ConfigWatcher struct {
	watcher  *fsnotify.Watcher
	outChan  chan struct{}
	file     string
	overlays string
}

// newConfigWatcher returns an active configuration watcher for a given config file and overlay directory
// either the file or overlays parameter may be empty, but not both.
func newConfigWatcher(file, overlays string) (cw ConfigWatcher, err error) {
	if file == `` && overlays == `` {
		err = errors.New("no paths provided")
		return
	}
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	if file != `` {
		if err = watcher.Add(file); err != nil {
			return
		}
	}
	if overlays != `` {
		if err = watcher.Add(overlays); err != nil {
			return
		}
	}
	cw = ConfigWatcher{
		watcher:  watcher,
		outChan:  make(chan struct{}),
		file:     file,
		overlays: overlays,
	}
	go cw.relay()
	return
}

// basic debouncer, we will only send a signal after things have gone quite for at least 1 second
// This will exit if the watcher closes
func (c *ConfigWatcher) relay() {
	//this probably isn't possible, but be paranoid
	if c == nil || c.watcher == nil || c.outChan == nil {
		return
	}

	var sig struct{}
	for event := range c.watcher.Events {
		if event.Op.Has(signalOps) {
			// ok, start consuming for our debounce time
			tmr := time.NewTimer(debounceTime)
		consumerLoop:
			for {
				select {
				case <-tmr.C:
					// go ahead and signal
					break consumerLoop
				case _, ok := <-c.watcher.Events:
					if !ok {
						break consumerLoop
					}
				}
			}
			tmr.Stop()
			c.outChan <- sig
		}
	}
}

// Close closes the watcher, no additional events will be sent
func (c *ConfigWatcher) Close() (err error) {
	if c == nil || c.watcher == nil {
		return // do nothing
	}
	//close the watcher and wait
	err = c.watcher.Close()
	return
}

// Signal returns a read-only reference on the fsnotify event channel if we are ready to do so, otherwise return nil
func (c *ConfigWatcher) Signal() <-chan struct{} {
	if c == nil || c.watcher == nil || c.outChan == nil {
		return nil
	}
	return c.outChan
}
