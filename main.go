package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

var (
	store   = map[string]*gtfsparser.GTFS{}
	storeMu sync.RWMutex
)

var WithFile = false

func main() {
	filePath := flag.String("f", "", "path to a GTFS zip file to load on startup")
	flag.Parse()

	if *filePath != "" {
		g := &gtfsparser.GTFS{FileName: *filePath}
		if err := g.ParseAll(); err != nil {
			log.Fatalf("failed to parse GTFS file %q: %v", *filePath, err)
		}
		g.Compile()
		name := filepath.Base(*filePath)
		storeMu.Lock()
		store[name] = g
		storeMu.Unlock()
		WithFile = true
		fmt.Printf("Loaded GTFS file: %s\n", name)
	}

	r := gin.Default()

	r.GET("/ping", Ping)

	gtfs := r.Group("/gtfs")
	{
		gtfs.GET("/", ListGTFS)
		gtfs.GET("/config", GetConfig)
		gtfs.POST("/upload", UploadGTFS)

		// Static route must be registered before the parameterized one.
		gtfs.GET("/files/stops", GetStops)
		gtfs.GET("/files/:fileName", GetGTFSFile)
		gtfs.GET("/files/:fileName/:id", GetGTFSFileRecord)

		gtfs.GET("/stop/:stopId", GetStop)
		gtfs.GET("/route/:routeId", GetRoute)
		gtfs.GET("/trip/:tripId", GetTrip)

		gtfs.GET("/trip", TripGeojson)
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(host + ":" + port)
}
