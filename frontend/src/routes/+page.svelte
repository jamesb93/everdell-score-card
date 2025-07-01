<script lang="ts">
    import { onMount } from 'svelte';
    import { type GameData } from '$lib/scores.svelte'
    import GameSummary from '$lib/GameSummary.svelte';
    import { API_URL } from "$lib/api"
    
    let gameDataLog: GameData[] = $state([])
    let loading = $state(true)
        
    onMount(async () => {
        const response = await fetch(`${API_URL}/games`)
        const data = await response.json()
        gameDataLog = data ?? []
        loading = false
    })
</script>

<a href="/new">Add New Game</a>

{#if loading}
<div>Loading...</div>
{:else}

{#if gameDataLog.length >= 1 }
<div class="games">
    {#each gameDataLog as gameData}
    <GameSummary gameData={gameData}/>
    {/each}
</div>
{:else}
<div>no games recorded!</div>
{/if}
{/if}

<style>
    .games {
        display: flex;
        flex-direction: column;
        gap: 1em;
    }
</style>
