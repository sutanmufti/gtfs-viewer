package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

//go:embed viewer/dist
var distFS embed.FS

var (
	store   = map[string]*gtfsparser.GTFS{}
	storeMu sync.RWMutex
)

var WithFile = false

func main() {
	filePath := flag.String("f", "", "path to a GTFS zip file to load on startup")
	flag.Parse()

	if *filePath != "" {
		g, err := loadGTFS(*filePath)
		if err != nil {
			log.Fatalf("failed to parse GTFS file %q: %v", *filePath, err)
		}
		name := filepath.Base(*filePath)
		storeMu.Lock()
		store[name] = g
		storeMu.Unlock()
		WithFile = true
		fmt.Printf("Loaded GTFS file: %s\n", name)
	}

	r := gin.Default()
	r.Use(cors.Default())

	// Trust X-Forwarded-* headers from proxies.
	// TRUSTED_PROXIES is a comma-separated list of CIDRs/IPs.
	// Defaults to loopback and private network ranges.
	trustedProxies := os.Getenv("TRUSTED_PROXIES")
	if trustedProxies == "" {
		trustedProxies = "127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
	}
	if err := r.SetTrustedProxies(strings.Split(trustedProxies, ",")); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
	}
	r.ForwardedByClientIP = true

	// Serve embedded frontend at /.
	sub, err := fs.Sub(distFS, "viewer/dist")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}
	r.NoRoute(func(c *gin.Context) {
		filePath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if filePath == "" {
			filePath = "index.html"
		}
		f, err := sub.Open(filePath)
		if err != nil {
			c.FileFromFS("index.html", http.FS(sub))
			return
		}
		stat, err := f.Stat()
		f.Close()
		if err != nil || stat.IsDir() {
			c.FileFromFS("index.html", http.FS(sub))
			return
		}
		http.FileServer(http.FS(sub)).ServeHTTP(c.Writer, c.Request)
	})

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
