<script lang="ts">
	import { gameState, isMobile } from "./stores";
	import type { Player } from "./stores";

	// SVG silhouettes traced from 3D model side profiles
	// Kitten: low body, tail curving up on left, ears on right
	const kittenSvg = `<svg viewBox="0 0 20 16" fill="currentColor"><path d="M1 14 L1 10 L2 10 L2 8 L1 6 L1 4 L2 3 L3 4 L3 6 L4 7 L5 8 L8 8 L9 6 L10 5 L12 5 L13 6 L14 8 L16 8 L17 6 L18 5 L19 6 L19 8 L18 9 L18 14 Z"/></svg>`;

	// Cat: tall trapezoid body, V-notch ears at top
	const catSvg = `<svg viewBox="0 0 16 20" fill="currentColor"><path d="M2 18 L4 6 L3 4 L3 2 L5 1 L6 3 L7 5 L8 6 L9 5 L10 3 L11 1 L13 2 L13 4 L12 6 L14 18 Z"/></svg>`;
</script>

{#snippet playerBox(name: string, player: Player, color: string)}
	<h3 style="color: {color}">{name}</h3>
	<div class="desktop-pieces">
		<div class="piece-row">
			<span class="stat-label">Kittens ({player.kittens})</span>
			<div class="piece-icons">
				{#each Array(player.kittens) as _}
					<span class="piece-icon" style="color: {color}">{@html kittenSvg}</span>
				{/each}
			</div>
		</div>
		<div class="piece-row">
			<span class="stat-label">Cats ({player.cats})</span>
			<div class="piece-icons">
				{#each Array(player.cats) as _}
					<span class="piece-icon piece-icon-cat" style="color: {color}">{@html catSvg}</span>
				{/each}
			</div>
		</div>
		<div class="piece-row">
			<span class="stat-label">Placed: {player.placed}</span>
		</div>
	</div>
{/snippet}

{#if $isMobile}
	<!-- Mobile layout -->
	<div class="mobile-bar">
		<div class="mobile-player" style="opacity: {$gameState.turnNumber % 2 === 0 ? 1 : 0.5}">
			<span class="compact-name" style="color: orange">P1</span>
			<div class="mobile-pieces">
				{#each Array($gameState.p1.kittens) as _}
					<span class="piece-icon" style="color: orange">{@html kittenSvg}</span>
				{/each}
				{#each Array($gameState.p1.cats) as _}
					<span class="piece-icon piece-icon-cat" style="color: orange">{@html catSvg}</span>
				{/each}
			</div>
		</div>
		<span class="mobile-turn">Turn {$gameState.turnNumber}</span>
		<div class="mobile-player" style="opacity: {$gameState.turnNumber % 2 === 1 ? 1 : 0.5}">
			<span class="compact-name" style="color: lightblue">P2</span>
			<div class="mobile-pieces">
				{#each Array($gameState.p2.kittens) as _}
					<span class="piece-icon" style="color: lightblue">{@html kittenSvg}</span>
				{/each}
				{#each Array($gameState.p2.cats) as _}
					<span class="piece-icon piece-icon-cat" style="color: lightblue">{@html catSvg}</span>
				{/each}
			</div>
		</div>
	</div>
{:else}
	<!-- Desktop layout -->
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

	<div class="player-box player-1">
		{@render playerBox("Player 1", $gameState.p1, "orange")}
	</div>

	<div class="player-box player-2">
		{@render playerBox("Player 2", $gameState.p2, "lightblue")}
	</div>
{/if}

{#if $gameState.state === "MAX_WAITING"}
	<div class="board-message">
		<p class="state-alert"><strong>Board full!</strong> Select a piece to remove. Kittens graduate into cats!</p>
	</div>
{:else if $gameState.state === "MULTIPLE_WAITING"}
	<div class="board-message">
		<p class="state-alert"><strong>Multiple rows!</strong> Click the middle piece of a row to select it. Kittens graduate into cats!</p>
	</div>
{/if}

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
		font-family: "Nunito", sans-serif;
		font-size: 0.85rem;
		font-weight: 700;
		color: #5a4a3a;
		white-space: nowrap;
	}

	.stat-faces {
		font-weight: 600;
		color: #9a8a7a !important;
		word-break: break-all;
	}

	.desktop-pieces {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.piece-row {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.piece-icons {
		display: flex;
		flex-wrap: wrap;
		gap: 0.1rem;
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

	/* Mobile bar */
	.mobile-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.6rem 1.2rem;
		background: rgba(250, 246, 240, 0.92);
		border-bottom: 2px solid rgba(180, 160, 140, 0.2);
		box-shadow: 0 2px 8px rgba(100, 80, 60, 0.06);
	}

	.mobile-player {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.15rem;
		transition: opacity 0.2s ease;
	}

	.mobile-player:last-child {
		align-items: flex-end;
	}

	.compact-name {
		font-family: "Cherry Bomb One", serif;
		font-size: 1.1rem;
		font-weight: 400;
	}

	.mobile-pieces {
		display: flex;
		flex-wrap: wrap;
		gap: 0.05rem;
		max-width: 120px;
	}

	.piece-icon {
		width: 14px;
		height: 14px;
		display: inline-flex;
	}

	.piece-icon-cat {
		width: 16px;
		height: 16px;
	}

	.mobile-turn {
		font-family: "Nunito", sans-serif;
		font-size: 0.75rem;
		font-weight: 600;
		color: #9a8a7a;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}
</style>
