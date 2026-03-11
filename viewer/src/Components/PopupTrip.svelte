<script lang="ts">

    let {stop_name, stop_id, stop_code, stop_desc}: {stop_name: string, stop_id: string, stop_code?: string, stop_desc?: string } = $props()

</script>


<strong>{stop_name || stop_id}</strong>

{#if stop_code}
<br/>Code: {stop_code}
{/if}

{#if stop_desc}
<br/>{stop_desc}
{/if}