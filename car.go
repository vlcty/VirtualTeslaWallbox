package main

import (
	"github.com/vlcty/TeslaWallbox"
)

type Car interface {
	charge() chan *teslaWallbox.Vitals
}
