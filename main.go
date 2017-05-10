package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/ewhal/nyaa/cache"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/network"
	"github.com/ewhal/nyaa/router"
	"github.com/ewhal/nyaa/util/log"
	"github.com/ewhal/nyaa/util/signals"
	"github.com/nicksnyder/go-i18n/i18n"
)

func initI18N() {
	/* Initialize the languages translation */
	i18n.MustLoadTranslationFile("translations/en-us.all.json")
	paths, err := filepath.Glob("translations/*.json")
	if err == nil {
		for _, path := range paths {
			i18n.LoadTranslationFile(path)
		}
	}
}

func RunServer(conf *config.Config) {
	http.Handle("/", router.Router)

	// Set up server,
	srv := &http.Server{
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	l, err := network.CreateHTTPListener(conf)
	log.CheckError(err)
	if err == nil {
		// add http server to be closed gracefully
		signals.RegisterCloser(&network.GracefulHttpCloser{
			Server:   srv,
			Listener: l,
		})
		log.Infof("listening on %s", l.Addr())
		err := srv.Serve(l)
		if err != nil && err != network.ErrListenerStopped {
			log.CheckError(err)
		}

	}
}

var cpuProfileFile = flag.String("cpu", "", "cpu profiler dump file")

func main() {
	conf := config.New()
	processFlags := conf.BindFlags()
	defaults := flag.Bool("print-defaults", false, "print the default configuration file on stdout")
	flag.Float64Var(&cache.Size, "c", cache.Size, "size of the search cache in MB")
	flag.Parse()
	if *defaults {
		stdout := bufio.NewWriter(os.Stdout)
		err := conf.Pretty(stdout)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = stdout.Flush()
		if err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(0)
	} else {
		err := processFlags()
		if err != nil {
			log.CheckError(err)
		}
		db.ORM, err = db.GormInit(conf)
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(*cpuProfileFile) > 0 {
			fpath := *cpuProfileFile
			var f *os.File
			f, err = os.Create(fpath)
			if err != nil {
				log.Fatal(err.Error())
			}
			err = pprof.StartCPUProfile(f)
			if err != nil {
				log.Fatal(err.Error())
			}
			defer func() {
				pprof.StopCPUProfile()
				f.Sync()
				f.Close()
			}()
		}
		initI18N()
		go signals.Handle()
		if len(config.TorrentFileStorage) > 0 {
			err := os.MkdirAll(config.TorrentFileStorage, 0700)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		RunServer(conf)
	}
}
