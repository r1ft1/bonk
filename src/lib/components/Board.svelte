<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import {
		interactivity,
		useTexture,
		Outlines,
		Edges,
	} from "@threlte/extras";
	import * as THREE from "three";
	import { gameState, pieceChoice, webSocket } from "./stores.svelte";
	import Cat from "./Cat.svelte";
	import Kitten from "./Kitten.svelte";
	import { GLTFLoader } from "three/examples/jsm/Addons.js";
	import { animate } from "motion";
	import Piece from "./Piece.svelte";

	const map = useTexture("/tile.png", {
		transform: (texture) => {
			texture.wrapS = THREE.RepeatWrapping;
			texture.wrapT = THREE.RepeatWrapping;
			texture.repeat.set(4, 4);
			return texture;
		},
	});

	let lastMove = { x: 0, z: 0 };
	let color = "orange";

	$: if ($gameState.turnNumber % 2 === 0) {
		color = "orange";
		console.log(color);
	} else {
		color = "lightblue";
		console.log(color);
	}


	const wsSendMove = (move: THREE.Vector3) => {
		$webSocket.send(
			JSON.stringify({
				position: { x: move.x + 2.5, y: move.z + 2.5 },
				piece: $pieceChoice,
			})
		);
		console.log(lastMove);
	};

	interactivity();

	const { camera, scene, renderMode, autoRender } = useThrelte();

	let planeMesh: THREE.Mesh;
	let highlightMesh: THREE.Mesh;

	$: {
		lastMove.x = $gameState.placed.position.x;
		lastMove.z = $gameState.placed.position.y;
		console.log("lastmove: ",lastMove, "color: ", color);
	}

	// for (let j = 0; j < $gameState.board.length; j++) {
	// 	for (let i=0; i<$gameState.board[j].length; i++) {
	// 		console.log(j,i,$gameState.board[j][i]);
	// 	}
	// }

	// let objects: THREE.Mesh[] | undefined = [];
	let time = 0;
	useTask((delta) => {
		time += delta;
		(highlightMesh.material as THREE.Material).opacity = 1 + Math.sin(time);
	});
</script>

<!-- Board Base -->
{#await map then value}
	<T.Mesh position.y={-0.5}>
		<T.BoxGeometry args={[6, 1, 6]} />
		<T.MeshBasicMaterial map={value} />
		<Outlines color="black" thickness={0.02} />
		<Edges color="black" />
	</T.Mesh>
{/await}

<!-- Piece generation from $gameState.board -->
{#each $gameState.board as row, i}
	{#each row as cell, j}
		{#if j == lastMove.x && i == lastMove.z}
			<Piece piece={$gameState.placed.piece} position={[j - 2.5, 2, i - 2.5]} placed={true}/>
		{:else if cell == 1}
			<Kitten
				position={[j - 2.5, 2, i - 2.5]}
				color={"orange"}
				placed={false}
			/>
		{:else if cell == 8}
			<Kitten
				position={[j - 2.5, 0, i - 2.5]}
				color={"lightblue"}
				placed={false}
			/>
		{:else if cell == 2}
			<Cat
				position={[j - 2.5, 0, i - 2.5]}
				color={"orange"}
				placed={false}
			/>
		{:else if cell == 9}
			<Cat
				position={[j - 2.5, 0, i - 2.5]}
				color={"lightblue"}
				placed={false}
			/>
		{/if}
	{/each}
{/each}

<!-- Invisible Ground Plane -->
<T.Mesh
	rotation.x={-Math.PI / 2}
	position.y={0}
	visible={false}
	name="ground"
	on:create={({ ref }) => {
		planeMesh = ref;
	}}
	on:pointermove={(e) => {
		if (e.intersections.length > 0) {
			const { x, z } = e.intersections[0].point;
			highlightMesh.position.set(
				Math.floor(x) + 0.5,
				0.01,
				Math.floor(z) + 0.5
			);
		}

		// const objectExists = objects.find((obj) => {
		// 	return (
		// 		obj.position.x === highlightMesh.position.x &&
		// 		obj.position.z === highlightMesh.position.z
		// 	);
		// });

		// if (!objectExists) {
		// 	// @ts-ignore
		// 	highlightMesh.material.color.setHex(0xffffff);
		// } else {
		// 	// @ts-ignore
		// 	highlightMesh.material.color.setHex(0xff0000);
		// }
	}}
	on:pointerdown={(e) => {
		console.log("pointerdown", highlightMesh.position);
		// const objectExists = objects.find((obj) => {
		// 	return (
		// 		obj.position.x === highlightMesh.position.x &&
		// 		obj.position.z === highlightMesh.position.z
		// 	);
		// });
		// if (!objectExists) {
		wsSendMove(highlightMesh.position);
		// @ts-ignore
		highlightMesh.material.color.setHex(0xff0000);
		// }
		// console.log(scene.children.length);
	}}
>
	<T.PlaneGeometry args={[6, 6]} />
	<T.MeshBasicMaterial side={THREE.DoubleSide} />
</T.Mesh>

<!-- Tile Cursor -->
<T.Mesh
	rotation.x={-Math.PI / 2}
	position.y={0}
	on:create={({ ref }) => {
		highlightMesh = ref;
	}}
>
	<T.PlaneGeometry args={[1, 1]} />
	<T.MeshBasicMaterial side={THREE.DoubleSide} transparent={true} />
</T.Mesh>

<!-- Grid -->
<T.GridHelper args={[6, 6]} position.y={0.01} />
