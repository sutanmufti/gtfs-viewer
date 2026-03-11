<script lang="ts">
  import { showRoutes } from './app.svelte'

  let {
    stop_id,
    stop_name,
    stop_code,
    stop_desc,
    ParentStation,
  }: {
    stop_id: string
    stop_name: string
    stop_code?: string
    stop_desc?: string
    ParentStation: string
  } = $props()
</script>

<strong>{stop_name || stop_id}</strong>

{#if stop_code}
<br/>→ {stop_code}
{/if}

{#if stop_desc}
<br/>{stop_desc}
{/if}

<br/><span style="font-size:0.75em;color:#666">{ParentStation}</span>

<br/><button onclick={()=>showRoutes(stop_id)}>Show Routes</button>
