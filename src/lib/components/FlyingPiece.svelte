<script lang="ts">
	import { Group } from "three";
	import { T } from "@threlte/core";
	import { useGltf, Outlines, Edges } from "@threlte/extras";
	import { animate } from "motion";
	import { onMount } from "svelte";

	let {
		tile,
		startPos,
		direction,
		onDone,
	}: {
		tile: number;
		startPos: [number, number, number];
		direction: [number, number];
		onDone: () => void;
	} = $props();

	const ref = new Group();
	ref.position.set(startPos[0], startPos[1], startPos[2]);

	const isCat = tile === 2 || tile === 9;
	const color = tile === 1 || tile === 2 ? "orange" : "lightblue";
	const gltf = useGltf(isCat ? "/cat.glb" : "/kitten.glb");

	onMount(() => {
		const isP1 = tile === 1 || tile === 2;
		const cornerX = isP1 ? -4 : 4;
		const cornerZ = isP1 ? -4 : 4;

		// Phase 1: Hop up
		animate(ref.position, { y: startPos[1] + 0.8 }, { duration: 0.15, ease: "easeOut" });

		// Phase 2: Slide to player's corner
		setTimeout(() => {
			animate(
				ref.position,
				{ x: cornerX, y: 0.52, z: cornerZ },
				{ duration: 0.7, ease: "easeInOut", onComplete: onDone },
			);
		}, 150);
	});
</script>

{#await gltf then gltf}
	<T is={ref} dispose={false}>
		{#if isCat}
			<T.Mesh geometry={gltf.nodes.Cube.geometry} scale={[0.5, 0.5, 0.5]}>
				<T.MeshStandardMaterial {color} />
				<Outlines color="black" />
				<Edges color="black" />
			</T.Mesh>
		{:else}
			<T.Mesh geometry={gltf.nodes.Kitten.geometry} scale={[0.5, 0.5, 0.5]}>
				<T.MeshStandardMaterial {color} />
				<Outlines color="black" />
			</T.Mesh>
		{/if}
	</T>
{/await}
