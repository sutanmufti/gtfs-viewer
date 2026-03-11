<script lang="ts">
  import { appstate } from './app.svelte'
  import { onMount, onDestroy } from 'svelte'

  type StopTime = {
    ArrivalTime: string
    DepartureTime: string
    StopSequence: number
    StopID: { StopID: string; StopName: string }
  }

  type Frequency = {
    StartTime: string
    EndTime: string
    HeadwaySecs: number
  }

  type TripDetail = {
    trip: {
      TripID: string
      TripHeadsign: string
      DirectionID: number
      RouteID: { RouteID: string; RouteShortName: string; RouteLongName: string; RouteColor: string }
    }
    stop_times: StopTime[]
    frequencies: Frequency[]
  }

  let detail = $state<TripDetail | null>(null)
  let loading = $state(true)
  let activeFeatureId = $state<number | null>(null)

  function setActiveStop(stopId: string | undefined) {
    const map = appstate.map
    if (!map || !stopId) return
    // Clear previous active feature.
    if (activeFeatureId !== null) {
      map.removeFeatureState({ source: 'gtfs-stops', id: activeFeatureId }, 'active')
    }
    // Find the numeric feature ID assigned by generateId.
    const features = map.querySourceFeatures('gtfs-stops', {
      filter: ['==', ['get', 'stop_id'], stopId],
    })
    if (features.length > 0 && features[0].id !== undefined) {
      activeFeatureId = features[0].id as number
      map.setFeatureState({ source: 'gtfs-stops', id: activeFeatureId }, { active: true })
    }
  }

  function clearActiveStop() {
    const map = appstate.map
    if (!map || activeFeatureId === null) return
    map.removeFeatureState({ source: 'gtfs-stops', id: activeFeatureId }, 'active')
    activeFeatureId = null
  }

  onMount(async () => {
    const gtfs = appstate.selectedGtfs
    const tripId = appstate.viewTrip
    if (!gtfs || !tripId) return

    const res = await fetch(`/gtfs/trip/${encodeURIComponent(tripId)}?gtfs=${encodeURIComponent(gtfs)}`)
    detail = await res.json()
    loading = false

    const map = appstate.map
    if (!map) return

    if (map.getLayer('gtfs-trips-layer')) {
      map.setFilter('gtfs-trips-layer', ['==', ['get', 'trip_id'], tripId])
    }

    const stopsRes = await fetch(`/gtfs/files/stops?gtfs=${encodeURIComponent(gtfs)}&trip=${encodeURIComponent(tripId)}`)
    const stopsData = await stopsRes.json()
    const stopIds = (stopsData.data ?? []).map((s: { StopID: string }) => s.StopID)
    if (map.getLayer('gtfs-stops-layer')) {
      map.setFilter('gtfs-stops-layer', ['in', ['get', 'stop_id'], ['literal', stopIds]])
    }
  })

  onDestroy(() => {
    const map = appstate.map
    if (!map) return
    if (map.getLayer('gtfs-trips-layer')) map.setFilter('gtfs-trips-layer', null)
    if (map.getLayer('gtfs-stops-layer')) map.setFilter('gtfs-stops-layer', null)
  })

  function back() {
    appstate.viewTrip = undefined
  }

  const color = $derived(
    detail?.trip?.RouteID?.RouteColor
      ? `#${detail.trip.RouteID.RouteColor}`
      : '#94a3b8'
  )
</script>

<div class="col-span-1 flex flex-col overflow-hidden border-r border-gray-200 bg-gray-50">

  <!-- Header -->
  <div class="flex items-center gap-2 p-4 border-b border-gray-200 shrink-0">
    <button
      onclick={back}
      class="text-xs text-blue-600 hover:text-blue-800 font-semibold"
    >← Back</button>
    <span class="text-xs text-gray-400">|</span>
    <span class="text-xs font-semibold uppercase tracking-wide text-gray-500 truncate">Trip Detail</span>
  </div>

  {#if loading}
    <div class="flex items-center justify-center flex-1 text-sm text-gray-400">Loading…</div>
  {:else if detail}
    <!-- Trip summary -->
    <div class="p-4 border-b border-gray-200 shrink-0" style="border-left: 4px solid {color};">
      <div class="font-semibold text-sm">
        {detail.trip.RouteID?.RouteShortName || detail.trip.RouteID?.RouteID}
      </div>
      {#if detail.trip.RouteID?.RouteLongName}
        <div class="text-xs text-gray-500">{detail.trip.RouteID.RouteLongName}</div>
      {/if}
      {#if detail.trip.TripHeadsign}
        <div class="text-xs mt-1">→ {detail.trip.TripHeadsign}</div>
      {/if}
      <div class="text-xs text-gray-400 mt-1">{detail.trip.TripID}</div>

      {#if detail.frequencies?.length > 0}
        <div class="mt-2 text-xs font-semibold text-gray-500 uppercase tracking-wide">Frequencies</div>
        {#each detail.frequencies as f}
          <div class="text-xs text-gray-600">
            {f.StartTime} – {f.EndTime} · every {Math.round(f.HeadwaySecs / 60)} min
          </div>
        {/each}
      {/if}
    </div>

    <!-- Stop times -->
    <div class="overflow-auto flex-1 p-2 text-xs">
      <div class="font-semibold uppercase tracking-wide text-gray-500 mb-2 px-2">
        Stop Times ({detail.stop_times?.length ?? 0})
      </div>
      {#each detail.stop_times ?? [] as st}
        <div
          role="listitem"
          class="flex items-start gap-2 px-2 py-1 rounded hover:bg-gray-100"
          onmouseenter={() => setActiveStop(st.StopID?.StopID)}
          onmouseleave={clearActiveStop}
        >
          <div class="text-gray-400 w-5 shrink-0 text-right">{st.StopSequence}</div>
          <div class="flex-1 min-w-0">
            <div class="truncate text-gray-800">{st.StopID?.StopName || st.StopID?.StopID}</div>
            <div class="text-gray-400">{st.ArrivalTime}</div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
