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

	const map = useTexture("/tile.png", {
		transform: (texture) => {
			texture.wrapS = THREE.RepeatWrapping;
			texture.wrapT = THREE.RepeatWrapping;
			texture.repeat.set(4, 4);
			return texture;
		},
	});

	const gravity = 10;

	let newPiece = false;
	let newPieceRef: THREE.Group;
	let previousBoard = $gameState.board;
	let lastMove = { x: 0, z: 0 };
	let color = "orange";

	$: if ($gameState.turnNumber % 2 === 0) {
		color = "orange";
	} else {
		color = "lightblue";
	}

	// executes when the server validates the move
	$: if (previousBoard !== $gameState.board) {
		newPiece = true;
		console.log("board changed");
		// animatePiecePlacement();
	}

	const wsSendMove = (move: THREE.Vector3) => {
		$webSocket.send(
			JSON.stringify({
				position: { x: move.x + 2.5, y: move.z + 2.5 },
				piece: $pieceChoice,
			})
		);
		lastMove = { x: move.x, z: move.z };
	};

	interactivity();

	const { camera, scene, renderMode, autoRender } = useThrelte();

	let planeMesh: THREE.Mesh;
	let highlightMesh: THREE.Mesh;

	// let objects: THREE.Mesh[] | undefined = [];
	let time = 0;
	useTask((delta) => {
		time += delta;
		(highlightMesh.material as THREE.Material).opacity = 1 + Math.sin(time);

		if (newPieceRef && newPieceRef.position.y > 0) {
			// newPieceRef.position.set(
			// 	newPieceRef.position.x,
			// 	newPieceRef.position.y - delta * 1,
			// 	newPieceRef.position.z
			// );
			// newPieceRef.position.y -= delta*0.1;
		} else if (newPiece) {
			previousBoard = $gameState.board;
			newPiece = false;
		}
	});
</script>

<!-- Board Base -->
{#await map then value}
	<T.Mesh position.y={-0.5}>
		<T.BoxGeometry args={[6, 1, 6]} />
		<T.MeshBasicMaterial map={value} />
		<Outlines color="black" thickness={0.02}/>
		<Edges color="black" />
	</T.Mesh>
{/await}

{#if newPiece}
	{#if $pieceChoice == 0}
		<Kitten
			position={[lastMove.x, 0, lastMove.z]}
			scale={0.5}
			{color}
			bind:ref={newPieceRef}
		/>
	{:else if $pieceChoice == 1}
		<Cat
			position={[lastMove.x, 0, lastMove.z]}
			scale={0.5}
			{color}
			bind:ref={newPieceRef}
		/>
	{/if}
{/if}

<!-- Piece generation from $gameState.board -->
{#each previousBoard as column, i}
	{#each column as cell, j}
		{#if cell == 1}
			<Kitten
				position={[j - 2.5, 0, i - 2.5]}
				scale={0.5}
				color={"orange"}
			/>
		{:else if cell == 8}
			<Kitten
				position={[j - 2.5, 0, i - 2.5]}
				scale={0.5}
				color={"lightblue"}
			/>
		{:else if cell == 2}
			<Cat
				position={[j - 2.5, 0, i - 2.5]}
				scale={0.5}
				color={"orange"}
			/>
		{:else if cell == 9}
			<Cat
				position={[j - 2.5, 0, i - 2.5]}
				scale={0.5}
				color={"lightblue"}
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
