<script lang="ts">
  import { appstate } from "./app.svelte";

  let {
    trip_id,
    headsign,
    route_short_name,
    route_long_name,
    route_color,
  }: {
    trip_id: string
    headsign?: string
    route_short_name?: string
    route_long_name?: string
    route_color?: string
  } = $props()

  const color = $derived(route_color ? `#${route_color}` : '#94a3b8')
</script>

<div style="border-left: 4px solid {color}; padding-left: 6px;">
  <strong>{route_short_name || route_long_name || trip_id}</strong>
  {#if route_long_name && route_short_name}
    <div style="font-size:0.8em; color:#555">{route_long_name}</div>
  {/if}
  {#if headsign}
    <div>→ {headsign}</div>
  {/if}
  <div style="font-size:0.75em; color:#888">{trip_id}</div>

  <div>
    <button onclick={()=>{
      appstate.viewTrip = trip_id
    }}>View Trip</button>
  </div>
</div>
