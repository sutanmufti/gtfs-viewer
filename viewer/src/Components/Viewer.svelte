<script lang="ts">
  import { appstate, loadViewerPage, switchLayer } from './app.svelte'

  let showStops = $state(true)
  let showTrips = $state(true)

  function toggleStops() {
    showStops = !showStops
    if (appstate.map) switchLayer(appstate.map, 'gtfs-stops-layer', !showStops)
  }

  function toggleTrips() {
    showTrips = !showTrips
    if (appstate.map) switchLayer(appstate.map, 'gtfs-trips-layer', !showTrips)
  }

  // Derive column headers from the first row of data.
  let columns = $derived(
    appstate.viewerData.length > 0 ? Object.keys(appstate.viewerData[0]) : []
  )

  function prev() {
    if (appstate.viewerPage > 1) loadViewerPage(appstate.selectedFile, appstate.viewerPage - 1)
  }

  function next() {
    if (appstate.viewerPage < appstate.viewerTotalPages) loadViewerPage(appstate.selectedFile, appstate.viewerPage + 1)
  }

  const PAGE_SIZE = 10

  // Derives visible rows for the current page from viewerData.
  let pagedViewerData = $derived(
    appstate.viewerData.filter((_, i) => i < appstate.viewerPage * PAGE_SIZE)
  )
</script>

<div class="row-span-1 h-full flex flex-col overflow-hidden border-t border-gray-200 bg-white">
  {#if !appstate.selectedFile}
    <div class="flex items-center justify-center h-full text-sm text-gray-400 italic">
      Select a file from the sidebar to view its contents.
    </div>
  {:else if appstate.viewerLoading}
    <div class="flex items-center justify-center h-full text-sm text-gray-400">Loading…</div>
  {:else if appstate.viewerData.length === 0}
    <div class="flex items-center justify-center h-full text-sm text-gray-400 italic">No data.</div>
  {:else}

    <!-- toolbar -->
    <div class="flex items-center gap-4 px-4 py-2 border-b border-gray-200 bg-gray-50 text-xs text-gray-600 shrink-0">
      <span class="font-semibold text-gray-500 uppercase tracking-wide">Layers</span>
      <label class="flex items-center gap-1.5 cursor-pointer select-none">
        <input type="checkbox" checked={showStops} onchange={toggleStops} class="accent-blue-600" />
        Stops
      </label>
      <label class="flex items-center gap-1.5 cursor-pointer select-none">
        <input type="checkbox" checked={showTrips} onchange={toggleTrips} class="accent-blue-600" />
        Trips
      </label>
    </div>


    <!-- Table -->
    
    <div class="overflow-auto flex-1 text-xs min-h-0 relative">
      <table class="min-w-full border-collapse overflow-auto">
        <thead class="bg-gray-100 sticky top-0">
          <tr>
            {#each columns as col}
              <th class="px-3 py-2 text-left font-semibold text-gray-600 border-b border-gray-200 whitespace-nowrap">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>

          {#each pagedViewerData as row, i}
            <tr
              class="{i % 2 === 0 ? 'bg-white' : 'bg-gray-50'} hover:bg-blue-50 cursor-pointer"
              onclick={() => appstate.activeRecord = row}
            >
              {#each columns as col}
                <td class="px-3 py-1.5 border-b border-gray-100 whitespace-nowrap text-gray-700 max-w-50 truncate">
                  {row[col] ?? ''}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-between px-4 py-2 border-t border-gray-200 bg-gray-50 text-xs text-gray-600 shrink-0">
      <span>{appstate.selectedFile} — {appstate.viewerTotal} records</span>
      <div class="flex items-center gap-2">
        <button
          onclick={prev}
          disabled={appstate.viewerPage <= 1}
          class="px-2 py-1 rounded border border-gray-300 disabled:opacity-40 hover:bg-gray-200"
        >← Prev</button>
        <span>Page {appstate.viewerPage} / {appstate.viewerTotalPages}</span>
        <button
          onclick={next}
          disabled={appstate.viewerPage >= appstate.viewerTotalPages}
          class="px-2 py-1 rounded border border-gray-300 disabled:opacity-40 hover:bg-gray-200"
        >Next →</button>
      </div>
    </div>
  {/if}
</div>
