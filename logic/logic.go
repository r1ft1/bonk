package main

import (
	"fmt"
)

type Position struct {
	X uint8 `json:"x"`
	Y uint8 `json:"y"`
}

type Direction struct {
	X, Y int8
}

type Booped struct {
	Direction Direction `json:"direction"`
	Position  Position  `json:"position"`
	Tile      uint8     `json:"tile"`
	BoopedBy  uint8     `json:"boopedBy"`
}

type BoopMovement struct {
	Position      Position `json:"position"`
	FinalPosition Position `json:"finalPosition"`
	Tile          uint8    `json:"tile"`
}

var directionMap = map[string]Direction{
	"topLeft":     {-1, -1},
	"above":       {0, -1},
	"topRight":    {1, -1},
	"right":       {1, 0},
	"bottomRight": {1, 1},
	"below":       {0, 1},
	"bottomLeft":  {-1, 1},
	"left":        {-1, 0},
}

type Board [6][6]uint8

type GameState struct {
	TurnNumber uint8  `json:"turnNumber"`
	PlayerTurn uint8  `json:"playerTurn"`
	Board      Board  `json:"board"`
	P1         Player `json:"p1"`
	P2         Player `json:"p2"`
	// previousBoard Board
	State  string   `json:"state"`
	Booped []Booped `json:"booped,omitempty"`
	// Waiting       bool     `json:"waiting"`
	//GraduationDecision is the outwardsmost position of a 3 in a row line
	Lines             [][]Position   `json:"lines,omitempty"`
	GraduationChoices Position       `json:"graduationChoices"`
	ThreeChoices      []Position     `json:"threeChoices"`
	Winner            uint8          `json:"winner"`
	Placed            Move           `json:"placed"`
	BoopMovement      []BoopMovement `json:"boopMovement"`
	Original          Board          `json:"original"`
	PreviousBoard     Board          `json:"previousBoard"`
}

func comparePosition(a, b Position) bool {
	return a.X == b.X && a.Y == b.Y
}

func (gameState *GameState) calculateOriginal() {
	gameState.Original = Board{
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
	}
	//Remove Placed and BoopMovement from Board and add to Original
	for y, row := range gameState.Board {
		for x, boardTile := range row {
			boardPos := Position{X: uint8(x), Y: uint8(y)}
			skip := false
			for _, boopPos := range gameState.BoopMovement {
				if comparePosition(boardPos, boopPos.FinalPosition) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			if !comparePosition(boardPos, gameState.Placed.Position) {
				// fmt.Println("Adding to Original: ", boardPos)
				gameState.Original[y][x] = boardTile
			}
		}
	}
	// fmt.Print("Original: ", gameState.Original)
}

// type Refresh struct {
// 	Placed NewMove  `json:"placed"`
// 	Booped []Booped `json:"booped"`
// }

type Player struct {
	Kittens uint8 `json:"kittens"`
	Cats    uint8 `json:"cats"`
	Placed  uint8 `json:"placed"`
}

func NewGameState() *GameState {
	gameState := new(GameState)

	gameState.TurnNumber = 0
	gameState.State = "WAITING"

	gameState.Board = Board{
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
	}
	gameState.PreviousBoard = gameState.Board
	// gameState.previousBoard = Board{
	// 	{0, 0, 0, 0, 0, 0},
	// 	{0, 0, 0, 0, 0, 0},
	// 	{0, 0, 0, 0, 0, 0},
	// 	{0, 0, 0, 0, 0, 0},
	// 	{0, 0, 0, 0, 0, 0},
	// 	{0, 0, 0, 0, 0, 0},
	// }
	gameState.P1 = Player{Kittens: 8, Cats: 0, Placed: 0}
	gameState.P2 = Player{Kittens: 8, Cats: 0, Placed: 0}

	gameState.Booped = []Booped{}

	return gameState
}

func (board *Board) move(position Position, tile uint8, gameState *GameState) error {
	if int(position.X) > len(gameState.Board)-1 || int(position.Y) > len(gameState.Board)-1 || int(position.X) < 0 || int(position.Y) < 0 {
		return fmt.Errorf("invalid position")
	}
	if tile != 1 && tile != 8 && tile != 2 && tile != 9 {
		return fmt.Errorf("invalid tile")
	}
	if (*board)[position.Y][position.X] != 0 {
		return fmt.Errorf("can't place piece. This position is already occupied")
	}
	if tile == 1 && gameState.P1.Kittens == 0 || tile == 8 && gameState.P2.Kittens == 0 || tile == 2 && gameState.P1.Cats == 0 || tile == 9 && gameState.P2.Cats == 0 {
		return fmt.Errorf("can't place piece. You have no more pieces of this type")
	}

	// gameState.previousBoard = gameState.Board

	(*board)[position.Y][position.X] = tile

	//could combine this into a next turn function
	if gameState.isPlayer1() {
		if tile == 1 {
			gameState.P1.Kittens--
		} else {
			gameState.P1.Cats--
		}
		gameState.P1.Placed++
	} else {
		if tile == 8 {
			gameState.P2.Kittens--
		} else {
			gameState.P2.Cats--
		}
		gameState.P2.Placed++
	}
	board.adjacencyCheck(position, gameState)
	// board.display(gameState)
	// gameState.TurnNumber++

	return nil
}

// func (board *Board) display(gamesState *GameState) {
// 	if gamesState.isPlayer1() {
// 		fmt.Printf("%v: Player 1 did this\n", gamesState.TurnNumber)
// 	} else {
// 		fmt.Printf("%v: Player 2 did this\n", gamesState.TurnNumber)
// 	}

// 	for i := 0; i < len(*board); i++ {
// 		fmt.Println(gamesState.previousBoard[i], (gamesState.Board)[i])
// 	}
// 	fmt.Printf("Player 1 Placed: %v, remaining Cats: %v, Kittens %v\n", gamesState.P1.Placed, gamesState.P1.Cats, gamesState.P1.Kittens)
// 	fmt.Printf("Player 2 Placed: %v, remaining Cats: %v, Kittens %v\n\n", gamesState.P2.Placed, gamesState.P2.Cats, gamesState.P2.Kittens)
// }

// func (gameState *GameState) displayBoard() {
// 	return gameState.board(gameState)
// }

func (gameState *GameState) isPlayer1() bool {
	if gameState.TurnNumber%2 == 0 {
		return true
	} else {
		return false
	}
}

func (board *Board) adjacencyCheck(newMove Position, gameState *GameState) {
	//newMove eg: {X:3, Y:1}
	//check order: left, up, right, down
	// [0 0 0 0 0 0]
	// [0 0 0 N 0 0]
	// [0 0 0 0 0 0]
	// [0 0 0 1 0 0]
	// [0 0 0 0 0 0]
	// [0 0 0 0 0 0]

	//slice of booped pieces ie reference to an array of booped pieces
	var booped []Booped

	// fmt.Printf("Checking for adjacency at position %v\n", newMove)

	for _, direction := range directionMap {
		// fmt.Printf("key[%v], value[%v]\n", directionName, direction)

		if isInBounds, contentsAtPosition := board.isDirectionInBounds(newMove, direction); isInBounds {
			//can move this if we return whether the direction is in bounds AND on an empty square
			if contentsAtPosition != 0 {
				booped = append(booped, Booped{direction, Position{newMove.X + uint8(direction.X), newMove.Y + uint8(direction.Y)}, (*board)[int8(newMove.Y)+direction.Y][int8(newMove.X)+direction.X], (*board)[int8(newMove.Y)][int8(newMove.X)]})
				// fmt.Printf("There is something to the %v\n", directionName)
			}
		}
	}

	// var newBooped = board.boopCheck(booped, gameState)
	board.boopCheck(booped, gameState)
	// gameState.Booped = newBooped
	// board.threeCheck(newMove, newBooped, gameState)
	board.checkBoardForThreeInARows(gameState)
}

// Check if middle of a 3 in a row
func (board *Board) isMiddleOfThreeInARow(position Position) []Position {
	// Check if the given position is in the middle of a 3 in a row line
	// by checking if the positions in all four directions have the same tile value

	tile := (*board)[position.Y][position.X]

	// Helper function to check if two tiles are in the same player category
	sameCategory := func(a, b uint8) bool {
		// Assuming 1 and 2 are player 1's pieces, and 8 and 9 are player 2's pieces
		return (a == 1 || a == 2) && (b == 1 || b == 2) || (a == 8 || a == 9) && (b == 8 || b == 9)
	}

	// Check left and right directions
	if position.X > 0 && position.X < 5 {
		if sameCategory((*board)[position.Y][position.X-1], tile) && sameCategory((*board)[position.Y][position.X+1], tile) {

			fmt.Println("Found 3 in a row where placed piece was in the middle at position: ", position)
			return []Position{
				{X: position.X - 1, Y: position.Y},
				{X: position.X, Y: position.Y},
				{X: position.X + 1, Y: position.Y},
			}
		}
	}

	// Check up and down directions
	if position.Y > 0 && position.Y < 5 {
		if sameCategory((*board)[position.Y-1][position.X], tile) && sameCategory((*board)[position.Y+1][position.X], tile) {

			fmt.Println("Found 3 in a row where placed piece was in the middle at position: ", position)
			return []Position{
				{X: position.X, Y: position.Y - 1},
				{X: position.X, Y: position.Y},
				{X: position.X, Y: position.Y + 1},
			}
		}
	}

	// Check top-left to bottom-right diagonal
	if position.X > 0 && position.X < 5 && position.Y > 0 && position.Y < 5 {
		if sameCategory((*board)[position.Y-1][position.X-1], tile) && sameCategory((*board)[position.Y+1][position.X+1], tile) {

			fmt.Println("Found 3 in a row where placed piece was in the middle at position: ", position)
			return []Position{
				{X: position.X - 1, Y: position.Y - 1},
				{X: position.X, Y: position.Y},
				{X: position.X + 1, Y: position.Y + 1},
			}
		}
	}

	// Check top-right to bottom-left diagonal
	if position.X > 0 && position.X < 5 && position.Y > 0 && position.Y < 5 {
		if sameCategory((*board)[position.Y-1][position.X+1], tile) && sameCategory((*board)[position.Y+1][position.X-1], tile) {
			fmt.Println("Found 3 in a row where placed piece was in the middle at position: ", position)
			return []Position{
				{X: position.X + 1, Y: position.Y - 1},
				{X: position.X, Y: position.Y},
				{X: position.X - 1, Y: position.Y + 1},
			}
		}
	}

	return nil
}

func (board *Board) checkBoardForThreeInARows(gameState *GameState) {
	// Use a map to track unique lines
	uniqueLines := make(map[string]bool)
	gameState.Lines = nil
	gameState.ThreeChoices = nil

	// Helper function to generate a unique key for a line
	generateKey := func(line []Position) string {
		key := ""
		for _, pos := range line {
			key += fmt.Sprintf("%d,%d;", pos.X, pos.Y)
		}
		return key
	}

	// Check the entire board for any 3 in a row lines
	for y := 0; y < len(*board); y++ {
		for x := 0; x < len(*board); x++ {
			position := Position{X: uint8(x), Y: uint8(y)}
			if line := board.isMiddleOfThreeInARow(position); line != nil {
				key := generateKey(line)
				if !uniqueLines[key] {
					uniqueLines[key] = true
					if player, err := board.checkLinePlayer(line); err == nil {
						// Check if the line belongs to the current player
						if (gameState.isPlayer1() && player == 1) || (!gameState.isPlayer1() && player == 2) {
							gameState.Lines = append(gameState.Lines, line)
							gameState.ThreeChoices = append(gameState.ThreeChoices, position)
							board.winCheck(line, gameState)
							fmt.Println(x, y)
						}
					}
				}
			}
		}
	}
	// fmt.Println("Lines found on the board: ", gameState.Lines, "Three choices: ", gameState.ThreeChoices)

}

func (board *Board) winCheck(line []Position, gameState *GameState) {
	//check if a player has won, if 3 Cats are in a row
	//if the 3 Cats are in a row, then the player has won
	// fmt.Println("Checking for a win")
	countCats := 0
	for _, position := range line {
		// fmt.Println("Checking position: ", position, "Contents: ", (*board)[position.Y][position.X])
		if (*board)[position.Y][position.X] == 2 || (*board)[position.Y][position.X] == 9 {
			countCats++
			fmt.Println("count cats: ", countCats)
		}
	}
	if countCats == 3 {
		fmt.Println("Player has won")
		if gameState.isPlayer1() {
			gameState.Winner = 1
		} else {
			gameState.Winner = 2
		}
		//end the game
	}
}

func (board *Board) winCheckMaxCats(gameState *GameState) bool {
	// Check if the current player has 8 cats on the board
	countCats := 0
	for y := 0; y < len(*board); y++ {
		for x := 0; x < len(*board); x++ {
			tile := (*board)[y][x]
			if gameState.isPlayer1() && tile == 2 {
				countCats++
			} else if !gameState.isPlayer1() && tile == 9 {
				countCats++
			}
		}
	}
	if countCats >= 8 {
		fmt.Println("Player has won by having 8 cats on the board")
		if gameState.isPlayer1() {
			gameState.Winner = 1
		} else {
			gameState.Winner = 2
		}
		return true
	}
	return false
}

func (board *Board) getPlayerPiecePositions(gameState *GameState) []Position {
	var positions []Position
	for y := 0; y < len(*board); y++ {
		for x := 0; x < len(*board); x++ {
			tile := (*board)[y][x]
			if gameState.isPlayer1() && (tile == 1 || tile == 2) {
				positions = append(positions, Position{X: uint8(x), Y: uint8(y)})
			} else if !gameState.isPlayer1() && (tile == 8 || tile == 9) {
				positions = append(positions, Position{X: uint8(x), Y: uint8(y)})
			}
		}
	}
	return positions
}

func (board *Board) validateLine(line []Position) bool {
	// Check if the line is valid
	// A line is valid if it contains exactly 3 positions, and all positions are within the board
	if len(line) != 3 {
		return false
	}

	for _, position := range line {
		if position.X > 5 || position.Y > 5 {
			return false
		}
	}

	// Check if the positions are adjacent
	return (isHorizontal(line) || isVertical(line) || isDiagonal(line))
}

// Helper function to check if positions are horizontally adjacent
func isHorizontal(line []Position) bool {
	return line[0].Y == line[1].Y && line[1].Y == line[2].Y &&
		((line[0].X == line[1].X-1 && line[1].X == line[2].X-1) ||
			(line[0].X == line[1].X+1 && line[1].X == line[2].X+1))
}

// Helper function to check if positions are vertically adjacent
func isVertical(line []Position) bool {
	return line[0].X == line[1].X && line[1].X == line[2].X &&
		((line[0].Y == line[1].Y-1 && line[1].Y == line[2].Y-1) ||
			(line[0].Y == line[1].Y+1 && line[1].Y == line[2].Y+1))
}

// Helper function to check if positions are diagonally adjacent
func isDiagonal(line []Position) bool {
	return ((line[0].X == line[1].X-1 && line[1].X == line[2].X-1) ||
		(line[0].X == line[1].X+1 && line[1].X == line[2].X+1)) &&
		((line[0].Y == line[1].Y-1 && line[1].Y == line[2].Y-1) ||
			(line[0].Y == line[1].Y+1 && line[1].Y == line[2].Y+1))
}

func (board *Board) checkLinePlayer(line []Position) (uint8, error) {
	// Check if all pieces in the line belong to the same player
	// Return the player number if all pieces belong to a player, otherwise return 0
	if !board.validateLine(line) {
		return 0, fmt.Errorf("invalid line")
	}

	var player uint8 = 0
	for _, position := range line {
		tile := (*board)[position.Y][position.X]

		if tile == 1 || tile == 2 {
			player = 1
		} else if tile == 8 || tile == 9 {
			player = 2
		}

		if player != 0 {
			if (player == 1 && (tile == 8 || tile == 9)) || (player == 2 && (tile == 1 || tile == 2)) {
				return 0, fmt.Errorf("line contains pieces from both players")
			}
		}
	}
	return player, nil
}

func (board *Board) graduatePieces(removedPiecePositions []Position, gameState *GameState) {
	// Remove the pieces from the board
	for _, position := range removedPiecePositions {
		(*board)[position.Y][position.X] = 0
	}

	//give player 3 Cats back
	if gameState.isPlayer1() {
		gameState.P1.Cats += 3
		gameState.P1.Placed -= 3
	} else {
		gameState.P2.Cats += 3
		gameState.P2.Placed -= 3
	}
}

func (board *Board) graduatePiece(piecePosition Position, gameState *GameState) {
	// Remove the piece from the board
	(*board)[piecePosition.Y][piecePosition.X] = 0

	// Give the player a Cat back
	if gameState.isPlayer1() {
		gameState.P1.Cats++
		gameState.P1.Placed--
	} else {
		gameState.P2.Cats++
		gameState.P2.Placed--
	}
}

func (gameState *GameState) getLineContainingPosition(position Position) []Position {
	// Return the line in which the position is in the middle of it
	fmt.Println(gameState.Lines)
	for _, line := range gameState.Lines {
		if position == line[1] {
			return line
		}
	}
	return nil
}

func (position Position) positionAtDirection(direction Direction) Position {
	return Position{position.X + uint8(direction.X), position.Y + uint8(direction.Y)}
}

// contents at position
func (board *Board) contentsAtPosition(position Position) uint8 {
	return board[position.Y][position.X]
}

// Loop through the booped array and check if the piece is boopable, if so, move off board or to new position
func (board *Board) boopCheck(booped []Booped, gameState *GameState) {
	for _, piece := range booped {

		//if the piece is a cat and the boopedBy is a kitten, then skip
		if (piece.Tile == 2 || piece.Tile == 9) && (piece.BoopedBy == 1 || piece.BoopedBy == 8) {
			continue
		}

		var isInBounds, outcomePositionContents = board.isDirectionInBounds(piece.Position, piece.Direction)
		//if the piece's direction is out of bounds - then it is boopable, add back to player's pieces
		if !isInBounds {
			fmt.Printf("The piece %v at position %v is boopable and taken off the board\n", piece.Tile, piece.Position)
			(*board)[piece.Position.Y][piece.Position.X] = 0
			if piece.Tile == 1 {
				gameState.P1.Kittens++
				gameState.P1.Placed--
			} else if piece.Tile == 2 {
				gameState.P1.Cats++
				gameState.P1.Placed--
			} else if piece.Tile == 8 {
				gameState.P2.Kittens++
				gameState.P2.Placed--
			} else {
				gameState.P2.Cats++
				gameState.P2.Placed--
			}
		}
		//if the piece's direction is in bounds and the outome square is empty - then it is boopable
		if isInBounds && outcomePositionContents == 0 {
			// fmt.Printf("The piece %v at position %v is boopable and is pushed\n", piece.Tile, piece.Position)
			(*board)[piece.Position.Y][piece.Position.X] = 0
			(*board)[int8(piece.Position.Y)+piece.Direction.Y][int8(piece.Position.X)+piece.Direction.X] = piece.Tile
			// newBooped = append(newBooped, Booped{piece.Direction, Position{piece.Position.X + uint8(piece.Direction.X), piece.Position.Y + uint8(piece.Direction.Y)}, piece.Tile, piece.BoopedBy})
			gameState.BoopMovement = nil
			gameState.BoopMovement = append(gameState.BoopMovement, BoopMovement{Position: piece.Position, FinalPosition: Position{piece.Position.X + uint8(piece.Direction.X), piece.Position.Y + uint8(piece.Direction.Y)}, Tile: piece.Tile})
		}
		//else it is not boopable - as there is a piece in the way
	}
	// return newBooped
}

// if the direction is in the bounds return true/false and what is at that position
func (board *Board) isDirectionInBounds(position Position, direction Direction) (bool, int8) {
	if (int8(position.X)+(direction.X) < 0) ||
		(int8(position.Y)+(direction.Y) < 0) ||
		(int8(position.X)+(direction.X) > int8(len(*board))-1) ||
		(int8(position.Y)+(direction.Y) > int8(len(*board))-1) {
		return false, -1
	}
	return true, int8((*board)[int8(position.Y)+direction.Y][int8(position.X)+direction.X])
}
