package cmd

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	mcli_redis "mcli/packages/mcli-redis"
	mcli_utils "mcli/packages/mcli-utils"
)

// Global Vars
var Ilogger, Elogger zerolog.Logger
var TermWidth, TermHeight int = 0, 0
var IsTerminal bool = false
var IsVerbose bool = false
var IsRedis bool = false
var OS string
var WgGlb sync.WaitGroup
var ConfigPath string
var InputDataFromFile string
var RedisHost string
var RedisPort string
var RedisDb int
var RedisPwd string
var RedisRequire string
var CommonRedisStore *mcli_redis.RedisStore

var RootPath string

var Ctx context.Context
var CtxCancel context.CancelFunc
var Notify chan interface{} = make(chan interface{})

var GlobalMap map[string]string = make(map[string]string)
var GlobalCache mcli_utils.CCache

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
