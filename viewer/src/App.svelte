<script lang="ts">
  import { appstate } from './Components/app.svelte';
  import Map from './Components/Map.svelte'
  import RecordViewer from './Components/RecordViewer.svelte';
  import Sidebar from './Components/Sidebar.svelte'
  import Viewer from './Components/Viewer.svelte'
</script>

<main class="grid grid-cols-4 w-screen h-screen overflow-hidden">
  {#if appstate.activeRecord}
    <RecordViewer/>
  {:else}
  <Sidebar />
  {/if}
  <div class="col-span-3 grid grid-rows-4 overflow-auto">
    <div class="row-span-3">
      <Map />
    </div>
    <Viewer />
  </div>
</main>
