<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import Board from "./Board.svelte";
	import { Stars } from "@threlte/extras";
	import { gameState } from "./stores";
	import { inGame } from "./stores";
	const { camera, renderMode } = useThrelte();
	renderMode.set("always");

	let innerWidth = $state(window.innerWidth);
	let innerHeight = $state(window.innerHeight);
	let cameraRef: any = $state(null);

	$effect(() => {
		if (!cameraRef) return;
		if (innerWidth < 768) {
			cameraRef.position.set(0, 14, 6);
			cameraRef.lookAt(0, -1, 0);
		} else {
			cameraRef.position.set(10, 10, 10);
			cameraRef.lookAt(0, 1, 0);
		}
	});
</script>

<svelte:window bind:innerWidth={innerWidth} bind:innerHeight={innerHeight} />

<T.PerspectiveCamera
	makeDefault
	position={[10, 10, 10]}
	oncreate={(ref) => { cameraRef = ref; }}
>
</T.PerspectiveCamera>

<T.DirectionalLight position={[-1, 10, 0]} intensity={1} />
<T.AmbientLight intensity={0.7} />


{#if $inGame}
	<Board />
{/if}
<Stars />
