package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

const testFeedName = "sample.zip"

// buildFixtureZip creates a minimal valid GTFS zip in memory and returns its bytes.
func buildFixtureZip() ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	files := map[string]string{
		"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n" +
			"AG1,Test Agency,https://example.com,Australia/Sydney\n",

		"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type,route_color\n" +
			"R1,AG1,101,City Loop,3,FF0000\n",

		"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
			"SVC1,1,1,1,1,1,0,0,20240101,20241231\n",

		"trips.txt": "route_id,service_id,trip_id,trip_headsign,direction_id\n" +
			"R1,SVC1,T1,Central,0\n" +
			"R1,SVC1,T2,Airport,1\n",

		"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n" +
			"S1,Stop One,-33.8688,151.2093\n" +
			"S2,Stop Two,-33.8700,151.2100\n" +
			"S3,Stop Three,-33.8720,151.2120\n",

		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
			"T1,08:00:00,08:00:00,S1,1\n" +
			"T1,08:10:00,08:10:00,S2,2\n" +
			"T1,08:20:00,08:20:00,S3,3\n" +
			"T2,09:00:00,09:00:00,S3,1\n" +
			"T2,09:10:00,09:10:00,S2,2\n" +
			"T2,09:20:00,09:20:00,S1,3\n",

		"transfers.txt": "from_stop_id,to_stop_id,transfer_type\n" +
			"S1,S2,0\n",
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Write fixture zip to a temp file so gtfsparser can read it.
	zipBytes, err := buildFixtureZip()
	if err != nil {
		panic("failed to build fixture zip: " + err.Error())
	}
	tmp, err := os.CreateTemp("", "gtfs-test-*.zip")
	if err != nil {
		panic("failed to create temp file: " + err.Error())
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(zipBytes); err != nil {
		panic("failed to write fixture zip: " + err.Error())
	}
	tmp.Close()

	// Parse and compile the fixture into the global store.
	g, err := loadGTFS(tmp.Name())
	if err != nil {
		panic("failed to load fixture GTFS: " + err.Error())
	}
	storeMu.Lock()
	store[testFeedName] = g
	storeMu.Unlock()

	os.Exit(m.Run())
}

// setupRouter returns a gin router wired with all routes (no embedded FS needed).
func setupRouter() *gin.Engine {
	r := gin.New()
	r.GET("/ping", Ping)
	gtfs := r.Group("/gtfs")
	{
		gtfs.GET("/", ListGTFS)
		gtfs.GET("/config", GetConfig)
		gtfs.GET("/files/stops", GetStops)
		gtfs.GET("/files/:fileName", GetGTFSFile)
		gtfs.GET("/files/:fileName/:id", GetGTFSFileRecord)
		gtfs.GET("/stop/:stopId", GetStop)
		gtfs.GET("/route/:routeId", GetRoute)
		gtfs.GET("/trip/:tripId", GetTrip)
		gtfs.GET("/trip", TripGeojson)
	}
	return r
}

func do(r *gin.Engine, method, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return m
}

// --- Tests ---

func TestPing(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/ping")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListGTFS(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	names, ok := body["gtfs"].([]any)
	if !ok || len(names) == 0 {
		t.Fatalf("expected non-empty gtfs list, got %v", body)
	}
	if names[0] != testFeedName {
		t.Errorf("expected feed %q, got %q", testFeedName, names[0])
	}
}

func TestGetConfig(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/config")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if _, ok := body["withFile"]; !ok {
		t.Error("response missing 'withFile' field")
	}
}

func TestGetStops_Paginated(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/stops?gtfs="+testFeedName+"&page=1")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["total"].(float64) != 3 {
		t.Errorf("expected 3 stops, got %v", body["total"])
	}
	if _, ok := body["data"]; !ok {
		t.Error("response missing 'data' field")
	}
	if body["totalPages"].(float64) != 1 {
		t.Errorf("expected 1 page, got %v", body["totalPages"])
	}
}

func TestGetStops_GeoJSON(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/stops?gtfs="+testFeedName+"&geojson=true")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["type"] != "FeatureCollection" {
		t.Errorf("expected FeatureCollection, got %v", body["type"])
	}
	features := body["features"].([]any)
	if len(features) != 3 {
		t.Errorf("expected 3 features, got %d", len(features))
	}
}

func TestGetStops_ByTrip(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/stops?gtfs="+testFeedName+"&trip=T1")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	data := body["data"].([]any)
	if len(data) != 3 {
		t.Errorf("expected 3 stops for trip T1, got %d", len(data))
	}
}

func TestGetGTFSFile_Routes(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/routes?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["total"].(float64) != 1 {
		t.Errorf("expected 1 route, got %v", body["total"])
	}
}

func TestGetGTFSFile_Trips(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/trips?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["total"].(float64) != 2 {
		t.Errorf("expected 2 trips, got %v", body["total"])
	}
}

func TestGetGTFSFile_Unknown(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/nonexistent?gtfs="+testFeedName)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetGTFSFile_MissingGTFS(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/routes")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetGTFSFileRecord(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/stops/S1?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["StopID"] != "S1" {
		t.Errorf("expected StopID S1, got %v", body["StopID"])
	}
}

func TestGetGTFSFileRecord_NotFound(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/files/stops/NOPE?gtfs="+testFeedName)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetStop(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/stop/S1?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if _, ok := body["routes"]; !ok {
		t.Error("response missing 'routes' field")
	}
	if _, ok := body["stop"]; !ok {
		t.Error("response missing 'stop' field")
	}
}

func TestGetStop_NotFound(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/stop/NOPE?gtfs="+testFeedName)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetRoute(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/route/R1?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	trips, ok := body["trips"].([]any)
	if !ok || len(trips) == 0 {
		t.Errorf("expected trips, got %v", body["trips"])
	}
}

func TestGetRoute_NotFound(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/route/NOPE?gtfs="+testFeedName)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetTrip(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/trip/T1?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if _, ok := body["stop_times"]; !ok {
		t.Error("response missing 'stop_times'")
	}
	if _, ok := body["frequencies"]; !ok {
		t.Error("response missing 'frequencies'")
	}
}

func TestGetTrip_GeoJSON(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/trip/T1?gtfs="+testFeedName+"&geojson=true")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["type"] != "Feature" {
		t.Errorf("expected Feature, got %v", body["type"])
	}
	geom := body["geometry"].(map[string]any)
	if geom["type"] != "LineString" {
		t.Errorf("expected LineString, got %v", geom["type"])
	}
}

func TestGetTrip_NotFound(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/trip/NOPE?gtfs="+testFeedName)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTripGeojson(t *testing.T) {
	r := setupRouter()
	w := do(r, "GET", "/gtfs/trip?gtfs="+testFeedName)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := decode(t, w)
	if body["type"] != "FeatureCollection" {
		t.Errorf("expected FeatureCollection, got %v", body["type"])
	}
	features := body["features"].([]any)
	if len(features) != 2 {
		t.Errorf("expected 2 trip features, got %d", len(features))
	}
}
