<script lang="ts">
    import { inGame, waitingGameIDs, webSocket } from "./stores.svelte";

    const fetchGames = async () => {
        const response = await fetch("http://localhost:8080/getWaitingGame");

        const data = await response.json();
        $waitingGameIDs = data.ids;
        console.log(data);
    };

    const joinGame = async (gameID: string) => {
        $webSocket = new WebSocket(`ws://localhost:8080/ws?gameID=${gameID}`);
        $inGame = true;
    };

    const createGame = async () => {
        $webSocket = new WebSocket("ws://localhost:8080/ws");
        $inGame = true;
    };
</script>

<!-- When clicked will call fetch to gameBrowser endpoint -->
<div class="buttons-container">
    <button on:click={createGame}>Create Game</button>

    <div class="join-games-container">
        <button on:click={fetchGames}>Fetch Games</button>
        {#each $waitingGameIDs as gameID}
            <button on:click={() => joinGame(gameID)}>Join Game {gameID}</button
            >
        {/each}
    </div>
</div>

<style>
    .buttons-container {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        display: flex;
        flex-direction: row;
        align-items: center;
    }
    .join-games-container {
        display: flex;
        flex-direction: column;
    }
    button {
        background-color: #4caf50;
        border: none;
        color: white;
        padding: 15px 32px;
        text-align: center;
        text-decoration: none;
        display: inline-block;
        font-size: 16px;
        margin: 4px 2px;
        cursor: pointer;
    }

    button:hover {
        background-color: #45a049;
    }
</style>
