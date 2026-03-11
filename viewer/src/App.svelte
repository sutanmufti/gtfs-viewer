<script lang="ts">
  import { onMount } from 'svelte'
  import { appstate, selectGtfs } from './Components/app.svelte';
  import Map from './Components/Map.svelte'
  import RecordViewer from './Components/RecordViewer.svelte';
  import Sidebar from './Components/Sidebar.svelte'
  import TripDetailView from './Components/TripDetailView.svelte';
  import TripsInStop from './Components/TripsInStop.svelte';
  import Viewer from './Components/Viewer.svelte'

  onMount(async () => {
    const res = await fetch('/gtfs/config')
    const config = await res.json()
    if (config.withFile) {
      appstate.withFile = true
      selectGtfs(config.fileName)
    }
  })
</script>

<main class="grid grid-cols-4 w-screen h-screen overflow-hidden">

  {#if appstate.viewTrip}
  <TripDetailView/>
  
  {:else if appstate.stopShowRoute}
      <TripsInStop/>
    {:else}
  {#if appstate.activeRecord}
  <RecordViewer/>
  {:else}
  <Sidebar />
  {/if}
  {/if}
  <div class="col-span-3 grid grid-rows-4 overflow-auto">
    <div class="row-span-3">
      <Map />
    </div>
    <Viewer />
  </div>
</main>
