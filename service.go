package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

const pageSize = 10
const uploadDir = "./uploads"

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// getGTFS retrieves a GTFS instance from the store by name (query param "gtfs").
func getGTFS(c *gin.Context) (*gtfsparser.GTFS, bool) {
	name := c.Query("gtfs")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'gtfs' query parameter"})
		return nil, false
	}
	storeMu.RLock()
	g, ok := store[name]
	storeMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("GTFS '%s' not found", name)})
		return nil, false
	}
	return g, true
}

// ListGTFS handles GET /gtfs/ — returns the names of all loaded GTFS feeds.
func ListGTFS(c *gin.Context) {
	storeMu.RLock()
	names := make([]string, 0, len(store))
	for k := range store {
		names = append(names, k)
	}
	storeMu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"gtfs": names})
}

// UploadGTFS handles POST /gtfs/upload — saves, parses, and stores a GTFS zip.
func UploadGTFS(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'file' form field"})
		return
	}
	defer file.Close()

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create upload directory"})
		return
	}

	destPath := uploadDir + "/" + header.Filename
	dest, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}
	defer dest.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := dest.Write(buf[:n]); writeErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error writing file"})
				return
			}
		}
		if readErr != nil {
			break
		}
	}
	dest.Close()

	g := &gtfsparser.GTFS{FileName: destPath}
	if err := g.ParseAll(); err != nil {
		os.Remove(destPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if c.Query("validate") == "true" {
		if validationErrs := g.ValidateAll(); len(validationErrs) > 0 {
			os.Remove(destPath)
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validationErrs})
			return
		}
	}

	g.Compile()

	storeMu.Lock()
	store[header.Filename] = g
	storeMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "uploaded", "name": header.Filename})
}

// stopToFeature converts a Stop to a GeoJSON Feature.
func stopToFeature(s gtfsparser.Stop) map[string]any {
	var coords [2]float64
	if s.StopLon != nil && s.StopLat != nil {
		coords[0] = *s.StopLon
		coords[1] = *s.StopLat
	}
	return map[string]any{
		"type": "Feature",
		"geometry": map[string]any{
			"type":        "Point",
			"coordinates": coords,
		},
		"properties": map[string]any{
			"stop_id":   s.StopID,
			"stop_name": s.StopName,
			"stop_code": s.StopCode,
			"stop_desc": s.StopDesc,
		},
	}
}

// GetGTFSFile handles GET /gtfs/files/:fileName — returns paginated file content.
// For "stops", content is returned as GeoJSON.
func GetGTFSFile(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	fileName := c.Param("fileName")

	switch fileName {
	case "stops":
		stops := g.StopData
		offset, limit, totalPages := paginate(len(stops), page)
		features := make([]map[string]any, 0, limit)
		for i := offset; i < offset+limit; i++ {
			features = append(features, stopToFeature(stops[i]))
		}
		c.JSON(http.StatusOK, gin.H{
			"type":       "FeatureCollection",
			"features":   features,
			"page":       page,
			"total":      len(stops),
			"totalPages": totalPages,
		})

	case "routes":
		routes := g.RouteData
		offset, limit, totalPages := paginate(len(routes), page)
		c.JSON(http.StatusOK, gin.H{"data": routes[offset : offset+limit], "page": page, "total": len(routes), "totalPages": totalPages})

	case "trips":
		trips := g.TripData
		offset, limit, totalPages := paginate(len(trips), page)
		c.JSON(http.StatusOK, gin.H{"data": trips[offset : offset+limit], "page": page, "total": len(trips), "totalPages": totalPages})

	case "agency":
		agencies := g.AgencyData
		offset, limit, totalPages := paginate(len(agencies), page)
		c.JSON(http.StatusOK, gin.H{"data": agencies[offset : offset+limit], "page": page, "total": len(agencies), "totalPages": totalPages})

	case "calendar":
		calendars := g.CalendarData
		offset, limit, totalPages := paginate(len(calendars), page)
		c.JSON(http.StatusOK, gin.H{"data": calendars[offset : offset+limit], "page": page, "total": len(calendars), "totalPages": totalPages})

	case "calendar_dates":
		dates := g.CalendarDates
		offset, limit, totalPages := paginate(len(dates), page)
		c.JSON(http.StatusOK, gin.H{"data": dates[offset : offset+limit], "page": page, "total": len(dates), "totalPages": totalPages})

	case "shapes":
		shapes := g.ShapeData
		offset, limit, totalPages := paginate(len(shapes), page)
		c.JSON(http.StatusOK, gin.H{"data": shapes[offset : offset+limit], "page": page, "total": len(shapes), "totalPages": totalPages})

	case "frequencies":
		freqs := g.FrequencyData
		offset, limit, totalPages := paginate(len(freqs), page)
		c.JSON(http.StatusOK, gin.H{"data": freqs[offset : offset+limit], "page": page, "total": len(freqs), "totalPages": totalPages})

	case "transfers":
		transfers := g.TransferData
		offset, limit, totalPages := paginate(len(transfers), page)
		c.JSON(http.StatusOK, gin.H{"data": transfers[offset : offset+limit], "page": page, "total": len(transfers), "totalPages": totalPages})

	case "fare_attributes":
		fa := g.FareAttributes
		offset, limit, totalPages := paginate(len(fa), page)
		c.JSON(http.StatusOK, gin.H{"data": fa[offset : offset+limit], "page": page, "total": len(fa), "totalPages": totalPages})

	case "fare_rules":
		fr := g.FareRules
		offset, limit, totalPages := paginate(len(fr), page)
		c.JSON(http.StatusOK, gin.H{"data": fr[offset : offset+limit], "page": page, "total": len(fr), "totalPages": totalPages})

	case "feed_info":
		fi := g.FeedInfo
		offset, limit, totalPages := paginate(len(fi), page)
		c.JSON(http.StatusOK, gin.H{"data": fi[offset : offset+limit], "page": page, "total": len(fi), "totalPages": totalPages})

	case "pathways":
		pw := g.PathwayData
		offset, limit, totalPages := paginate(len(pw), page)
		c.JSON(http.StatusOK, gin.H{"data": pw[offset : offset+limit], "page": page, "total": len(pw), "totalPages": totalPages})

	case "levels":
		lv := g.LevelData
		offset, limit, totalPages := paginate(len(lv), page)
		c.JSON(http.StatusOK, gin.H{"data": lv[offset : offset+limit], "page": page, "total": len(lv), "totalPages": totalPages})

	case "attributions":
		attr := g.Attributions
		offset, limit, totalPages := paginate(len(attr), page)
		c.JSON(http.StatusOK, gin.H{"data": attr[offset : offset+limit], "page": page, "total": len(attr), "totalPages": totalPages})

	case "translations":
		tr := g.Translations
		offset, limit, totalPages := paginate(len(tr), page)
		c.JSON(http.StatusOK, gin.H{"data": tr[offset : offset+limit], "page": page, "total": len(tr), "totalPages": totalPages})

	default:
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("unknown file '%s'", fileName)})
	}
}

// GetGTFSFileRecord handles GET /gtfs/files/:fileName/:id — returns a single record.
func GetGTFSFileRecord(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	fileName := c.Param("fileName")
	id := c.Param("id")

	switch fileName {
	case "stops":
		for _, s := range g.StopData {
			if s.StopID == id {
				c.JSON(http.StatusOK, s)
				return
			}
		}
	case "routes":
		for _, r := range g.RouteData {
			if r.RouteID == id {
				c.JSON(http.StatusOK, r)
				return
			}
		}
	case "trips":
		for _, t := range g.TripData {
			if t.TripID == id {
				c.JSON(http.StatusOK, t)
				return
			}
		}
	case "agency":
		for _, a := range g.AgencyData {
			if a.AgencyID == id {
				c.JSON(http.StatusOK, a)
				return
			}
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file '%s' does not support record lookup by ID", fileName)})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("record '%s' not found in '%s'", id, fileName)})
}

// GetStops handles GET /gtfs/files/stops — returns all stops as a GeoJSON FeatureCollection.
func GetStops(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	// If trip= is provided, resolve the stop list from TripStopTimes.
	stops := g.StopData
	if tripID := c.Query("trip"); tripID != "" {
		var tripPtr *gtfsparser.Trip
		for i := range g.TripData {
			if g.TripData[i].TripID == tripID {
				tripPtr = &g.TripData[i]
				break
			}
		}
		if tripPtr == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("trip '%s' not found", tripID)})
			return
		}
		stopTimes := g.TripStopTimes[tripPtr]
		filtered := make([]gtfsparser.Stop, 0, len(stopTimes))
		for _, st := range stopTimes {
			if st.StopID != nil {
				filtered = append(filtered, *st.StopID)
			}
		}
		stops = filtered
	}

	if c.Query("geojson") == "true" {
		features := make([]map[string]any, 0, len(stops))
		for _, s := range stops {
			features = append(features, stopToFeature(s))
		}
		c.JSON(http.StatusOK, gin.H{
			"type":     "FeatureCollection",
			"features": features,
		})
		return
	}

	if c.Query("trip") != "" {
		c.JSON(http.StatusOK, gin.H{
			"data":       stops,
			"total":      len(stops),
			"totalPages": 1,
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset, limit, totalPages := paginate(len(stops), page)
	c.JSON(http.StatusOK, gin.H{
		"data":       stops[offset : offset+limit],
		"page":       page,
		"total":      len(stops),
		"totalPages": totalPages,
	})
}

// GetStop handles GET /gtfs/stop/:stopId — returns routes that serve the stop,
// including routes reachable via transfers.
func GetStop(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	stopID := c.Param("stopId")

	// Find the *Stop pointer.
	var stopPtr *gtfsparser.Stop
	for i := range g.StopData {
		if g.StopData[i].StopID == stopID {
			stopPtr = &g.StopData[i]
			break
		}
	}
	if stopPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("stop '%s' not found", stopID)})
		return
	}

	// Collect routes for this stop.
	routeSet := make(map[string]*gtfsparser.Route)
	for _, r := range g.StopRoutes[stopPtr] {
		routeSet[r.RouteID] = r
	}

	// Collect routes from transfer stops.
	transfers := g.TransfersFromStop[stopPtr]
	transferStops := make([]gtfsparser.Stop, 0)
	for _, tr := range transfers {
		if tr.ToStopID != nil {
			transferStops = append(transferStops, *tr.ToStopID)
			for _, r := range g.StopRoutes[tr.ToStopID] {
				routeSet[r.RouteID] = r
			}
		}
	}

	routes := make([]*gtfsparser.Route, 0, len(routeSet))
	for _, r := range routeSet {
		routes = append(routes, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"stop":           stopPtr,
		"routes":         routes,
		"transfers":      transfers,
		"transfer_stops": transferStops,
	})
}

// GetRoute handles GET /gtfs/route/:routeId — returns trips on the route.
func GetRoute(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	routeID := c.Param("routeId")

	var routePtr *gtfsparser.Route
	for i := range g.RouteData {
		if g.RouteData[i].RouteID == routeID {
			routePtr = &g.RouteData[i]
			break
		}
	}
	if routePtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("route '%s' not found", routeID)})
		return
	}

	trips := g.RouteTrips[routePtr]
	c.JSON(http.StatusOK, gin.H{"route": routePtr, "trips": trips})
}

// GetTrip handles GET /gtfs/trip/:tripId — returns stop times and frequencies for the trip.
func GetTrip(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	tripID := c.Param("tripId")

	var tripPtr *gtfsparser.Trip
	for i := range g.TripData {
		if g.TripData[i].TripID == tripID {
			tripPtr = &g.TripData[i]
			break
		}
	}
	if tripPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("trip '%s' not found", tripID)})
		return
	}

	stopTimes := g.TripStopTimes[tripPtr]
	frequencies := g.FrequenciesByTrip[tripPtr]

	if c.Query("geojson") == "true" {
		coords := make([][2]float64, 0, len(stopTimes))
		for _, st := range stopTimes {
			if st.StopID == nil || st.StopID.StopLon == nil || st.StopID.StopLat == nil {
				continue
			}
			coords = append(coords, [2]float64{*st.StopID.StopLon, *st.StopID.StopLat})
		}

		props := map[string]any{
			"trip_id":  tripPtr.TripID,
			"headsign": tripPtr.TripHeadsign,
		}
		if tripPtr.RouteID != nil {
			props["route_id"] = tripPtr.RouteID.RouteID
			props["route_short_name"] = tripPtr.RouteID.RouteShortName
			props["route_long_name"] = tripPtr.RouteID.RouteLongName
			props["route_color"] = tripPtr.RouteID.RouteColor
		}

		c.JSON(http.StatusOK, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "LineString",
				"coordinates": coords,
			},
			"properties": props,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trip":        tripPtr,
		"stop_times":  stopTimes,
		"frequencies": frequencies,
	})
}

// TripGeojson handles GET /gtfs/trip — returns all trips as a GeoJSON
// FeatureCollection of LineString features, each constructed from the ordered
// stop coordinates in TripStopTimes.
func TripGeojson(c *gin.Context) {
	g, ok := getGTFS(c)
	if !ok {
		return
	}

	features := make([]map[string]any, 0, len(g.TripStopTimes))

	for trip, stopTimes := range g.TripStopTimes {
		// Build coordinate sequence from stop positions.
		coords := make([][2]float64, 0, len(stopTimes))
		for _, st := range stopTimes {
			if st.StopID == nil || st.StopID.StopLon == nil || st.StopID.StopLat == nil {
				continue
			}
			coords = append(coords, [2]float64{*st.StopID.StopLon, *st.StopID.StopLat})
		}

		// A LineString requires at least 2 points.
		if len(coords) < 2 {
			continue
		}

		props := map[string]any{
			"trip_id":   trip.TripID,
			"headsign":  trip.TripHeadsign,
			"direction": int(trip.DirectionID),
		}
		if trip.RouteID != nil {
			props["route_id"] = trip.RouteID.RouteID
			props["route_short_name"] = trip.RouteID.RouteShortName
			props["route_long_name"] = trip.RouteID.RouteLongName
			props["route_color"] = trip.RouteID.RouteColor
		}

		features = append(features, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "LineString",
				"coordinates": coords,
			},
			"properties": props,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "FeatureCollection",
		"features": features,
	})
}
