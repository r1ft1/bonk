<script lang="ts">
    import {
        inGame,
        gameState,
        waitingGameIDs,
        webSocket,
        p1WebSocket,
        p2WebSocket,
    } from "./stores.svelte";
    import type { ServerMessage } from "./stores.svelte";

    const fetchGames = async () => {
        const response = await fetch(
            import.meta.env.VITE_SERVER_HTTP_URL + "/getWaitingGame",
        );

        const data = await response.json();
        $waitingGameIDs = data.ids;
        console.log(data);
    };

    const joinGame = async (gameID: string) => {
        $webSocket = new WebSocket(
            `${import.meta.env.VITE_SERVER_WS_URL}/ws?gameID=${gameID}`,
        );
        $webSocket.addEventListener("message", messageEvent);
    };

    const startLocalPassAndPlay = async () => {
        $p1WebSocket = new WebSocket(
            import.meta.env.VITE_SERVER_WS_URL + "/ws",
        );
        $p1WebSocket.onmessage = (event: MessageEvent<any>) => {
            const msg: ServerMessage = JSON.parse(event.data);
            if (msg.type == "joined") {
                console.log(msg.gameID);
                $p2WebSocket = new WebSocket(
                    `${import.meta.env.VITE_SERVER_WS_URL}/ws?gameID=${msg.gameID}`,
                );
                $p1WebSocket.addEventListener("message", messageEvent);
                $p2WebSocket.addEventListener("message", messageEvent);
                $inGame = true;
            }
        };
    };

    const createGame = async () => {
        $webSocket = new WebSocket(import.meta.env.VITE_SERVER_WS_URL + "/ws");
        $webSocket.addEventListener("message", messageEvent);
    };

    const messageEvent = (event: MessageEvent<any>) => {
        //server will send a ping every 30 seconds, when received, send a pong back
        const msg: ServerMessage = JSON.parse(event.data);
        if (msg.type == "ping") {
            //Piece 99 is a pong
            if ($webSocket != null) {
                $webSocket.send(
                    JSON.stringify({
                        position: { x: 0, y: 0 },
                        piece: 99,
                    }),
                );
            }

            if ($p1WebSocket != null) {
                $p1WebSocket.send(
                    JSON.stringify({
                        position: { x: 0, y: 0 },
                        piece: 99,
                    }),
                );
            }

            if ($p2WebSocket != null) {
                $p2WebSocket.send(
                    JSON.stringify({
                        position: { x: 0, y: 0 },
                        piece: 99,
                    }),
                );
            }

            return;
        }
        if (msg.type != "error") {
            $gameState = msg.payload;
            console.log($gameState);
        }
        if (msg.type == "error" && msg.payload == "Could not join game") {
            console.log("Could not join game");
            if ($webSocket != null) $webSocket.close();
            if ($p1WebSocket != null) $p1WebSocket.close();
            if ($p2WebSocket != null) $p2WebSocket.close();
            return;
        }
        if (msg.type == "joined") {
            $inGame = true;
            console.log(msg.payload);
        } else {
            console.log(msg.payload);
        }
    };
</script>

<!-- When clicked will call fetch to gameBrowser endpoint -->
<div class="buttons-container">
    <button on:click={startLocalPassAndPlay}
        >Start Local Pass and Play Game</button
    >

    <button on:click={createGame}>Create Online Game</button>

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
