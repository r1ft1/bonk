<script lang="ts">
    import {
        inGame,
        gameState,
        waitingGameIDs,
        webSocket,
    } from "./stores.svelte";
    import type { ServerMessage } from "./stores.svelte";

    const fetchGames = async () => {
        const response = await fetch("http://localhost:8080/getWaitingGame");

        const data = await response.json();
        $waitingGameIDs = data.ids;
        console.log(data);
    };

    const joinGame = async (gameID: string) => {
        $webSocket = new WebSocket(`ws://localhost:8080/ws?gameID=${gameID}`);
        $webSocket.addEventListener("message", function (event) {
            const msg: ServerMessage = JSON.parse(event.data);
            if (msg.type == "error" && msg.payload == "Could not join game") {
                console.log("Could not join game");
                $webSocket.close();
                return;
            }
            if (msg.type == "joined") {
                $inGame = true;
            }
            if (msg.type != "error") {
                $gameState = msg.payload;
            } else {
                console.log(msg.payload);
            }
            console.log($gameState);
        });
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
    @import url("https://fonts.googleapis.com/css2?family=Cherry+Bomb+One&display=swap");

    * {
        color: white;
        font-family: "Cherry Bomb One", serif;
        font-weight: 400;
        font-style: normal;
    }
    div {
        border-style: dashed;
        border-radius: 25px;
        background-color: rgba(98, 163, 169, 0.5);
        margin: 1rem;
    }

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
        border-style: dashed;
        border-radius: 25px;
        background-color: rgba(98, 163, 169, 0.5);
        margin: 1rem;
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
