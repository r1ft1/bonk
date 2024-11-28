<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import CameraControls from "./CameraControls/CameraControls.svelte";
	import { cameraControls, mesh } from "./CameraControls/stores";
	import Board from "./Board.svelte";
	import Logo from "./Logo.svelte";
	import { Stars } from "@threlte/extras";
	import { gameState } from "./stores.svelte";
	import { inGame } from "./stores.svelte";
	import GameInfo from "./GameInfo.svelte";
	const { camera, renderMode } = useThrelte();
	renderMode.set("always");
</script>

<T.PerspectiveCamera
	makeDefault
	position={[10, 10, 10]}
	on:create={({ ref }) => {
		ref.lookAt(0, 1, 0);
		// camera = ref;
	}}
>
	<!-- <CameraControls
		on:create={({ ref }) => {
			$cameraControls = ref;
			console.log($cameraControls)
			console.log(ref)
		}}
	/> -->
</T.PerspectiveCamera>

<T.DirectionalLight position={[-1, 10, 0]} intensity={1} />
<T.AmbientLight intensity={0.7} />

<Logo text={"boop!"} position={[0, 4, 0]} />

{#if $gameState.winner == 1}
	<Logo text={"Player 1 Wins!"} position={[-4, 1, 0]} />
{:else if $gameState.winner == 2}
	<Logo text={"Player 2 Wins!"} position={[-4, 1, 0]} />
{/if}

{#if $inGame}
	<Board />
{/if}
<Stars />
