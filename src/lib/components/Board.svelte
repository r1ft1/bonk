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
		slidingPieces,
		placementLanded,
		arcTrigger,
		isMobile as isMobileStore,
	} from "./stores";
	import { animate } from "motion";
	import Piece from "./Piece.svelte";
	import GraduatingLine from "./GraduatingLine.svelte";
	import FlyingPiece from "./FlyingPiece.svelte";
	import SlidingPiece from "./SlidingPiece.svelte";
	import CardboardBox from "./CardboardBox.svelte";

	const map = useTexture("/tile.png", {
		transform: (texture) => {
			texture.wrapS = THREE.RepeatWrapping;
			texture.wrapT = THREE.RepeatWrapping;
			texture.repeat.set(4, 4);
			return texture;
		},
	});

	let color = $derived($gameState.turnNumber % 2 === 0 ? "orange" : "lightblue");

	// Auto-switch piece selection if current choice is unavailable
	$effect(() => {
		if ($gameState.state !== "WAITING") return;
		const isP1 = $gameState.turnNumber % 2 === 0;
		const player = isP1 ? $gameState.p1 : $gameState.p2;
		const available = $pieceChoice === 0 ? player.kittens : player.cats;
		if (available <= 0) {
			$pieceChoice = $pieceChoice === 0 ? 1 : 0;
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
	};

	interactivity();

	//const { camera, scene, renderMode, autoRender } = useThrelte();

	let innerWidth = $state(window.innerWidth);
	let innerHeight = $state(window.innerHeight);
	let isMobile = $derived(innerWidth < 768);
	$effect(() => { $isMobileStore = isMobile; }); // sync to store for other components

	let planeMesh: THREE.Mesh;
	let highlightMesh: THREE.Mesh;

	let lastMove = $derived({ x: $gameState.placed.position.x, z: $gameState.placed.position.y });

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

<svelte:window bind:innerWidth={innerWidth} bind:innerHeight={innerHeight} />

<!-- Bed -->
<!-- Mattress (the game board surface) -->
{#await map then value}
	<T.Mesh position.y={-0.35}>
		<T.BoxGeometry args={[6.7, 0.7, 6.7]} oncreate={(geo) => {
			// Fix UVs on side faces so texture is cropped, not stretched
			const uv = geo.attributes.uv;
			const ratio = 0.7 / 6.7; // height / width of side face
			// BoxGeometry face order: +x, -x, +y, -y, +z, -z (4 verts each)
			// Faces 0-1 (+x, -x) and 4-5 (+z, -z) are the sides
			for (let face = 0; face < 6; face++) {
				if (face === 2 || face === 3) continue; // skip top/bottom
				const offset = face * 4;
				for (let v = 0; v < 4; v++) {
					const idx = offset + v;
					const vVal = uv.getY(idx);
					// Remap V from [0,1] to cropped range centered
					uv.setY(idx, 0.5 + (vVal - 0.5) * ratio);
				}
			}
			uv.needsUpdate = true;
		}} />
		<T.MeshBasicMaterial map={value} />
		<Outlines color="black" thickness={0.02} />
	</T.Mesh>
{/await}
<!-- Mattress edge lines -->
<T.Mesh position.y={-0.35}>
	<T.BoxGeometry args={[6.72, 0.72, 6.72]} />
	<T.MeshStandardMaterial color="#f5efe6" transparent opacity={0.0} />
	<Edges color="#c4b8a8" />
</T.Mesh>

<!-- Bed frame (wooden box around mattress) -->
<!-- Left rail -->
<T.Mesh position={[-3.35, -0.95, 0]}>
	<T.BoxGeometry args={[0.3, 0.5, 6.8]} />
	<T.MeshStandardMaterial color="#8B6F4E" />
	<Outlines color="black" thickness={0.015} />
	<Edges color="#5a4232" />
</T.Mesh>
<!-- Right rail -->
<T.Mesh position={[3.35, -0.95, 0]}>
	<T.BoxGeometry args={[0.3, 0.5, 6.8]} />
	<T.MeshStandardMaterial color="#8B6F4E" />
	<Outlines color="black" thickness={0.015} />
	<Edges color="#5a4232" />
</T.Mesh>
<!-- Front rail -->
<T.Mesh position={[0, -0.95, 3.35]}>
	<T.BoxGeometry args={[6.4, 0.5, 0.3]} />
	<T.MeshStandardMaterial color="#8B6F4E" />
	<Outlines color="black" thickness={0.015} />
	<Edges color="#5a4232" />
</T.Mesh>
<!-- Back rail (shorter, headboard goes above) -->
<T.Mesh position={[0, -0.95, -3.35]}>
	<T.BoxGeometry args={[6.4, 0.5, 0.3]} />
	<T.MeshStandardMaterial color="#8B6F4E" />
	<Outlines color="black" thickness={0.015} />
	<Edges color="#5a4232" />
</T.Mesh>


<!-- Legs -->
<!-- Back left -->
<T.Mesh position={[-3.25, -1.75, -3.25]}>
	<T.CylinderGeometry args={[0.18, 0.15, 2.2, 8]} />
	<T.MeshStandardMaterial color="#6B5337" />
	<Outlines color="black" thickness={0.015} />
</T.Mesh>
<!-- Back right -->
<T.Mesh position={[3.25, -1.75, -3.25]}>
	<T.CylinderGeometry args={[0.18, 0.15, 2.2, 8]} />
	<T.MeshStandardMaterial color="#6B5337" />
	<Outlines color="black" thickness={0.015} />
</T.Mesh>
<!-- Front left -->
<T.Mesh position={[-3.25, -1.75, 3.25]}>
	<T.CylinderGeometry args={[0.18, 0.15, 1.5, 8]} />
	<T.MeshStandardMaterial color="#6B5337" />
	<Outlines color="black" thickness={0.015} />
</T.Mesh>
<!-- Front right -->
<T.Mesh position={[3.25, -1.75, 3.25]}>
	<T.CylinderGeometry args={[0.18, 0.15, 1.5, 8]} />
	<T.MeshStandardMaterial color="#6B5337" />
	<Outlines color="black" thickness={0.015} />
</T.Mesh>


<!-- Cardboard boxes (hidden on mobile — pieces shown in overlay) -->
{#if !isMobile}
	<CardboardBox position={[-4.5, -1.5, -1]} rotation={0.15} scale={1.3} kittens={$gameState.p1.kittens} cats={$gameState.p1.cats} color="orange" />
	<CardboardBox position={[0, -1.5, 4.8]} rotation={0.7} scale={1.3} kittens={$gameState.p2.kittens} cats={$gameState.p2.cats} color="lightblue" />
{/if}

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
		{@const isSliding = $slidingPieces.some(s =>
			Math.abs(s.endPos[0] - (x - 2.5)) < 0.01 && Math.abs(s.endPos[2] - (y - 2.5)) < 0.01
		)}
		{@const isArcingHere = !$placementLanded && $arcTrigger.piece !== 0 && $arcTrigger.x === x && $arcTrigger.y === y}
		{@const displayPiece = piece !== 0 ? piece : (isArcingHere ? $arcTrigger.piece : 0)}
		{#if displayPiece != 0 && !isSliding}
			<Piece
				piece={displayPiece}
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

{#each $slidingPieces as piece (piece.id)}
  <SlidingPiece
    tile={piece.tile}
    startPos={piece.startPos}
    endPos={piece.endPos}
    onDone={() => { $slidingPieces = $slidingPieces.filter(p => p !== piece); }}
  />
{/each}

{#each $boopedOffPieces as piece (piece.id)}
  <FlyingPiece
    tile={piece.tile}
    startPos={piece.startPos}
    direction={piece.direction}
    delay={piece.delay}
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
		if (isMobile) return;
		if (e.intersections.length > 0) {
			const { x, z } = e.intersections[0].point;
			highlightMesh.position.set(
				Math.floor(x) + 0.5,
				0.01,
				Math.floor(z) + 0.5,
			);
		}
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
		wsSendMove(highlightMesh.position);
		// @ts-ignore
		highlightMesh.material.color.setHex(0xff0000);
	}}
>
	<T.PlaneGeometry args={[6, 6]} />
	<T.MeshBasicMaterial side={THREE.DoubleSide} />
</T.Mesh>

<!-- Tile Cursor (outline only, hidden on mobile) -->
<T.Mesh
	position.y={0.03}
	visible={!isMobile}
	oncreate={(ref) => {
		highlightMesh = ref;
	}}
>
	<T.BoxGeometry args={[1, 0.01, 1]} />
	<T.MeshBasicMaterial transparent={true} opacity={0} />
	<Edges color="#5a4a3a" />
</T.Mesh>

<!-- Grid -->
<!-- Grid lines (thick, visible) -->
{#each [-3, -2, -1, 0, 1, 2, 3] as x}
	<T.Mesh position={[x, 0.02, 0]}>
		<T.BoxGeometry args={[0.04, 0.01, 6]} />
		<T.MeshBasicMaterial color="#5a4a3a" transparent opacity={0.45} />
	</T.Mesh>
{/each}
{#each [-3, -2, -1, 0, 1, 2, 3] as z}
	<T.Mesh position={[0, 0.02, z]}>
		<T.BoxGeometry args={[6, 0.01, 0.04]} />
		<T.MeshBasicMaterial color="#5a4a3a" transparent opacity={0.45} />
	</T.Mesh>
{/each}
