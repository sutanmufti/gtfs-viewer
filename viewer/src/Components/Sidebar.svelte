<script lang="ts">
  import { onMount } from 'svelte'
  import {
    appstate,
    GTFS_FILES,
    fetchGtfsList,
    loadViewerPage,
    loadStopsOnMap,
    loadTripStoptimesOnMap,
    selectGtfs,
  } from './app.svelte'

  onMount(fetchGtfsList)

  async function selectFile(file: string) {
    await loadViewerPage(file, 1)
    if (file === 'stops') loadStopsOnMap()
    if (file === 'trips') loadTripStoptimesOnMap()
  }

  let uploading = $state(false)
  let uploadError = $state('')

  async function handleUpload(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return

    uploading = true
    uploadError = ''
    try {
      const form = new FormData()
      form.append('file', file)
      const res = await fetch('/gtfs/upload', { method: 'POST', body: form })
      if (!res.ok) {
        const data = await res.json()
        uploadError = data.error ?? 'Upload failed'
      } else {
        await fetchGtfsList()
      }
    } catch {
      uploadError = 'Network error'
    } finally {
      uploading = false
      input.value = ''
    }
  }
</script>

<div class="col-span-1 flex flex-col overflow-hidden border-r border-gray-200 bg-gray-50">

    <!-- Section 0: Upload GTFS Inputs -->
     {#if !appstate.withFile}
      <div class="p-4 border-b border-gray-200 {appstate.withFile ? 'opacity-40 pointer-events-none' : ''}">
        <h2 class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">Upload GTFS</h2>
        <label class="flex items-center justify-center w-full px-3 py-2 rounded border border-dashed border-gray-300 text-sm text-gray-500 cursor-pointer hover:bg-gray-100 {uploading ? 'opacity-50 pointer-events-none' : ''}">
          {uploading ? 'Uploading…' : 'Choose zip file'}
          <input type="file" accept=".zip" class="hidden" onchange={handleUpload} disabled={uploading || appstate.withFile} />
        </label>
        {#if uploadError}
          <p class="mt-1 text-xs text-red-500">{uploadError}</p>
        {/if}
      </div>
    {/if}

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
