# GTFS Viewer


## About

GTFS Viewer is a go gin app that serves vite svelte assets to view GTFS (Generalised Transit Feed Specification). The user can run the server and serves the static assets (the UI). The UI is the interface to upload GTFS zip files to the go programme which are registered in the memory (`store`). Viewing a gtfs should be as simple as selecting an uploaded GTFS and toggling the controls in the UI.

Go is used as the backend app that handles the logic behind the app. We will use "github.com/sutanmufti/gtfs-parser" to parse the GTFS.


## File Structure

`service.go` is where we create the service functions.

`main.go` is the server function where we plug in the service functions to the appropriate endpoints.

## API Structure

Here are the endpoints

GET /gtfs/ -> lists available GTFS zip files that have been uploaded.
POST /gtfs/upload -> uploads a GTFS zip file to be processed by `gtfsparser`. `gtfsparser` will validate the zip file by calling `gtfsparser.GTFS.Compile()`, if not, return error. Do not run `gtfsparser.GTFS.ValidateAll()` unless query param `validate=true`.

GET /gtfs/files/:fileName?page=1 -> return the file content, paginated. each page contains 10 records. For stops, we return in geojson format. Look at the `/gtfs/files/stops` endpoint.
GET /gtfs/files/:fileName/:id -> return information on a record of the file. Not used for gtfs files that don't have primary keys.
GET /gtfs/files/stops -> returns all stops in geojson format. `FeatureCollection<Feature<Point>>`. This is to be displayed by the svelte viewer in a map format.

GET /gtfs/stop/:stopId -> returns routes that stop including routes in transfers. this is obtained by `gtfsparser.GTFS.StopRoutes` and `gtfsparser.GTFS.TransfersFromStop`. After running `gtfsparser.GTFS.TransfersFromStop`, go obtains additional stops that can be reached from the `:stopId`. go will return all routes on the stops identified in `gtfsparser.GTFS.TransfersFromStop`.

GET /gtfs/route/:routeId -> returns trips on the route `gtfsparser.GTFS.RouteTrips`.
GET /gtfs/trip/:tripId -> returns stopTimes on the trip `gtfsparser.GTFS.TripStopTimes`. also returns the frequencies from `gtfsparser.GTFS.FrequenciesByTrip`


## Building the App

To build the app, first we must run `npm run build` on the `./viewer`. This generates the static assets. Then we'll embed into our go programme. For development, we do not need to do this as we can run 2 separate servers one for go and one for the vite svelte.