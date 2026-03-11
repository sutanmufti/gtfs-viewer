package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

var (
	store   = map[string]*gtfsparser.GTFS{}
	storeMu sync.RWMutex
)

func main() {
	r := gin.Default()

	r.GET("/ping", Ping)

	gtfs := r.Group("/gtfs")
	{
		gtfs.GET("/", ListGTFS)
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

	r.Run()
}
