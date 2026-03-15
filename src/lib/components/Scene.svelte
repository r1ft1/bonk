<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import Board from "./Board.svelte";
	import { Stars } from "@threlte/extras";
	import { gameState } from "./stores";
	import { inGame } from "./stores";
	const { camera, renderMode } = useThrelte();
	renderMode.set("always");
</script>

<T.PerspectiveCamera
	makeDefault
	position={[10, 10, 10]}
	oncreate={(ref) => {
		ref.lookAt(0, 1, 0);
		// camera = ref;
	}}
>
</T.PerspectiveCamera>

<T.DirectionalLight position={[-1, 10, 0]} intensity={1} />
<T.AmbientLight intensity={0.7} />


{#if $inGame}
	<Board />
{/if}
<Stars />
