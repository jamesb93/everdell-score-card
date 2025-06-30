<script lang="ts">
    import PlayerSummary from "./PlayerSummary.svelte";
    import { deriveTotalScore, type ScoreForPlayer } from "./scores.svelte";
    let { gameData } = $props()
    const getWinner = () => {
        if (!gameData) return null
        const scoresSorted = gameData.scores.sort((a: ScoreForPlayer, b: ScoreForPlayer) => {
            const aTotalScore = deriveTotalScore(a)
            const bTotalScore = deriveTotalScore(b)
            return bTotalScore - aTotalScore
        })
        return scoresSorted[0].player_name
    }
    const winner = getWinner()
</script>

<div class="container">
    {#if gameData}
    <div class="date">{new Date(gameData.game_date).toLocaleDateString()}</div>
    <div class="scores">
        {#each gameData.scores as score}
        <PlayerSummary score={score} winner={winner}/>
        {/each}
    </div>
    {/if}
</div>

<style>
    .container {
        display: flex;
        flex-direction: column;
        gap: 1em;
        outline: 1px solid grey;
        padding: 0.5em;
    }
</style>