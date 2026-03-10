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
  viewerLoading: boolean
} = $state({
  gtfsZipFiles: [],
  selectedGtfs: '',
  selectedFile: '',
  viewerData: [],
  viewerPage: 1,
  viewerTotal: 0,
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
    // Stops return GeoJSON; everything else returns { data, total }
    if (file === 'stops') {
      appstate.viewerData = (data.features ?? []).map((f: any) => f.properties)
      appstate.viewerTotal = data.total ?? appstate.viewerData.length
    } else {
      appstate.viewerData = data.data ?? []
      appstate.viewerTotal = data.total ?? 0
    }
  } finally {
    appstate.viewerLoading = false
  }
}

export async function loadStopsOnMap() {
  const map = appstate.map
  if (!map || !appstate.selectedGtfs) return

  const res = await fetch(`/gtfs/files/stops?gtfs=${encodeURIComponent(appstate.selectedGtfs)}`)
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
