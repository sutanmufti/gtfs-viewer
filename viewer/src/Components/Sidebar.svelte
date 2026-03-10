<script lang="ts">
  import { onMount } from 'svelte'
  import {
    appstate,
    GTFS_FILES,
    fetchGtfsList,
    loadViewerPage,
    loadStopsOnMap,
  } from './app.svelte'

  onMount(fetchGtfsList)

  async function selectGtfs(name: string) {
    appstate.selectedGtfs = name
    appstate.selectedFile = ''
    appstate.viewerData = []
    appstate.viewerTotal = 0

    // Always load stops onto the map when a feed is selected.
    // Wait for the map to be ready if it isn't yet.
    if (appstate.map) {
      loadStopsOnMap()
    } else {
      const interval = setInterval(() => {
        if (appstate.map) {
          clearInterval(interval)
          loadStopsOnMap()
        }
      }, 200)
    }
  }

  async function selectFile(file: string) {
    await loadViewerPage(file, 1)
    if (file === 'stops') loadStopsOnMap()
  }
</script>

<div class="col-span-1 flex flex-col overflow-hidden border-r border-gray-200 bg-gray-50">

  <!-- Section 1: GTFS zip selector -->
  <div class="p-4 border-b border-gray-200">
    <h2 class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">GTFS Feeds</h2>
    {#if appstate.gtfsZipFiles.length === 0}
      <p class="text-sm text-gray-400 italic">No feeds uploaded.</p>
    {:else}
      <ul class="space-y-1">
        {#each appstate.gtfsZipFiles as name}
          <li>
            <button
              onclick={() => selectGtfs(name)}
              class="w-full text-left px-3 py-2 rounded text-sm truncate
                     {appstate.selectedGtfs === name
                       ? 'bg-blue-600 text-white font-semibold'
                       : 'hover:bg-gray-200 text-gray-700'}"
            >
              {name}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <!-- Section 2: GTFS file buttons -->
  <div class="p-4 overflow-auto flex-1">
    <h2 class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">Files</h2>
    {#if !appstate.selectedGtfs}
      <p class="text-sm text-gray-400 italic">Select a feed above.</p>
    {:else}
      <ul class="space-y-1">
        {#each GTFS_FILES as file}
          <li>
            <button
              onclick={() => selectFile(file)}
              class="w-full text-left px-3 py-2 rounded text-sm
                     {appstate.selectedFile === file
                       ? 'bg-blue-100 text-blue-700 font-semibold'
                       : 'hover:bg-gray-200 text-gray-700'}"
            >
              {file}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
