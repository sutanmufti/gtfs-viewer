<script lang="ts">
  import { appstate, clearStopView } from './app.svelte'
  import { Eye } from '@lucide/svelte'

  type Trip = { TripID: string; TripHeadsign: string; DirectionID: number }
  type Route = { RouteID: string; RouteShortName: string; RouteLongName: string; RouteColor: string }
  type RouteResult = { route: Route; trips: Trip[] }

  const results = $derived((appstate.routeResults ?? []) as RouteResult[])

  function viewTrip(tripId: string) {
    appstate.viewTrip = tripId
  }

  function routeColor(r: Route) {
    return r.RouteColor ? `#${r.RouteColor}` : '#94a3b8'
  }

  let activeFeatureId = $state<number | null>(null)

  function setActiveTrip(tripId: string) {
    const map = appstate.map
    if (!map) return
    if (activeFeatureId !== null) {
      map.removeFeatureState({ source: 'gtfs-trips', id: activeFeatureId }, 'active')
    }
    const features = map.querySourceFeatures('gtfs-trips', {
      filter: ['==', ['get', 'trip_id'], tripId],
    })
    if (features.length > 0 && features[0].id !== undefined) {
      activeFeatureId = features[0].id as number
      map.setFeatureState({ source: 'gtfs-trips', id: activeFeatureId }, { active: true })
    }
  }

  function clearActiveTrip() {
    const map = appstate.map
    if (!map || activeFeatureId === null) return
    map.removeFeatureState({ source: 'gtfs-trips', id: activeFeatureId }, 'active')
    activeFeatureId = null
  }
</script>

<div class="col-span-1 flex flex-col overflow-hidden border-r border-gray-200 bg-gray-50">

  <!-- Header -->
  <div class="flex items-center gap-2 p-4 border-b border-gray-200 shrink-0">
    <button
      onclick={clearStopView}
      class="text-xs text-blue-600 hover:text-blue-800 font-semibold"
    >← Back</button>
    <span class="text-xs text-gray-400">|</span>
    <span class="text-xs font-semibold uppercase tracking-wide text-gray-500">Routes at Stop</span>
  </div>

  <!-- Route list -->
  <div class="overflow-auto flex-1">
    {#each results as result}
      {@const color = routeColor(result.route)}
      <div class="border-b border-gray-200">

        <!-- Route header -->
        <div class="flex items-center gap-2 px-4 py-2" style="border-left: 4px solid {color};">
          <span class="font-semibold text-sm">{result.route.RouteShortName || result.route.RouteID}</span>
          {#if result.route.RouteLongName}
            <span class="text-xs text-gray-500 truncate">{result.route.RouteLongName}</span>
          {/if}
        </div>

        <!-- Trip list -->
        {#each result.trips as trip}
          <div
            role="listitem"
            class="flex items-center gap-2 px-4 py-1.5 hover:bg-gray-100 text-xs"
            onmouseenter={() => setActiveTrip(trip.TripID)}
            onmouseleave={clearActiveTrip}
          >
            <div class="flex-1 min-w-0">
              <span class="text-gray-700 truncate block">
                {trip.TripHeadsign || trip.TripID}
              </span>
              <span class="text-gray-400">{trip.TripID}</span>
            </div>
            <button
              onclick={() => viewTrip(trip.TripID)}
              class="shrink-0 text-blue-500 hover:text-blue-700"
              title="View trip"
            >
              <Eye size={14} />
            </button>
          </div>
        {/each}

      </div>
    {/each}
  </div>
</div>
