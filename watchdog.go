package main

import (
	"fmt"
)

// WatchDog follows workshop mods and
// peridocally checks if there was new release to mods
// based on a date
type WatchDog struct {
	mods *Mods
}

func (w *WatchDog) Init(mods *Mods) error {
	if mods == nil {
		return fmt.Errorf("nil mods")
	}
	w.mods = mods
	return nil
}
