<script lang="ts">
	import { Canvas } from "@threlte/core";
	import Scene from "./Scene.svelte";
	import {
		gameState,
		pieceChoice,
		webSocket,
		type GameState,
	} from "./stores.svelte";
	import { onMount } from "svelte";
	import { Centrifuge } from "centrifuge";
	import Renderer from "./Renderer.svelte";

	//0 - kitten, 1 - cat
	// let pieceChoiceOptions = [0, 1];

	$: console.log($pieceChoice);

	// const centrifuge = new Centrifuge(
	// 	"ws://localhost:8000/connection/websocket",
	// 	{
	// 		token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM3MjIiLCJleHAiOjE3MzEzNTY1MDgsImlhdCI6MTczMDc1MTcwOH0.HqiZ-jQnPw9EbQRRG8i_RcxNDhUT9KsHJgS70sOtYr4",
	// 	}
	// );

	// centrifuge
	// 	.on("connecting", (ctx) => {
	// 		console.log(`connecting: ${ctx.code}, ${ctx.reason}`);
	// 	})
	// 	.on("connected", (ctx) => {
	// 		console.log(`connected over ${ctx.transport}`);
	// 	})
	// 	.on("disconnected", (ctx) => {
	// 		console.log(`disconnected: ${ctx.code}, ${ctx.reason}`);
	// 	})
	// 	.connect();

	// const sub = centrifuge.newSubscription("channel");

	// sub.on("publication", function (ctx) {
	// 	counter = ctx.data.value;
	// 	// document.title = ctx.data.value;
	// })
	// 	.on("subscribing", function (ctx) {
	// 		console.log(`subscribing: ${ctx.code}, ${ctx.reason}`);
	// 	})
	// 	.on("subscribed", function (ctx) {
	// 		console.log("subscribed", ctx);
	// 	})
	// 	.on("unsubscribed", function (ctx) {
	// 		console.log(`unsubscribed: ${ctx.code}, ${ctx.reason}`);
	// 	})
	// 	.subscribe();

	onMount(async () => {
		// const getWebsocketUpgradeResponse =
		// await fetch("ws://localhost:8080/ws",{
		//   method: "GET",
		//   headers: {
		//     "Host": "http://localhost:8080",
		//     "Upgrade": "websocket",
		//     "Connection": "Upgrade",
		//     "Sec-WebSocket-Key": "",
		//     "Sec-WebSocket-Version": "13",
		//   }
		// });
		// const testGameState = {
		// 	board: [
		// 		[1, 0, 0, 0, 0, 0],
		// 		[0, 0, 0, 0, 0, 0],
		// 		[0, 0, 0, 1, 0, 0],
		// 		[0, 0, 0, 0, 0, 0],
		// 		[0, 0, 0, 0, 0, 0],
		// 		[0, 1, 0, 0, 0, 1],
		// 	],
		// 	turnNumber: 0,
		// 	p1: {
		// 		kittens: 8,
		// 		cats: 0,
		// 		placed: 0,
		// 	},
		// 	p2: {
		// 		kittens: 8,
		// 		cats: 0,
		// 		placed: 0,
		// 	},
		// };
		// const boardJSON = {
		// 	board: testGameState.board,
		// };
		// const postResponse = await fetch("http://localhost:8080/postBoard", {
		// 	method: "POST",
		// 	headers: {
		// 		"Content-Type": "application/json",
		// 	},
		// 	body: JSON.stringify(boardJSON),
		// });
		// console.log(boardJSON);
	});

	$webSocket.addEventListener("open", function (event) {
		// $webSocket.send(JSON.stringify({ position: { X: 1, Y: 1 } }));
	});

	$webSocket.addEventListener("message", function (event) {
		$gameState = JSON.parse(event.data);
		console.log($gameState);
	});
</script>

<div class="game-info">
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

<Canvas>
	<Renderer />
	<Scene />
</Canvas>

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
