<script lang="ts">
	import { gameState } from "./stores";
	import type { Player } from "./stores";
</script>

{#snippet playerBox(name: string, player: Player, color: string)}
	<h3 style="color: {color}">{name}</h3>
	<table class="stats-table"><tbody>
		<tr><td class="stat-label">Kittens ({player.kittens})</td><td class="stat-faces">{':3 '.repeat(player.kittens).trim()}</td></tr>
		<tr><td class="stat-label">Cats ({player.cats})</td><td class="stat-faces">{'>:3 '.repeat(player.cats).trim()}</td></tr>
		<tr><td class="stat-label">Placed</td><td class="stat-faces">{player.placed}</td></tr>
	</tbody></table>
{/snippet}

<div class="game-info">
	<h1 class="title">boop.</h1>
	<p class="turn-label" style="color: {$gameState.turnNumber % 2 == 0 ? 'orange' : 'lightblue'}">
		{#if $gameState.turnNumber % 2 == 0}
			Player 1's Turn
		{:else}
			Player 2's Turn
		{/if}
	</p>
	<p class="turn-number">Turn {$gameState.turnNumber}</p>
</div>

{#if $gameState.state === "MAX_WAITING"}
	<div class="board-message">
		<p class="state-alert"><strong>Board full!</strong> Select a piece to remove. Kittens graduate into cats!</p>
	</div>
{:else if $gameState.state === "MULTIPLE_WAITING"}
	<div class="board-message">
		<p class="state-alert"><strong>Multiple rows!</strong> Click the middle piece of a row to select it. Kittens graduate into cats!</p>
	</div>
{/if}

<div class="player-box player-1">
	{@render playerBox("Player 1", $gameState.p1, "orange")}
</div>

<div class="player-box player-2">
	{@render playerBox("Player 2", $gameState.p2, "lightblue")}
</div>

<style>
	.game-info {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 10;
		background: rgba(250, 246, 240, 0.92);
		border: 2px solid rgba(180, 160, 140, 0.3);
		border-radius: 20px;
		padding: 1.2rem 1.5rem;
		margin: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		box-shadow: 0 4px 16px rgba(100, 80, 60, 0.08);
	}

	.title {
		font-family: "Cherry Bomb One", serif;
		font-size: 1.8rem;
		font-weight: 400;
		color: #5a4a3a;
		margin: 0;
		line-height: 1.1;
	}

	.turn-label {
		font-family: "Nunito", sans-serif;
		font-size: 1rem;
		font-weight: 700;
		color: #5a4a3a;
		margin: 0.25rem 0 0 0;
	}

	.turn-number {
		font-family: "Nunito", sans-serif;
		font-size: 0.8rem;
		font-weight: 600;
		color: #9a8a7a;
		margin: 0;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.board-message {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -320%);
		z-index: 20;
	}

	.state-alert {
		font-family: "Nunito", sans-serif;
		font-size: 0.85rem;
		font-weight: 700;
		color: #7a6a4a;
		background: #faf0d8;
		border: 2px solid rgba(200, 180, 120, 0.3);
		border-radius: 10px;
		padding: 0.4rem 0.8rem;
		margin: 0.4rem 0 0 0;
		text-align: center;
	}

	.player-box {
		position: absolute;
		z-index: 10;
		background: rgba(250, 246, 240, 0.92);
		border: 2px solid rgba(180, 160, 140, 0.3);
		border-radius: 20px;
		padding: 0.8rem 1.2rem;
		margin: 1rem;
		box-shadow: 0 4px 16px rgba(100, 80, 60, 0.08);
	}

	.player-box h3 {
		font-family: "Cherry Bomb One", serif;
		font-size: 1.1rem;
		font-weight: 400;
		color: #5a4a3a;
		margin: 0 0 0.3rem 0;
	}

	.stats-table {
		width: 100%;
		border-collapse: collapse;
		font-family: "Nunito", sans-serif;
	}

	.stats-table td {
		padding: 0.15rem 0;
		font-size: 0.85rem;
		color: #5a4a3a;
	}

	.stat-label {
		font-weight: 700;
		white-space: nowrap;
		padding-right: 0.6rem !important;
	}

	.stat-faces {
		font-weight: 600;
		color: #9a8a7a !important;
		word-break: break-all;
	}

	.player-1 {
		top: 0;
		right: 0;
		border-left: 4px solid #d4eef6;
	}

	.player-2 {
		bottom: 0;
		left: 0;
		border-left: 4px solid #f6ddd4;
	}
</style>
