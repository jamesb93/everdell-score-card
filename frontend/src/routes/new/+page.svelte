<script lang="ts">
    import { type GameData, createPlayer, type ScoreForPlayer, deriveTotalScore } from "$lib/scores.svelte"

    const createNewGameData = () => {
        return {
            game_date: '',
            scores: [createPlayer(), createPlayer()]
        }
    }
    let game: GameData = $state(createNewGameData())

    function addPlayer(event: MouseEvent) {
        event.preventDefault()
        game.scores.push(createPlayer())
    }

    async function handleSubmit() {
        const payload = JSON.stringify(game)
        const response = await fetch('http://localhost:8080/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: payload
        })
    }

    function removePlayer(event: MouseEvent, playerId: string) {
        event.preventDefault()
        console.log(playerId, game.scores)
        if (game.scores.length <= 2) {
            return
        }
        game.scores = game.scores.filter(player => player.id !== playerId)
    }
</script>


<form onsubmit={handleSubmit}>
    <div class="top-row">
        <div>
            <label for="game_date" >Game Date</label>
            <input type="date" id="gameDate" bind:value={game.game_date}>
        </div>
        <button onclick={addPlayer}>Add Player</button>
    </div>

    <div class="player-container">
        {#each game.scores as player}
        <div class="player">
            <button class="remove-player" onclick={(event) => removePlayer(event, player.id)}>X</button>
            <div class="property">
                <label for="player_name-{player.id}">Name</label>
                <input type="text" id="player_name-{player.id}" bind:value={player.player_name}>
            </div>
            <div class="property">
                <label for="base_cards-{player.id}">Base Cards</label>
                <input type="number" id="base_cards-{player.id}" bind:value={player.base_cards}>
            </div>
            <div class="property">
                <label for="extra_vp-{player.id}">Extra VP</label>
                <input type="number" id="extra_vp-{player.id}" bind:value={player.extra_vp}>
            </div>
            <div class="property">
                <label for="basic_events-{player.id}">Basic Events</label>
                <input type="number" id="basic_events-{player.id}" bind:value={player.basic_events}>
            </div>
            <div class="property">
                <label for="special_events-{player.id}">Special Events</label>
                <input type="number" id="special_events-{player.id}" bind:value={player.special_events}>
            </div>
            <div class="property">
                <label for="prosperity_cards-{player.id}">Prosperity Cards</label>
                <input type="number" id="prosperity_cards-{player.id}" bind:value={player.prosperity_cards}>
            </div>
            <div class="property">
                <label for="visitors-{player.id}">Visitors</label>
                <input type="number" id="visitors-{player.id}" bind:value={player.visitors}>
            </div>
            <div class="property">
                <label for="journey-{player.id}">Journey</label>
                <input type="number" id="journey-{player.id}" bind:value={player.journey}>
            </div>
            <div class="property">
                <label for="garland_award-{player.id}">Garland Award</label>
                <input type="number" id="garland_award-{player.id}" bind:value={player.garland_award}>
            </div>
            <div class="property">Total Score: {deriveTotalScore(player)}</div>
        </div>
        
        {/each}
    </div>

    <input type="submit" value="Log!" />
</form>

<style>
    form {
        display: flex;
        flex-direction: column;
        gap: 0.5em;
    }

    .top-row {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        width: 100%;
        padding-bottom: 1em;
    }
    .remove-player {
        display: grid;
        place-items: center;
        max-width: 2em;
    }
    .property {
        width: 100%;
        display: flex;
        justify-content: space-between;
    }

    .property input {
        max-width: 40%;
    }
    .player-container {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        gap: 0.5em;
        width: 100%;
        flex-grow: 1;
        max-width: 400px;
    }

    .player {
        display:flex;
        flex-direction: column;
        gap: 0.5em;
        padding: 0.5em;
        border: 1px solid grey;
        flex-grow: 1;
    }
</style>