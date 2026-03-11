<script lang="ts">
  import mapboxgl from 'mapbox-gl'
  import { appstate } from './app.svelte'
  import { onMount } from 'svelte'
  import 'mapbox-gl/dist/mapbox-gl.css'

  const envApiKey = import.meta.env.VITE_MAPBOX_KEY

  onMount(() => {
    const queryString = window.location.search
    const urlParams = new URLSearchParams(queryString)

    mapboxgl.accessToken = envApiKey ?? urlParams.get('api_key') ?? ''

    appstate.map = new mapboxgl.Map({
      container: appstate.mapdiv!,
      style: 'mapbox://styles/mapbox/light-v11',
      center: [0, 20],
      zoom: 2,
    })

    appstate.map.on('load', () => {
      // Pointer cursor on stop hover.
      appstate.map!.on('mouseenter', 'gtfs-stops-layer', () => {
        appstate.map!.getCanvas().style.cursor = 'pointer'
      })
      appstate.map!.on('mouseleave', 'gtfs-stops-layer', () => {
        appstate.map!.getCanvas().style.cursor = ''
      })

      // Popup on stop click.
      appstate.map!.on('click', 'gtfs-stops-layer', (e) => {
        const feature = e.features?.[0]
        if (!feature) return
        const props = feature.properties as Record<string, string>
        const coords = (feature.geometry as GeoJSON.Point).coordinates as [number, number]
        new mapboxgl.Popup()
          .setLngLat(coords)
          .setHTML(
            `<strong>${props.stop_name || props.stop_id}</strong>` +
            (props.stop_code ? `<br/>Code: ${props.stop_code}` : '') +
            (props.stop_desc ? `<br/>${props.stop_desc}` : '')
          )
          .addTo(appstate.map!)
      })

      // Pointer cursor on trip line hover.
      appstate.map!.on('mouseenter', 'gtfs-trips-layer', () => {
        appstate.map!.getCanvas().style.cursor = 'pointer'
      })
      appstate.map!.on('mouseleave', 'gtfs-trips-layer', () => {
        appstate.map!.getCanvas().style.cursor = ''
      })

      // Popup on trip line click.
      appstate.map!.on('click', 'gtfs-trips-layer', (e) => {
        const feature = e.features?.[0]
        if (!feature) return
        const props = feature.properties as Record<string, string>
        const coords = e.lngLat
        new mapboxgl.Popup()
          .setLngLat(coords)
          .setHTML(
            `<strong>${props.route_short_name || props.route_id}</strong>` +
            (props.route_long_name ? `<br/>${props.route_long_name}` : '') +
            (props.headsign ? `<br/>→ ${props.headsign}` : '') +
            `<br/><span style="font-size:0.75em;color:#666">${props.trip_id}</span>`
          )
          .addTo(appstate.map!)
      })
    })
  })
</script>

<div class="grow w-full h-full" bind:this={appstate.mapdiv}></div>
