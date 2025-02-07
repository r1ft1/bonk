import { writable } from "svelte/store";
import type { Writable } from "svelte/store";

export type GameState = {
	lines: Position[][];
	board: number[][];
	previousBoard: number[][];
	state: string;
	//original: number[][];
	turnNumber: number;
	p1: Player;
	p2: Player;
	winner: number;
	placed: NewMove;
	boopMovement: BoopMovement[];
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

export type ServerMessage = {
	type: string;
	gameID: string;
	state: string;
	payload: GameState | any;
};

type Player = {
	kittens: number;
	cats: number;
	placed: number;
};

export let gameState: Writable<GameState> = writable(newGameState());
// const gs: GameState = {};
//
function newGameState() {
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
export let message = writable({ type: "", gameID: "", state: "", payload: {} } as ServerMessage);
export let pieceChoice = writable(0);
export let inGame = writable(false);
export let waitingGameIDs = writable("");
