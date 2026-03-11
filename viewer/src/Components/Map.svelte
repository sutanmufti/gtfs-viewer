<script lang="ts">
  import mapboxgl from 'mapbox-gl'
  import { appstate } from './app.svelte'
  import { onMount, mount } from 'svelte'
  import 'mapbox-gl/dist/mapbox-gl.css'
  import PopupTrip from './PopupTrip.svelte'
  import PopupStop from './PopupStop.svelte';

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
      appstate.map!.on('click', 'gtfs-trips-layer', (e) => {
        const feature = e.features?.[0]
        console.log("trip:",feature)
        if (!feature) return
        const props = feature.properties as Record<string, string>
        const coords = e.lngLat


        const divPopupTrip = document.createElement('div')
        mount(PopupTrip, {
          target: divPopupTrip,
          props: {
            trip_id:          props.trip_id,
            headsign:         props.headsign,
            route_short_name: props.route_short_name,
            route_long_name:  props.route_long_name,
            route_color:      props.route_color,
          },
        })
        new mapboxgl.Popup()
          .setLngLat(coords)
          .setDOMContent(divPopupTrip)
          .addTo(appstate.map!)
      })

      // Pointer cursor on trip line hover.
      appstate.map!.on('mouseenter', 'gtfs-trips-layer', () => {
        appstate.map!.getCanvas().style.cursor = 'pointer'
      })
      appstate.map!.on('mouseleave', 'gtfs-trips-layer', () => {
        appstate.map!.getCanvas().style.cursor = ''
      })

      appstate.map!.on('click', 'gtfs-stops-layer', (e) => {
        const feature = e.features?.[0]
        if (!feature) return
        const props = feature.properties as Record<string, string>
        const coords = (feature.geometry as GeoJSON.Point).coordinates as [number, number]

        const div = document.createElement('div')
        mount(PopupStop, {
          target: div,
          props: {
            stop_id:       props.stop_id,
            stop_name:     props.stop_name,
            stop_code:     props.stop_code,
            stop_desc:     props.stop_desc,
            ParentStation: props.ParentStation ?? '',
          },
        })
        new mapboxgl.Popup()
          .setLngLat(coords)
          .setDOMContent(div)
          .addTo(appstate.map!)
      })

    })
  })
</script>

<div class="grow w-full h-full" bind:this={appstate.mapdiv}></div>
