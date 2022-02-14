package main

import (
	"fmt"

	"io"
	"os"
	"time"

	"encoding/json"

	flag "github.com/spf13/pflag"

	"github.com/hydronica/toml"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	. "glauth-ui-light/config"
	"glauth-ui-light/handlers"
	"glauth-ui-light/helpers"
	"glauth-ui-light/routes"
)

var log = logrus.New()

func confLog(cfg *WebConfig) {
	level := logrus.InfoLevel
	debug := cfg.Debug
	path := cfg.Logs.Path

	if debug {
		level = logrus.DebugLevel
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	log = &logrus.Logger{
		Out:   os.Stderr,
		Level: level,
		Formatter: &easy.Formatter{
			TimestampFormat: time.RFC3339,
			LogFormat:       "%lvl% - [%time%] %msg%\n",
		},
	}

	if path != "" && !debug {
		writer, _ := rotatelogs.New(
			path+"app.%Y%m%d",
			rotatelogs.WithLinkName(path),
			rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
			rotatelogs.WithMaxAge(-1),
			rotatelogs.WithRotationCount(cfg.Logs.RotationCount),
		)
		log.SetOutput(writer)
		gin.DefaultWriter = io.MultiWriter(writer)
	}
}

func main() {
	var Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage of %s\n%s\n\n", os.Args[0], handlers.Version)
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Usage = Usage

	// Parameters
	confPtr := flag.StringP("conf", "c", "", "Json config file")
	debugPtr := flag.BoolP("debug", "d", false, "Debug mode")
	flag.Parse()

	conf := *confPtr
	Debug := *debugPtr

	// Load config from file
	cfg := WebConfig{}
	_, err := toml.DecodeFile(conf, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError on mandatory config file:\n %s\n", err)
		Usage()
	}

	if Debug {
		fmt.Println("Config file:")
		b, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Print(string(b))
		fmt.Println("")
		cfg.Debug = true
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	confLog(&cfg)
	handlers.Log = *log //nolint

	DataRead, _, _ := helpers.ReadDB(&cfg)

	handlers.Data = DataRead

	r := routes.SetRoutes(&cfg)

	if cfg.SSL.Crt != "" {
		err = r.RunTLS(cfg.Port, cfg.SSL.Crt, cfg.SSL.Key)
	} else {
		fmt.Println("Server started. Version: " + handlers.Version)
		log.Println("Server started. Version: " + handlers.Version)
		err = r.Run(cfg.Port)
	}
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}
