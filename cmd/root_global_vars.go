package cmd

import (
	"sync"

	"github.com/rs/zerolog"
)

// Global Vars
var Ilogger, Elogger zerolog.Logger
var TermWidth, TermHeight int = 0, 0
var IsTerminal bool = false
var OS string
var WgGlb sync.WaitGroup
var ConfigPath string
var RootPath string

var GlobalMap map[string]string = make(map[string]string)

var Version string = "0.2.0"
var Input InputData = InputData{InputSlice: []string{},
	InputMap:   make(map[string][]string),
	InputTable: make([]map[string]string, 0),
}

// https://habr.com/ru/company/macloud/blog/558316/
var ColorReset string = "\033[0m"

var ColorRed string = "\033[31m"
var ColorGreen string = "\033[32m"
var ColorYellow string = "\033[33m"
var ColorBlue string = "\033[34m"
var ColorPurple string = "\033[35m"
var ColorCyan string = "\033[36m"
var ColorWhite string = "\033[37m"
