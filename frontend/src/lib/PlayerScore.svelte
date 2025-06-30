<script lang="ts">
    import type { ScoreForPlayer } from "./scores.svelte";
    let { score } = $props<{score: ScoreForPlayer | null}>()
    
    let totalScore: number = $derived.by(() => {
        if (!score) return  0
        return (
            score.base_cards + 
            score.extra_vp + 
            score.basic_events + 
            score.special_events + 
            score.prosperity_cards + 
            score.visitors + 
            score.journey + 
            score.garland_award
        )
    })
</script>
<div class="container">
    {#if score}
    <div>{score.player_name}</div>
    <div>Total Score: {totalScore}</div>
    <details>
        <summary>Score Breakdown</summary>
        <div>Base Cards: {score.base_cards}</div>
        <div>Extra VP: {score.extra_vp}</div>
        <div>Basic Events: {score.basic_events}</div>
        <div>Special Events: {score.special_events}</div>
        <div>Prosperity Cards: {score.prosperity_cards}</div>
        <div>Visitors: {score.visitors}</div>
        <div>Journey: {score.journey}</div>
        <div>Garland Award: {score.garland_award}</div>
    </details>
    {/if}
</div>

<style>
    .container {
        padding: 0.5em;
        display: flex;
        flex-direction: column;
        gap: 0.5em;
    }
</style>