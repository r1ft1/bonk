<script lang="ts">
    import {
        inGame,
        gameState,
        waitingGameIDs,
        waitingForOpponent,
        onlineGameID,
        webSocket,
        p1WebSocket,
        p2WebSocket,
        graduatingLines,
        boopedOffPieces,
        slidingPieces,
        pieceChoice,
    } from "./stores";
    import type { ServerMessage } from "./stores";
    import { PUBLIC_SERVER_WS_URL, PUBLIC_SERVER_HTTP_URL } from "$env/static/public";

    let boopOffIdCounter = 0;
    let slidingIdCounter = 0;
    let lastProcessedTurn = -1;

    type ServerBooped = {
        direction: { x: number; y: number };
        position: { x: number; y: number };
        tile: number;
        boopedBy: number;
    };

    const fetchGames = async () => {
        const response = await fetch(
            PUBLIC_SERVER_HTTP_URL + "/getWaitingGame",
        );

        const data = await response.json();
        $waitingGameIDs = data.ids;
        console.log(data);
    };

    const joinGame = async (gameID: string) => {
        $webSocket = new WebSocket(
            `${PUBLIC_SERVER_WS_URL}/ws?gameID=${gameID}`,
        );
        $webSocket.addEventListener("message", messageEvent);
    };

    let statusMessage = "";

    const startLocalPassAndPlay = async () => {
        console.log("PUBLIC_SERVER_WS_URL:", PUBLIC_SERVER_WS_URL);
        if (!PUBLIC_SERVER_WS_URL) {
            statusMessage = "Error: PUBLIC_SERVER_WS_URL is not set";
            return;
        }
        statusMessage = "Connecting...";
        $p1WebSocket = new WebSocket(PUBLIC_SERVER_WS_URL + "/ws");
        $p1WebSocket.onerror = () => {
            statusMessage = `Error: could not connect to ${PUBLIC_SERVER_WS_URL}`;
        };
        $p1WebSocket.onmessage = (event: MessageEvent<any>) => {
            const msg: ServerMessage = JSON.parse(event.data);
            if (msg.type == "joined") {
                console.log(msg.gameID);
                $p2WebSocket = new WebSocket(
                    `${PUBLIC_SERVER_WS_URL}/ws?gameID=${msg.gameID}`,
                );
                $p1WebSocket.addEventListener("message", messageEvent);
                $p2WebSocket.addEventListener("message", messageEvent);
                $pieceChoice = 0;
                $inGame = true;
            }
        };
    };

    const createGame = async () => {
        $webSocket = new WebSocket(PUBLIC_SERVER_WS_URL + "/ws");
        $webSocket.addEventListener("message", messageEvent);
    };


    const messageEvent = (event: MessageEvent<any>) => {
        const msg: ServerMessage = JSON.parse(event.data);
        if (msg.type == "ping") {
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
            const newPayload = msg.payload;
            const newTurn = newPayload.turnNumber ?? -1;

            // Skip animation processing if:
            // - Same turn (duplicate from p1/p2 sockets)
            // - Transitioning from a waiting state (MAX_WAITING/MULTIPLE_WAITING → WAITING)
            //   These are selection responses, not new placements
            const oldState = $gameState;
            const isWaitingTransition = oldState.state === "MAX_WAITING" || oldState.state === "MULTIPLE_WAITING";
            if (newTurn === lastProcessedTurn || isWaitingTransition) {
                lastProcessedTurn = newTurn;
                $gameState = newPayload;
                // Fall through to handle joined/gameState UI logic below
            } else {
            lastProcessedTurn = newTurn;

            // Graduation: use lines from server directly
            const serverLines = newPayload.lines as { x: number; y: number }[][] | null;
            if (serverLines && serverLines.length > 0) {
                // Determine which player graduated from the placed piece
                const placedPiece = newPayload.placed?.piece ?? 0;
                const playerTile = (placedPiece === 1 || placedPiece === 2) ? 1 : 8;
                for (const line of serverLines) {
                    const worldPositions = line.map(
                        (p: { x: number; y: number }) => [p.x - 2.5, 0.52, p.y - 2.5] as [number, number, number]
                    );
                    $graduatingLines = [
                        ...$graduatingLines,
                        { positions: worldPositions, tile: playerTile },
                    ];
                }
            }

            // Sliding pieces: use boopMovement from server
            const boopMoves = newPayload.boopMovement as { position: { x: number; y: number }; finalPosition: { x: number; y: number }; tile: number }[] ?? [];
            const newSliding: typeof $slidingPieces = [];
            for (const bm of boopMoves) {
                newSliding.push({
                    id: slidingIdCounter++,
                    startPos: [bm.position.x - 2.5, 0.52, bm.position.y - 2.5],
                    endPos: [bm.finalPosition.x - 2.5, 0.52, bm.finalPosition.y - 2.5],
                    tile: bm.tile,
                });
            }

            // Booped-off pieces: use booped from server
            const serverBooped = (newPayload.booped as ServerBooped[]) ?? [];
            const newBoopedOff: typeof $boopedOffPieces = [];
            for (const b of serverBooped) {
                newBoopedOff.push({
                    id: boopOffIdCounter++,
                    startPos: [b.position.x - 2.5, 0.52, b.position.y - 2.5],
                    tile: b.tile,
                    direction: [b.direction.x, b.direction.y],
                });
            }

            // Animation sequencing is reactive:
            // 1. Placement arc runs → sets placementLanded when done
            // 2. SlidingPiece waits for placementLanded
            // 3. FlyingPiece waits for placementLanded AND slidingPieces to be empty
            if (newSliding.length > 0) {
                $slidingPieces = [...$slidingPieces, ...newSliding];
            }

            if (newBoopedOff.length > 0) {
                $boopedOffPieces = [...$boopedOffPieces, ...newBoopedOff];
            }

            $gameState = newPayload;
            console.log($gameState);
            } // end duplicate guard
        }
        if (msg.type == "error" && msg.payload == "Could not join game") {
            console.log("Could not join game");
            if ($webSocket != null) $webSocket.close();
            if ($p1WebSocket != null) $p1WebSocket.close();
            if ($p2WebSocket != null) $p2WebSocket.close();
            return;
        }
        if (msg.type == "joined") {
            if (msg.playerID == "player1") {
                // Game creator: wait for opponent to join
                $waitingForOpponent = true;
                $onlineGameID = msg.gameID;
            } else {
                // Joiner: go straight into the game
                $pieceChoice = 0;
                $inGame = true;
            }
            console.log(msg.payload);
        } else if (msg.type == "gameState" && $waitingForOpponent) {
            // Opponent has joined — enter the game
            $waitingForOpponent = false;
            $pieceChoice = 0;
                $inGame = true;
        } else {
            console.log(msg.payload);
        }
    };
</script>

<div class="menu-overlay">
    <div class="menu-card">
        <div class="paw paw-1"></div>
        <div class="paw paw-2"></div>
        <div class="paw paw-3"></div>
        <div class="paw paw-4"></div>
        <div class="paw paw-5"></div>

        <h1 class="title">boop.</h1>
        <p class="subtitle">a game of kittens & cats</p>

        {#if $waitingForOpponent}
            <div class="waiting-section">
                <p class="waiting-text">Waiting for opponent to join...</p>
                <div class="game-code">
                    <span class="game-code-label">Game Code</span>
                    <span class="game-code-value">{$onlineGameID}</span>
                </div>
                <div class="waiting-dots">
                    <span class="dot"></span>
                    <span class="dot"></span>
                    <span class="dot"></span>
                </div>
            </div>
        {:else}
            {#if statusMessage}
                <div class="status">{statusMessage}</div>
            {/if}

            <div class="buttons">
                <button class="btn btn-primary" onclick={startLocalPassAndPlay}>
                    <span class="btn-label">Local Game</span>
                    <span class="btn-desc">Pass & play on this device</span>
                </button>

                <button class="btn btn-secondary" onclick={createGame}>
                    <span class="btn-label">Create Online Game</span>
                    <span class="btn-desc">Host a new game room</span>
                </button>

                <button class="btn btn-tertiary" onclick={fetchGames}>
                    <span class="btn-label">Browse Games</span>
                    <span class="btn-desc">Join an existing game</span>
                </button>
            </div>

            {#if $waitingGameIDs.length > 0}
                <div class="game-list">
                    <p class="game-list-title">Open Games</p>
                    {#each $waitingGameIDs as gameID}
                        <button class="btn btn-join" onclick={() => joinGame(gameID)}>
                            Join Game {gameID}
                        </button>
                    {/each}
                </div>
            {/if}
        {/if}
    </div>
</div>

<style>
    .menu-overlay {
        position: absolute;
        inset: 0;
        z-index: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        pointer-events: none;
    }

    .menu-card {
        pointer-events: all;
        position: relative;
        background: #faf6f0;
        border: 2px solid rgba(180, 160, 140, 0.3);
        border-radius: 24px;
        padding: 2.5rem 2.5rem 2rem;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.25rem;
        box-shadow:
            0 4px 24px rgba(100, 80, 60, 0.1),
            0 1px 3px rgba(100, 80, 60, 0.08);
        max-width: 380px;
        width: 90vw;
        overflow: hidden;
    }

    /* Paw print decorations */
    .paw {
        position: absolute;
        width: 30px;
        height: 30px;
        pointer-events: none;
        opacity: 0.12;
        background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512'%3E%3Cpath fill='%238b7b6b' d='M256 310c-34-40-107-50-107-50s-37 85-3 141c26 42 76 54 110 54s84-12 110-54c34-56-3-141-3-141s-73 10-107 50z'/%3E%3Cellipse fill='%238b7b6b' cx='118' cy='226' rx='52' ry='72' transform='rotate(-20 118 226)'/%3E%3Cellipse fill='%238b7b6b' cx='394' cy='226' rx='52' ry='72' transform='rotate(20 394 226)'/%3E%3Cellipse fill='%238b7b6b' cx='214' cy='115' rx='48' ry='68' transform='rotate(-8 214 115)'/%3E%3Cellipse fill='%238b7b6b' cx='298' cy='115' rx='48' ry='68' transform='rotate(8 298 115)'/%3E%3C/svg%3E");
        background-size: contain;
        background-repeat: no-repeat;
    }

    .paw-1 { top: 12px; left: 18px; transform: rotate(-25deg); }
    .paw-2 { top: 8px; right: 24px; transform: rotate(15deg); }
    .paw-3 { bottom: 20px; left: 14px; transform: rotate(-40deg); }
    .paw-4 { bottom: 12px; right: 20px; transform: rotate(30deg); }
    .paw-5 { top: 50%; right: 10px; transform: rotate(10deg); }

    .title {
        font-family: "Cherry Bomb One", serif;
        font-size: 3.5rem;
        font-weight: 400;
        color: #5a4a3a;
        margin: 0;
        letter-spacing: 0.02em;
        line-height: 1.1;
    }

    .subtitle {
        font-family: "Nunito", sans-serif;
        font-size: 0.85rem;
        font-weight: 600;
        color: #9a8a7a;
        margin: 0.5rem 0 1.5rem 0;
        letter-spacing: 0.12em;
        text-transform: uppercase;
        white-space: nowrap;
    }

    .buttons {
        display: flex;
        flex-direction: column;
        gap: 0.6rem;
        width: 100%;
    }

    .btn {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.15rem;
        width: 100%;
        padding: 0.9rem 1.25rem;
        border: 2px solid transparent;
        border-radius: 16px;
        cursor: pointer;
        transition: transform 0.2s ease, box-shadow 0.2s ease;
        text-align: center;
        font-family: "Nunito", sans-serif;
    }

    .btn-primary {
        background: #d4eef6;
        border-color: rgba(142, 200, 219, 0.4);
    }

    .btn-secondary {
        background: #f6ddd4;
        border-color: rgba(219, 170, 142, 0.35);
    }

    .btn-tertiary {
        background: #e4daf6;
        border-color: rgba(170, 142, 219, 0.3);
    }

    .btn:hover {
        transform: scale(1.03);
        box-shadow: 0 3px 12px rgba(100, 80, 60, 0.12);
    }

    .btn:active {
        transform: scale(0.99);
    }

    .btn-label {
        font-size: 1rem;
        font-weight: 700;
        color: #4a5568;
    }

    .btn-desc {
        font-size: 0.75rem;
        font-weight: 400;
        color: #8a8a8a;
    }

    .btn-join {
        background: #d4f0e0;
        border-color: rgba(130, 200, 160, 0.4);
        padding: 0.7rem 1rem;
        color: #4a5568;
        font-weight: 600;
        font-size: 0.9rem;
    }

    .game-list {
        width: 100%;
        margin-top: 0.75rem;
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }

    .game-list-title {
        font-family: "Nunito", sans-serif;
        font-size: 0.75rem;
        font-weight: 700;
        color: #9a8a7a;
        text-transform: uppercase;
        letter-spacing: 0.12em;
        margin: 0 0 0.25rem 0.25rem;
    }

    .waiting-section {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        width: 100%;
        padding: 0.5rem 0;
    }

    .waiting-text {
        font-family: "Nunito", sans-serif;
        font-size: 1rem;
        font-weight: 600;
        color: #5a4a3a;
        margin: 0;
    }

    .game-code {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.3rem;
        background: #f0ebe4;
        border: 2px solid rgba(180, 160, 140, 0.3);
        border-radius: 16px;
        padding: 1rem 2rem;
    }

    .game-code-label {
        font-family: "Nunito", sans-serif;
        font-size: 0.7rem;
        font-weight: 700;
        color: #9a8a7a;
        text-transform: uppercase;
        letter-spacing: 0.12em;
    }

    .game-code-value {
        font-family: "Cherry Bomb One", serif;
        font-size: 2rem;
        color: #5a4a3a;
        letter-spacing: 0.15em;
    }

    .waiting-dots {
        display: flex;
        gap: 0.4rem;
    }

    .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #c4b4a4;
        animation: bounce 1.4s ease-in-out infinite;
    }

    .dot:nth-child(2) {
        animation-delay: 0.2s;
    }

    .dot:nth-child(3) {
        animation-delay: 0.4s;
    }

    @keyframes bounce {
        0%, 80%, 100% {
            transform: translateY(0);
            opacity: 0.4;
        }
        40% {
            transform: translateY(-8px);
            opacity: 1;
        }
    }

    .status {
        font-family: "Nunito", sans-serif;
        font-size: 0.85rem;
        color: #7a6a4a;
        background: #faf0d8;
        border: 2px solid rgba(200, 180, 120, 0.3);
        border-radius: 12px;
        padding: 0.6rem 1.2rem;
        margin-bottom: 0.75rem;
        width: 100%;
        text-align: center;
        box-sizing: border-box;
    }
</style>
