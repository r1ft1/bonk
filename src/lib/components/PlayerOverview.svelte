<script lang="ts">
	import { gameState } from "./stores.svelte";

	export let player = 1;
	let playerState = $gameState.p1;

	if (player == 1) {
		playerState = $gameState.p1;
	} else {
		playerState = $gameState.p2;
	}
</script>
<div class="game-info">
    <h3>Player {player}</h3>
    <p>Kittens: {playerState.kittens} Cats: {playerState.cats}</p>
    <p>Placed: {playerState.placed}</p>
</div>


<style>
	* {
		color: white;
	}
	.game-info {
		position: absolute;
		top: 0;
		left: 0;
		background-color: rgba(0, 0, 0, 0.5);
		padding: 1rem;
		display: flex;
	}
</style>
