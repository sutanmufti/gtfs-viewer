<script lang="ts">
  import { appstate } from './app.svelte'

  let {
    stop_id,
    stop_name,
    stop_code,
    stop_desc,
    ParentStation,
  }: {
    stop_id: string
    stop_name: string
    stop_code?: string
    stop_desc?: string
    ParentStation: string
  } = $props()

  async function showRoutes() {
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
  }
</script>

<strong>{stop_name || stop_id}</strong>

{#if stop_code}
<br/>→ {stop_code}
{/if}

{#if stop_desc}
<br/>{stop_desc}
{/if}

<br/><span style="font-size:0.75em;color:#666">{ParentStation}</span>

<br/><button onclick={showRoutes}>Show Routes</button>
