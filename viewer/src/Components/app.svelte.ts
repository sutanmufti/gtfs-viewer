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
} = $state({
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
