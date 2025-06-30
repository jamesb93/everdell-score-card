<script lang="ts">
    import { deriveTotalScore } from "./scores.svelte";
    let { score, winner } = $props();
    let totalScore: number = $derived.by(() => {
        if (!score) {
            return 0
        }
        return deriveTotalScore(score)
    })

    const avatarMap = {
        "Joseph" : "/rugwort.jpg",
        "Niamh" : "/mayor.jpg",
        "James" : "/turtle.jpeg"
    }
</script>

<div class="container">
    {#if score}
    <div class:winner={winner === score.player_name}>{score.player_name}: { totalScore }</div>
    {/if}
    {#if score.player_name in avatarMap}
    <img src={avatarMap[score.player_name]}/>
    {/if}
</div>

<style>
    .container {
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 0.5em;
    }
    .winner {
        color: var(--bronze);
    }

    img {
        max-width: 40px;
    }
</style>