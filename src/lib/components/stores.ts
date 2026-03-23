import { writable } from "svelte/store";
import type { Writable } from "svelte/store";

export type GameState = {
	lines: Position[][];
	board: number[][];
	previousBoard: number[][];
	state: string;
	//original: number[][];
	turnNumber: number;
	broadcastSeq: number;
	p1: Player;
	p2: Player;
	winner: number;
	placed: NewMove;
	boopMovement: BoopMovement[];
	booped: Booped[];
};

type Position = {
	x: number;
	y: number;
};

type NewMove = {
	position: { x: number; y: number };
	piece: number;
};

type BoopMovement = {
	position: { x: number; y: number };
	finalPosition: { x: number; y: number };
	piece: number;
};

type Booped = {
	direction: { x: number; y: number };
	position: { x: number; y: number };
	tile: number;
	boopedBy: number;
};

export type ServerMessage = {
	type: string;
	gameID: string;
	playerID: string;
	state: string;
	payload: GameState | any;
};

export type Player = {
	kittens: number;
	cats: number;
	placed: number;
};

export let gameState: Writable<GameState> = writable(newGameState());
// const gs: GameState = {};
//
export function newGameState() {
	const gs: GameState = {
		board: [
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
		],
		previousBoard: [
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
			[0, 0, 0, 0, 0, 0],
		],
		state: "WAITING",
		//original: [
		//	[0, 0, 0, 0, 0, 0],
		//	[0, 0, 0, 0, 0, 0],
		//	[0, 0, 0, 0, 0, 0],
		//	[0, 0, 0, 0, 0, 0],
		//	[0, 0, 0, 0, 0, 0],
		//	[0, 0, 0, 0, 0, 0],
		//],
		turnNumber: 0,
		p1: {
			kittens: 8,
			cats: 0,
			placed: 0,
		},
		p2: {
			kittens: 8,
			cats: 0,
			placed: 0,
		},
		winner: 0,
		placed: {
			position: { x: 0, y: 0 },
			piece: 0,
		},
		boopMovement: [
			{
				position: { x: 0, y: 0 },
				finalPosition: { x: 0, y: 0 },
				piece: 0,
			},
		],
		lines: [[]],
	}
	return gs;
}

export let webSocket: Writable<WebSocket> = writable();
export let p1WebSocket: Writable<WebSocket> = writable();
export let p2WebSocket: Writable<WebSocket> = writable();
export let message = writable({ type: "", gameID: "", playerID: "", state: "", payload: {} } as ServerMessage);
export let waitingForOpponent = writable(false);
export let onlineGameID = writable("");
export let pieceChoice = writable(0);
export let inGame = writable(false);
export let waitingGameIDs = writable("");
export let lastClickPos = writable({ x: 0, y: 0 });
export let noPiecesMsg = writable("");

export type GraduatingLineData = {
	positions: [number, number, number][]; // 3 world-space [x, y, z]
	tile: number; // 1 (P1 kitten) or 8 (P2 kitten)
};
export let graduatingLines = writable<GraduatingLineData[]>([]);

export type BoopedOffData = {
	id: number;
	startPos: [number, number, number];
	tile: number; // 1, 2, 8, or 9
	direction: [number, number]; // boop direction in board coords (dx, dy)
};
export let boopedOffPieces = writable<BoopedOffData[]>([]);

export type SlidingPieceData = {
	id: number;
	startPos: [number, number, number];
	endPos: [number, number, number];
	tile: number;
};
export let slidingPieces = writable<SlidingPieceData[]>([]);

export let isMobile = writable(false);

// Signals when the placement arc animation has landed
export let placementLanded = writable(true);

// Set only on real placements (not graduation selections) to trigger arc animation
export let arcTrigger = writable({ x: 0, y: 0, piece: 0, turn: -1 });

// Animation tuning parameters — controlled via debug overlay
export const animConfig = writable({
	// Placement arc
	arcDuration: 0.6,
	arcHeight: 5,
	arcLandThreshold: 0.85, // trigger boops at this % of arc

	// Timing between animations (negative = overlap, positive = gap)
	slideDelay: -0.05,
	flyDelay: -0.4,

	// Sliding piece (on-board boop)
	slideDuration: 0.25,
	slideArcHeight: 0.2,

	// Flying piece (off-board boop)
	bumpDuration: 0.15,
	bumpDistance: 1.5,
	bumpArcHeight: 0.3,
	gravity: 18,
	bumpVelocityY: 3,
	bumpVelocityXZ: 4,
	groundY: -2.1,
	bounceEnergyLoss: 0.4,
	bounceFriction: 0.5,
	bounceMinVelocity: 1,
	tumbleSpeed: 6,
	shrinkDuration: 0.4,
});
