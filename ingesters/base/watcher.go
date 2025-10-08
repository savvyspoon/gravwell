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

	"github.com/fsnotify/fsnotify"
)

type ConfigWatcher struct {
	watcher  *fsnotify.Watcher
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
		file:     file,
		overlays: overlays,
	}
	return
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
func (c *ConfigWatcher) Signal() <-chan fsnotify.Event {
	if c == nil || c.watcher == nil {
		return nil
	}
	return c.watcher.Events
}
