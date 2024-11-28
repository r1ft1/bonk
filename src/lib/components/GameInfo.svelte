<script lang="ts">
	import { gameState, pieceChoice } from "./stores.svelte";
</script>

<div class="game-info">
	<h1>Game Info</h1>
	<h2>
		{#if $gameState.turnNumber % 2 == 0}
			Player 1's Turn
		{:else}
			Player 2's Turn
		{/if}
	</h2>

	<h2>Turn Number: {$gameState.turnNumber}</h2>
	<input
		type="radio"
		bind:group={$pieceChoice}
		id="kitten"
		value="0"
		checked
	/>
	<label for="kitten">Kitten</label>
	<input type="radio" bind:group={$pieceChoice} id="cat" value="1" />
	<label for="cat">Cat</label>
</div>

<div class="player-box player1">
	<h3>Player 1</h3>
	<p>Kittens: {$gameState.p1.kittens} Cats: {$gameState.p1.cats}</p>
	<p>Placed: {$gameState.p1.placed}</p>
</div>

<div class="player-box player2">
	<h3>Player 2</h3>
	<p>Kittens: {$gameState.p2.kittens} Cats: {$gameState.p2.cats}</p>
	<p>Placed: {$gameState.p2.placed}</p>
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
	.game-info {
		position: absolute;
		top: 0;
		left: 0;
		padding: 1rem;
	}
	.player1 {
		top: 0;
		right: 0;
	}
	.player2 {
		bottom: 0;
		left: 0;
	}
	.player-box {
		position: absolute;
		padding: 1rem;
	}
</style>
