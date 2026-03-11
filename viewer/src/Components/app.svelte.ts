// The GTFS file types available to browse.
export const GTFS_FILES = [
  'stops', 'routes', 'trips', 'agency',
  'calendar', 'calendar_dates', 'shapes',
  'frequencies', 'transfers', 'feed_info',
  'fare_attributes', 'fare_rules', 'pathways',
  'levels', 'attributions', 'translations',
] as const

// Shared state across components.
export const appstate: {
  mapdiv?: HTMLDivElement
  map?: mapboxgl.Map

  // Uploaded GTFS zip names from GET /gtfs/
  gtfsZipFiles: string[]
  // Currently selected GTFS feed name
  selectedGtfs: string

  // Currently selected file within the feed
  selectedFile: string

  // Paginated viewer data
  viewerData: Record<string, unknown>[]
  viewerPage: number
  viewerTotal: number
  viewerTotalPages: number
  viewerLoading: boolean

  activeRecord?: Record<string, unknown>

  stopShowRoute: boolean
} = $state({
  stopShowRoute: false,
  gtfsZipFiles: [],
  selectedGtfs: '',
  selectedFile: '',
  viewerData: [],
  viewerPage: 1,
  viewerTotal: 0,
  viewerTotalPages: 1,
  viewerLoading: false,
})

export async function fetchGtfsList() {
  const res = await fetch('/gtfs/')
  const data = await res.json()
  appstate.gtfsZipFiles = data.gtfs ?? []
}

export async function loadViewerPage(file: string, page: number) {
  if (!appstate.selectedGtfs) return
  appstate.viewerLoading = true
  appstate.selectedFile = file
  appstate.viewerPage = page
  try {
    const res = await fetch(`/gtfs/files/${file}?gtfs=${encodeURIComponent(appstate.selectedGtfs)}&page=${page}`)
    const data = await res.json()
    appstate.viewerData = data.data ?? []
    appstate.viewerTotal = data.total ?? 0
    appstate.viewerTotalPages = data.totalPages ?? 1
  } finally {
    appstate.viewerLoading = false
  }
}

export async function loadStopsOnMap() {
  const map = appstate.map
  if (!map || !appstate.selectedGtfs) return

  const res = await fetch(`/gtfs/files/stops?gtfs=${encodeURIComponent(appstate.selectedGtfs)}&geojson=true`)
  const geojson = await res.json()

  const SOURCE_ID = 'gtfs-stops'
  const LAYER_ID = 'gtfs-stops-layer'

  if (map.getSource(SOURCE_ID)) {
    ;(map.getSource(SOURCE_ID) as mapboxgl.GeoJSONSource).setData(geojson)
  } else {
    map.addSource(SOURCE_ID, { type: 'geojson', data: geojson })
    map.addLayer({
      id: LAYER_ID,
      type: 'circle',
      source: SOURCE_ID,
      paint: {
        'circle-radius': 5,
        'circle-color': '#3b82f6',
        'circle-stroke-width': 1,
        'circle-stroke-color': '#1d4ed8',
      },
    })

    // Fit map to stops bounds.
    const coords: [number, number][] = geojson.features
      .map((f: any) => f.geometry.coordinates as [number, number])
      .filter((c: [number, number]) => c[0] !== 0 || c[1] !== 0)

    if (coords.length > 0) {
      const bounds = coords.reduce(
        (b, c) => b.extend(c),
        new (await import('mapbox-gl')).default.LngLatBounds(coords[0], coords[0])
      )
      map.fitBounds(bounds, { padding: 40 })
    }
  }
}

export async function loadTripStoptimesOnMap() {
  const map = appstate.map
  if (!map || !appstate.selectedGtfs) return

  const res = await fetch(`/gtfs/trip?gtfs=${encodeURIComponent(appstate.selectedGtfs)}`)
  const geojson = await res.json()

  const SOURCE_ID = 'gtfs-trips'
  const LAYER_ID = 'gtfs-trips-layer'

  if (map.getSource(SOURCE_ID)) {
    ;(map.getSource(SOURCE_ID) as mapboxgl.GeoJSONSource).setData(geojson)
  } else {
    map.addSource(SOURCE_ID, { type: 'geojson', data: geojson })
    // Add trip lines below the stops layer so stops render on top.
    const stopsLayerId = map.getLayer('gtfs-stops-layer') ? 'gtfs-stops-layer' : undefined
    map.addLayer(
      {
        id: LAYER_ID,
        type: 'line',
        source: SOURCE_ID,
        layout: { 'line-join': 'round', 'line-cap': 'round' },
        paint: {
          // Use route_color (hex without #) when present, otherwise a neutral grey.
          'line-color': [
            'case',
            ['!=', ['get', 'route_color'], ''],
            ['concat', '#', ['get', 'route_color']],
            '#94a3b8',
          ],
          'line-width': 2,
          'line-opacity': 0.8,
        },
      },
      stopsLayerId,
    )
  }
}


export function switchLayer(map: mapboxgl.Map,layerName: string, hide: boolean){
  const layer = map.getLayer(layerName);
  (hide && layer) ? map.setLayoutProperty(layerName, "visibility", "none") : map.setLayoutProperty(layerName, "visibility", "visible")
}



export function clearFilter(){
  
  appstate.map!.setFilter("gtfs-stops-layer", null)
  appstate.map!.setFilter("gtfs-trips-layer", null)

}

// set to default after showRoutes
export function clearStopView(){

  clearFilter()
  appstate.stopShowRoute = false
}

export async function showRoutes(stop_id: string) {
    const map = appstate.map
    const gtfs = appstate.selectedGtfs
    if (!map || !gtfs) return

    const enc = (s: string) => encodeURIComponent(s)

    // 1. Get routes for this stop.
    const stopRes = await fetch(`/gtfs/stop/${enc(stop_id)}?gtfs=${enc(gtfs)}`)
    const stopData = await stopRes.json()
    const routes: { RouteID: string }[] = stopData.routes ?? []

    // 2. Get trips for each route in parallel.
    const routeResults = await Promise.all(
      routes.map(r =>
        fetch(`/gtfs/route/${enc(r.RouteID)}?gtfs=${enc(gtfs)}`).then(res => res.json())
      )
    )

    // 3. Collect all trip IDs and filter the trips layer.
    const tripIds: string[] = routeResults.flatMap(r =>
      (r.trips ?? []).map((t: { TripID: string }) => t.TripID)
    )
    if (map.getLayer('gtfs-trips-layer')) {
      map.setFilter('gtfs-trips-layer', ['in', ['get', 'trip_id'], ['literal', tripIds]])
    }

    // 4. Fetch stops using the first trip of each route (representative — trips on
    //    the same route share the same stops), collect unique stop IDs.
    const representativeTripIds: string[] = routeResults
      .map(r => r.trips?.[0]?.TripID as string | undefined)
      .filter((id): id is string => !!id)

    const stopsResults = await Promise.all(
      representativeTripIds.map(id =>
        fetch(`/gtfs/files/stops?gtfs=${enc(gtfs)}&trip=${enc(id)}`).then(res => res.json())
      )
    )

    const stopIds = [...new Set<string>(
      stopsResults.flatMap(r => (r.data ?? []).map((s: { StopID: string }) => s.StopID))
    )]
    if (map.getLayer('gtfs-stops-layer')) {
      map.setFilter('gtfs-stops-layer', ['in', ['get', 'stop_id'], ['literal', stopIds]])
    }

    appstate.stopShowRoute = true
  }