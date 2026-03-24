<script lang="ts">
	import { Canvas } from "@threlte/core";
	import Scene from "./Scene.svelte";
	import Renderer from "./Renderer.svelte";
	import GameBrowser from "./GameBrowser.svelte";
	import { inGame, gameState, lastClickPos, pieceChoice, noPiecesMsg, webSocket, p1WebSocket, p2WebSocket, newGameState, graduatingLines, boopedOffPieces, slidingPieces, arcTrigger, placementLanded } from "./stores";
	import GameInfo from "./GameInfo.svelte";
	import AnimDebug from "./AnimDebug.svelte";

	let boopTexts: { id: number; x: number; y: number; color: string }[] = $state([]);
	let boopId = 0;
	let prevTurn = -1;
	let upgradeTexts: { id: number; x: number; y: number; color: string }[] = $state([]);
	let upgradeId = 0;
	let prevGraduatingLength = 0;

	const startOver = () => {
		if ($webSocket) $webSocket.close();
		if ($p1WebSocket) $p1WebSocket.close();
		if ($p2WebSocket) $p2WebSocket.close();
		$gameState = newGameState();
		$graduatingLines = [];
		$boopedOffPieces = [];
		$slidingPieces = [];
		$arcTrigger = { x: 0, y: 0, piece: 0, turn: -1 };
		$placementLanded = true;
		$inGame = false;
	};

	$effect(() => {
		const turn = $gameState.turnNumber;
		const boops = $gameState.boopMovement;
		if (turn !== prevTurn && prevTurn !== -1 && boops && boops.length > 0) {
			const hasRealBoop = boops.some(
				(b) => b.position.x !== 0 || b.position.y !== 0 || b.finalPosition.x !== 0 || b.finalPosition.y !== 0
			);
			if (hasRealBoop) {
				// Delay to match the sliding/push animation timing
				setTimeout(() => {
					const id = boopId++;
					const boopColor = (turn - 1) % 2 === 0 ? "orange" : "lightblue";
					boopTexts = [...boopTexts, { id, x: $lastClickPos.x, y: $lastClickPos.y, color: boopColor }];
					setTimeout(() => {
						boopTexts = boopTexts.filter((t) => t.id !== id);
					}, 800);
				}, 300);
			}
		}
		prevTurn = turn;
	});

	$effect(() => {
		const lines = $graduatingLines;
		if (lines.length > prevGraduatingLength) {
			const latest = lines[lines.length - 1];
			const color = latest.tile === 1 ? "orange" : "lightblue";
			const id = upgradeId++;
			upgradeTexts = [...upgradeTexts, { id, x: $lastClickPos.x, y: $lastClickPos.y, color }];
			setTimeout(() => {
				upgradeTexts = upgradeTexts.filter(t => t.id !== id);
			}, 1000);
		}
		prevGraduatingLength = lines.length;
	});
</script>

{#if !$inGame}
	<GameBrowser />
{:else}
	<GameInfo />
	<AnimDebug />
{/if}
<Canvas>
	<Renderer />
	<Scene />
</Canvas>

{#if $gameState.winner}
	<div class="winner-overlay">
		<div class="winner-card">
			<h1 class="winner-title" style="color: {$gameState.winner === 1 ? 'orange' : 'lightblue'}">Player {$gameState.winner} Wins!</h1>
			<p class="winner-sub">game over</p>
			<button class="start-over-btn" onclick={startOver}>Start Over</button>
		</div>
	</div>
{/if}

{#if $inGame}
	<div class="piece-selector">
		<button
			class="piece-btn"
			class:active={$pieceChoice == 0}
			onclick={() => { $pieceChoice = 0; }}
		>
			Kitten
		</button>
		<button
			class="piece-btn"
			class:active={$pieceChoice == 1}
			onclick={() => { $pieceChoice = 1; }}
		>
			Cat
		</button>
	</div>
{/if}

{#if $noPiecesMsg}
	<div class="no-pieces-msg">{$noPiecesMsg}</div>
{/if}

{#each boopTexts as boop (boop.id)}
	<span
		class="boop-text"
		style="left: {boop.x}px; top: {boop.y}px; color: {boop.color};"
	>
		boop!
	</span>
{/each}

{#each upgradeTexts as t (t.id)}
	<span class="upgrade-text" style="left: {t.x}px; top: {t.y}px; color: {t.color};">
		Upgraded!
	</span>
{/each}

<style>
	.winner-overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 10vh;
		background: rgba(90, 74, 58, 0.3);
		animation: fade-in 0.3s ease-out;
	}

	.winner-card {
		background: #faf6f0;
		border: 2px solid rgba(180, 160, 140, 0.3);
		border-radius: 24px;
		padding: 2.5rem 3.5rem;
		text-align: center;
		box-shadow: 0 8px 32px rgba(100, 80, 60, 0.15);
		animation: pop-in 0.4s ease-out;
	}

	.winner-title {
		font-family: "Cherry Bomb One", serif;
		font-size: 2.5rem;
		font-weight: 400;
		color: #5a4a3a;
		margin: 0;
	}

	.winner-sub {
		font-family: "Nunito", sans-serif;
		font-size: 0.85rem;
		font-weight: 600;
		color: #9a8a7a;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		margin: 0.5rem 0 0 0;
	}

	.start-over-btn {
		font-family: "Nunito", sans-serif;
		font-size: 1rem;
		font-weight: 700;
		color: #5a4a3a;
		background: #d4eef6;
		border: 2px solid rgba(142, 200, 219, 0.4);
		border-radius: 14px;
		padding: 0.7rem 2rem;
		margin-top: 1.2rem;
		cursor: pointer;
		transition: transform 0.2s ease, box-shadow 0.2s ease;
	}

	.start-over-btn:hover {
		transform: scale(1.05);
		box-shadow: 0 3px 12px rgba(100, 80, 60, 0.12);
	}

	.start-over-btn:active {
		transform: scale(0.97);
	}

	@keyframes fade-in {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	@keyframes pop-in {
		0% { opacity: 0; transform: scale(0.8); }
		100% { opacity: 1; transform: scale(1); }
	}

	.piece-selector {
		position: fixed;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		display: flex;
		gap: 0.5rem;
		background: rgba(250, 246, 240, 0.92);
		border: 2px solid rgba(180, 160, 140, 0.3);
		border-radius: 16px;
		padding: 0.4rem;
		box-shadow: 0 4px 16px rgba(100, 80, 60, 0.08);
	}

	.piece-btn {
		font-family: "Nunito", sans-serif;
		font-size: 1rem;
		font-weight: 700;
		color: #5a4a3a;
		background: none;
		border: 2px solid transparent;
		border-radius: 12px;
		padding: 0.6rem 1.5rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.piece-btn:hover {
		background: rgba(180, 160, 140, 0.15);
	}

	.piece-btn.active {
		background: #d4eef6;
		border-color: rgba(142, 200, 219, 0.4);
	}

	.no-pieces-msg {
		position: fixed;
		bottom: 6rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 100;
		font-family: "Cherry Bomb One", serif;
		font-size: 1.2rem;
		color: #5a4a3a;
		background: rgba(250, 246, 240, 0.95);
		border: 2px solid rgba(219, 170, 142, 0.5);
		border-radius: 14px;
		padding: 0.5rem 1.2rem;
		animation: shake 0.4s ease-out;
		pointer-events: none;
	}

	@keyframes shake {
		0%, 100% { transform: translateX(-50%) rotate(0); }
		25% { transform: translateX(-50%) rotate(-3deg); }
		75% { transform: translateX(-50%) rotate(3deg); }
	}

	.boop-text {
		position: fixed;
		z-index: 100;
		pointer-events: none;
		font-family: "Cherry Bomb One", serif;
		font-size: 1.4rem;
		color: #5a4a3a;
		transform: translate(-50%, -100%);
		animation: boop-float 0.8s ease-out forwards;
	}

	.upgrade-text {
		position: fixed;
		z-index: 100;
		pointer-events: none;
		font-family: "Cherry Bomb One", serif;
		font-size: 1.8rem;
		transform: translate(-50%, -100%);
		animation: boop-float 1.0s ease-out forwards;
	}

	@keyframes boop-float {
		0% {
			opacity: 1;
			transform: translate(-50%, -100%) scale(0.5);
		}
		20% {
			opacity: 1;
			transform: translate(-50%, -100%) scale(1.2);
		}
		40% {
			transform: translate(-50%, -100%) scale(1);
		}
		100% {
			opacity: 0;
			transform: translate(-50%, -180%) scale(1);
		}
	}
</style>
