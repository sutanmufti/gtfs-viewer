<script lang="ts">
  import { appstate } from './app.svelte'

  function back() {
    appstate.activeRecord = undefined
    if (appstate.map?.getLayer('gtfs-trips-layer')) {
      appstate.map.setFilter('gtfs-trips-layer', null)
    }
    if (appstate.map?.getLayer('gtfs-stops-layer')) {
      appstate.map.setFilter('gtfs-stops-layer', null)
    }
  }

  // Format a value for display: primitives inline, objects as indented JSON.
  function fmt(value: unknown): string {
    if (value === null || value === undefined) return '—'
    if (typeof value === 'object') return JSON.stringify(value, null, 2)
    return String(value)
  }

  function isObject(value: unknown): boolean {
    return typeof value === 'object' && value !== null
  }
</script>

<div class="col-span-1 flex flex-col overflow-hidden border-r border-gray-200 bg-gray-50">

  <!-- Header with back button -->
  <div class="flex items-center gap-2 p-4 border-b border-gray-200 shrink-0">
    <button
      onclick={back}
      class="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 font-semibold"
    >
      ← Back
    </button>
    <span class="text-xs text-gray-400">|</span>
    <span class="text-xs font-semibold uppercase tracking-wide text-gray-500 truncate">
      {appstate.selectedFile} detail
    </span>
  </div>

  <!-- Record fields -->
  <div class="overflow-auto flex-1 p-4 space-y-3 text-xs">
    {#each Object.entries(appstate.activeRecord ?? {}) as [key, value]}
      <div>
        <div class="font-semibold text-gray-500 uppercase tracking-wide mb-0.5">{key}</div>
        {#if isObject(value)}
          <pre class="bg-gray-100 rounded p-2 text-gray-700 overflow-auto whitespace-pre-wrap break-all">{fmt(value)}</pre>
        {:else}
          <div class="text-gray-800 break-all">{fmt(value)}</div>
        {/if}
      </div>
    {/each}
  </div>
</div>
