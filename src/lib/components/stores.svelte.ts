import { writable } from "svelte/store";
import type { WebSocketClient } from "vite";

export type GameState = {
	board: number[][];
	turnNumber: number;
	p1: Player;
	p2: Player;
	winner: number;
	placed: NewMove;
};

type NewMove = {
	position: { x: number; y: number };
	piece: number;
};

export type ServerMessage = {
	type:     string      
	gameID:   string    
	payload:  GameState | any
}



type Player = {
	kittens: number;
	cats: number;
	placed: number;
};

// const gs: GameState = {};

export let gameState = writable({
	board: [
		[0, 0, 0, 0, 0, 0],
		[0, 0, 0, 0, 0, 0],
		[0, 0, 0, 0, 0, 0],
		[0, 0, 0, 0, 0, 0],
		[0, 0, 0, 0, 0, 0],
		[0, 0, 0, 0, 0, 0],
	],
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
});


export let webSocket = writable(new WebSocket("ws://localhost:8080/ws"));
export let message = writable({type: "", gameID: "", payload: {}});
export let pieceChoice = writable(0);