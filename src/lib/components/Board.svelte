<script lang="ts">
	import { T, useThrelte, useTask } from "@threlte/core";
	import {
		interactivity,
		useTexture,
		Outlines,
		Edges,
	} from "@threlte/extras";
	import * as THREE from "three";
	import {
		gameState,
		pieceChoice,
		webSocket,
		type ServerMessage,
		p1WebSocket,
		p2WebSocket,
		lastClickPos,
		noPiecesMsg,
		graduatingLines,
		boopedOffPieces,
	} from "./stores";
	import { animate } from "motion";
	import Piece from "./Piece.svelte";
	import GraduatingLine from "./GraduatingLine.svelte";
	import FlyingPiece from "./FlyingPiece.svelte";

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

	$effect(() => {
		if ($gameState.turnNumber % 2 === 0) {
			color = "orange";
			console.log(color);
		} else {
			color = "lightblue";
			console.log(color);
		}
	});

	const wsSendMove = (move: THREE.Vector3) => {
		if ($gameState.state === "WAITING") {
			const isP1 = $gameState.turnNumber % 2 === 0;
			const player = isP1 ? $gameState.p1 : $gameState.p2;
			const pieceName = $pieceChoice == 0 ? "kittens" : "cats";
			const available = $pieceChoice == 0 ? player.kittens : player.cats;
			if (available <= 0) {
				$noPiecesMsg = `No ${pieceName} left!`;
				setTimeout(() => { $noPiecesMsg = ""; }, 1500);
				return;
			}
		}

		if ($webSocket != null) {
			$webSocket.send(
				JSON.stringify({
					position: {
						x: move.x + 2.5,
						y: move.z + 2.5,
					},
					piece: $pieceChoice,
				}),
			);
		}

		if ($p1WebSocket != null && $gameState.turnNumber % 2 === 0) {
			$p1WebSocket.send(
				JSON.stringify({
					position: {
						x: move.x + 2.5,
						y: move.z + 2.5,
					},
					piece: $pieceChoice,
				}),
			);
		} else if (
			$p2WebSocket != null &&
			$gameState.turnNumber % 2 === 1
		) {
			$p2WebSocket.send(
				JSON.stringify({
					position: {
						x: move.x + 2.5,
						y: move.z + 2.5,
					},
					piece: $pieceChoice,
				}),
			);
		}
		console.log(lastMove);
		console.log($gameState);
	};

	interactivity();

	//const { camera, scene, renderMode, autoRender } = useThrelte();

	let planeMesh: THREE.Mesh;
	let highlightMesh: THREE.Mesh;

	$effect(() => {
		lastMove.x = $gameState.placed.position.x;
		lastMove.z = $gameState.placed.position.y;
		console.log("lastmove: ", lastMove, "color: ", color);
	});

	function lineIndicesFor(x: number, y: number): number[] {
		if ($gameState.lines == null) return [];
		const indices: number[] = [];
		$gameState.lines.forEach((line, idx) => {
			if (line.some(pos => pos.x === x && pos.y === y)) indices.push(idx);
		});
		return indices;
	}

	function isMiddlePiece(x: number, y: number): boolean {
		if ($gameState.lines == null) return false;
		return $gameState.lines.some(line => line[1].x === x && line[1].y === y);
	}

	//mobile touch down will move highlightMesh to the touched position
	//ray trace the touch down to shoot a ray from the camera at that point, intersect with the ground plane to find the specific tile
	window.addEventListener("touchstart", (e) => {
		const touch = e.touches[0];
		const touchX = touch.clientX;
		const touchY = touch.clientY;
	});

	//console.log($gameState.lines);
	$effect(() => {
		if (
			$gameState.state == "MULTIPLE_WAITING" &&
			$gameState.lines != null
		) {
			console.log("Multiple waiting!!!", JSON.stringify($gameState.lines));
		}
	});
	//
	// for (let j = 0; j < $gameState.board.length; j++) {
	// 	for (let i=0; i<$gameState.board[j].length; i++) {
	// 		console.log(j,i,$gameState.board[j][i]);
	// 	}
	// }

	// let objects: THREE.Mesh[] | undefined = [];
	let time = 0;
	useTask((delta) => {
		time += delta;
		if (highlightMesh) {
			if ($gameState.winner) {
				(highlightMesh.material as THREE.Material).opacity = 0;
			} else {
				(highlightMesh.material as THREE.Material).opacity =
					1 + Math.sin(time);
			}
		}
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
<!--{#if $gameState.lines != null} 
	{#each $gameState.lines as line}
		{#each line as position}
			<Piece
				piece={0}
				position={[
					position.x - 2.5,
					0.52,
					position.y - 2.5,
				]}
				placed={false}
				booped={false}
				selectable={true}
			/>
		{/each}
	{/each}
{/if}-->
{#each $gameState.board as row, y}
	{#each row as piece, x}
		{#if piece != 0}
			<Piece
				{piece}
				position={[x - 2.5, 0.52, y - 2.5]}
				placed={false}
				booped={false}
				selectable={$gameState.state === "MULTIPLE_WAITING" ? lineIndicesFor(x, y) : []}
			isMiddle={$gameState.state === "MULTIPLE_WAITING" && isMiddlePiece(x, y)}
			/>
		{/if}
	{/each}
{/each}

<!-- {#if $gameState.boopMovement != null}
	{#each $gameState.boopMovement as boopedPiece}
		<Piece
			booped={true}
			piece={boopedPiece.piece}
			position={[
				boopedPiece.position.x - 2.5,
				0.52,
				boopedPiece.position.y - 2.5,
			]}
			placed={false}
			finalPosition={[
				boopedPiece.finalPosition.x - 2.5,
				0.52,
				boopedPiece.finalPosition.y - 2.5,
			]}
		/>
	{/each}
{/if} -->
<!-- <Piece
	piece={$gameState.placed.piece}
	position={[
		$gameState.placed.position.x - 2.5,
		2,
		$gameState.placed.position.y - 2.5,
	]}
	placed={true}
	booped={false}
/> -->

{#each $graduatingLines as line (line)}
  <GraduatingLine
    positions={line.positions}
    tile={line.tile}
    onDone={() => { $graduatingLines = $graduatingLines.filter(l => l !== line); }}
  />
{/each}

{#each $boopedOffPieces as piece (piece.id)}
  <FlyingPiece
    tile={piece.tile}
    startPos={piece.startPos}
    direction={piece.direction}
    onDone={() => { $boopedOffPieces = $boopedOffPieces.filter(p => p !== piece); }}
  />
{/each}

<!-- Invisible Ground Plane -->
<T.Mesh
	rotation.x={-Math.PI / 2}
	position.y={0}
	visible={false}
	name="ground"
	oncreate={(ref) => {
		planeMesh = ref;
	}}
	onpointermove={(e: any) => {
		if (e.intersections.length > 0) {
			const { x, z } = e.intersections[0].point;
			highlightMesh.position.set(
				Math.floor(x) + 0.5,
				0.01,
				Math.floor(z) + 0.5,
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
	onpointerdown={(e: any) => {
		if (e.intersections.length > 0) {
			const { x, z } = e.intersections[0].point;
			highlightMesh.position.set(
				Math.floor(x) + 0.5,
				0.01,
				Math.floor(z) + 0.5,
			);
		}
		$lastClickPos = { x: e.nativeEvent.clientX, y: e.nativeEvent.clientY };
		console.log("pointerdown", highlightMesh.position);
		wsSendMove(highlightMesh.position);
		// @ts-ignore
		highlightMesh.material.color.setHex(0xff0000);
	}}
>
	<T.PlaneGeometry args={[6, 6]} />
	<T.MeshBasicMaterial side={THREE.DoubleSide} />
</T.Mesh>

<!-- Tile Cursor -->
<T.Mesh
	rotation.x={-Math.PI / 2}
	position.y={0}
	oncreate={(ref) => {
		highlightMesh = ref;
	}}
>
	<T.PlaneGeometry args={[1, 1]} />
	<T.MeshBasicMaterial side={THREE.DoubleSide} transparent={true} />
</T.Mesh>

<!-- Grid -->
<T.GridHelper args={[6, 6]} position.y={0.01} />
