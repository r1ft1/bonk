<script lang="ts">
	import { gameState } from "./stores";
</script>

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
	<h3 style="color: orange">Player 1</h3>
	<div class="stats">
		<span>Kittens: {$gameState.p1.kittens}</span>
		<span>Cats: {$gameState.p1.cats}</span>
	</div>
	<p class="placed">Placed: {$gameState.p1.placed}</p>
</div>

<div class="player-box player-2">
	<h3 style="color: lightblue">Player 2</h3>
	<div class="stats">
		<span>Kittens: {$gameState.p2.kittens}</span>
		<span>Cats: {$gameState.p2.cats}</span>
	</div>
	<p class="placed">Placed: {$gameState.p2.placed}</p>
</div>

<style>
	@import url("https://fonts.googleapis.com/css2?family=Cherry+Bomb+One&family=Nunito:wght@400;600;700&display=swap");

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

	.stats {
		display: flex;
		gap: 0.75rem;
		background: none;
		border: none;
		box-shadow: none;
		padding: 0;
		margin: 0;
	}

	.stats span {
		font-family: "Nunito", sans-serif;
		font-size: 0.85rem;
		font-weight: 600;
		color: #5a4a3a;
	}

	.placed {
		font-family: "Nunito", sans-serif;
		font-size: 0.75rem;
		font-weight: 400;
		color: #9a8a7a;
		margin: 0.2rem 0 0 0;
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
