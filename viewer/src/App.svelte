<script lang="ts">
  import Map from './Components/Map.svelte'
  import Sidebar from './Components/Sidebar.svelte'
  import Viewer from './Components/Viewer.svelte'
</script>

<main class="grid grid-cols-4 w-screen h-screen overflow-hidden">
  <Sidebar />
  <div class="col-span-3 grid grid-rows-4">
    <div class="row-span-3">
      <Map />
    </div>
    <Viewer />
  </div>
</main>
