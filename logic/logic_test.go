package main

import "testing"

// --- Helpers ---

func newP1Turn() *GameState {
	return NewGameState() // TurnNumber=0 = P1's turn
}

func newP2Turn() *GameState {
	gs := NewGameState()
	gs.TurnNumber = 1
	return gs
}

// Place pieces directly on the board and adjust player state accordingly.
func place(gs *GameState, tile uint8, x, y uint8) {
	gs.Board[y][x] = tile
	if isP1Piece(tile) {
		if isKitten(tile) {
			gs.P1.Kittens--
		} else {
			gs.P1.Cats--
		}
		gs.P1.Placed++
	} else {
		if isKitten(tile) {
			gs.P2.Kittens--
		} else {
			gs.P2.Cats--
		}
		gs.P2.Placed++
	}
}

// --- Tile helpers ---

func TestTileHelpers(t *testing.T) {
	if !isP1Piece(P1Kitten) || !isP1Piece(P1Cat) {
		t.Error("isP1Piece: should be true for P1 tiles")
	}
	if isP1Piece(P2Kitten) || isP1Piece(P2Cat) || isP1Piece(0) {
		t.Error("isP1Piece: should be false for non-P1 tiles")
	}
	if !isP2Piece(P2Kitten) || !isP2Piece(P2Cat) {
		t.Error("isP2Piece: should be true for P2 tiles")
	}
	if isP2Piece(P1Kitten) || isP2Piece(0) {
		t.Error("isP2Piece: should be false for non-P2 tiles")
	}
	if !isCat(P1Cat) || !isCat(P2Cat) {
		t.Error("isCat: should be true for cats")
	}
	if isCat(P1Kitten) || isCat(P2Kitten) {
		t.Error("isCat: should be false for kittens")
	}
	if !isKitten(P1Kitten) || !isKitten(P2Kitten) {
		t.Error("isKitten: should be true for kittens")
	}
	if isKitten(P1Cat) || isKitten(P2Cat) {
		t.Error("isKitten: should be false for cats")
	}
	if tileOwner(P1Kitten) != 1 || tileOwner(P1Cat) != 1 {
		t.Error("tileOwner: should return 1 for P1 pieces")
	}
	if tileOwner(P2Kitten) != 2 || tileOwner(P2Cat) != 2 {
		t.Error("tileOwner: should return 2 for P2 pieces")
	}
	if tileOwner(0) != 0 {
		t.Error("tileOwner: should return 0 for empty square")
	}
}

// --- Placement validation ---

func TestPlacement_Valid(t *testing.T) {
	gs := newP1Turn()
	if err := gs.Board.move(Position{X: 3, Y: 3}, P1Kitten, gs); err != nil {
		t.Errorf("expected valid placement, got: %v", err)
	}
}

func TestPlacement_OccupiedSquare(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 3, 3)
	err := gs.Board.move(Position{X: 3, Y: 3}, P1Kitten, gs)
	if err == nil {
		t.Error("expected error placing on occupied square")
	}
}

func TestPlacement_OutOfBounds(t *testing.T) {
	gs := newP1Turn()
	err := gs.Board.move(Position{X: 6, Y: 0}, P1Kitten, gs)
	if err == nil {
		t.Error("expected error placing out of bounds")
	}
}

func TestPlacement_NoPiecesLeft(t *testing.T) {
	gs := newP1Turn()
	gs.P1.Kittens = 0
	err := gs.Board.move(Position{X: 0, Y: 0}, P1Kitten, gs)
	if err == nil {
		t.Error("expected error when no kittens left")
	}
}

// --- Boop mechanics ---

// A piece adjacent to the newly placed piece is pushed one square away.
func TestBoop_PushesAdjacentPiece(t *testing.T) {
	gs := newP1Turn()
	place(gs, P2Kitten, 3, 3)

	if err := gs.Board.move(Position{X: 2, Y: 3}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][3] != 0 {
		t.Error("expected original position (3,3) to be empty after boop")
	}
	if gs.Board[3][4] != P2Kitten {
		t.Errorf("expected P2Kitten at (4,3) after boop, got %d", gs.Board[3][4])
	}
}

// A piece booped off the edge returns to its owner's hand.
func TestBoop_PieceOffEdgeReturnsToHand(t *testing.T) {
	gs := newP1Turn()
	place(gs, P2Kitten, 5, 3) // rightmost column

	if err := gs.Board.move(Position{X: 4, Y: 3}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][5] != 0 {
		t.Error("expected (5,3) to be empty after piece booped off board")
	}
	if gs.P2.Kittens != 8 {
		t.Errorf("expected P2.Kittens=8 after returning, got %d", gs.P2.Kittens)
	}
	if gs.P2.Placed != 0 {
		t.Errorf("expected P2.Placed=0 after returning, got %d", gs.P2.Placed)
	}
}

// A cat booped off the edge by another cat returns as a cat, not a kitten.
func TestBoop_CatOffEdgeReturnsToCatPool(t *testing.T) {
	gs := newP1Turn()
	gs.P1.Cats = 1
	gs.P1.Kittens = 8
	gs.P2.Cats = 0
	gs.P2.Kittens = 8
	gs.P2.Placed = 1
	gs.Board[3][5] = P2Cat

	// P1 Cat at (4,3) boops P2 Cat off the right edge (kittens can't boop cats)
	if err := gs.Board.move(Position{X: 4, Y: 3}, P1Cat, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.P2.Cats != 1 {
		t.Errorf("expected P2.Cats=1 after cat returned from edge, got %d", gs.P2.Cats)
	}
	if gs.P2.Kittens != 8 {
		t.Errorf("expected P2.Kittens to be unchanged at 8, got %d", gs.P2.Kittens)
	}
}

// A piece with another piece behind it cannot be booped.
func TestBoop_BlockedByPiece(t *testing.T) {
	gs := newP1Turn()
	place(gs, P2Kitten, 2, 3)
	place(gs, P1Kitten, 3, 3) // blocks the boop

	if err := gs.Board.move(Position{X: 1, Y: 3}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][2] != P2Kitten {
		t.Errorf("expected P2Kitten to stay at (2,3) when blocked, got %d", gs.Board[3][2])
	}
}

// Kittens cannot boop cats — cats stay put when bumped by a kitten.
func TestBoop_KittenCannotBoopCat(t *testing.T) {
	gs := newP1Turn()
	gs.P2.Cats = 0
	gs.P2.Kittens = 8
	gs.P2.Placed = 1
	gs.Board[3][3] = P2Cat

	if err := gs.Board.move(Position{X: 2, Y: 3}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][3] != P2Cat {
		t.Error("expected P2Cat to remain at (3,3) — kittens cannot boop cats")
	}
}

// Cats can boop other cats.
func TestBoop_CatBoopsCat(t *testing.T) {
	gs := newP1Turn()
	gs.P1.Cats = 1
	gs.P1.Kittens = 8
	gs.P2.Cats = 0
	gs.P2.Kittens = 8
	gs.P2.Placed = 1
	gs.Board[3][3] = P2Cat

	if err := gs.Board.move(Position{X: 2, Y: 3}, P1Cat, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][3] != 0 {
		t.Error("expected (3,3) empty after cat boops cat")
	}
	if gs.Board[3][4] != P2Cat {
		t.Errorf("expected P2Cat at (4,3) after boop, got %d", gs.Board[3][4])
	}
}

// Cats can boop kittens.
func TestBoop_CatBoopsKitten(t *testing.T) {
	gs := newP1Turn()
	gs.P1.Cats = 1
	gs.P1.Kittens = 8
	place(gs, P2Kitten, 3, 3)

	if err := gs.Board.move(Position{X: 2, Y: 3}, P1Cat, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Board[3][4] != P2Kitten {
		t.Errorf("expected P2Kitten at (4,3) after cat boop, got %d", gs.Board[3][4])
	}
}

// --- Three-in-a-row detection ---

// When placing completes a horizontal line, Lines is populated and graduation is pending.
func TestThreeInARow_Horizontal(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 0, 0)
	place(gs, P1Kitten, 1, 0)
	// Placing at (2,0): boop on (1,0) is blocked by (0,0), line forms.
	if err := gs.Board.move(Position{X: 2, Y: 0}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gs.Lines) != 1 {
		t.Errorf("expected 1 horizontal line detected, got %d", len(gs.Lines))
	}
}

func TestThreeInARow_Vertical(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 2, 0)
	place(gs, P1Kitten, 2, 1)
	// Placing at (2,2): boop on (2,1) is blocked by (2,0), line forms.
	if err := gs.Board.move(Position{X: 2, Y: 2}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gs.Lines) != 1 {
		t.Errorf("expected 1 vertical line detected, got %d", len(gs.Lines))
	}
}

func TestThreeInARow_Diagonal(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 0, 0)
	place(gs, P1Kitten, 1, 1)
	// Placing at (2,2): boop on (1,1) is blocked by (0,0), diagonal forms.
	if err := gs.Board.move(Position{X: 2, Y: 2}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gs.Lines) != 1 {
		t.Errorf("expected 1 diagonal line detected, got %d", len(gs.Lines))
	}
}

func TestThreeInARow_AntiDiagonal(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 2, 0) // top-right
	place(gs, P1Kitten, 1, 1) // middle
	// Placing at (0,2): boop on (1,1) is blocked by (2,0), anti-diagonal forms.
	if err := gs.Board.move(Position{X: 0, Y: 2}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gs.Lines) != 1 {
		t.Errorf("expected 1 anti-diagonal line detected, got %d", len(gs.Lines))
	}
}

// Only the current player's lines are detected — opponent 3-in-a-rows are ignored.
func TestThreeInARow_OnlyCurrentPlayerDetected(t *testing.T) {
	gs := newP1Turn() // P1's turn
	// Set up a complete P2 line directly on the board
	gs.Board[0][0] = P2Kitten
	gs.Board[0][1] = P2Kitten
	gs.Board[0][2] = P2Kitten

	gs.Board.checkBoardForThreeInARows(gs)

	if len(gs.Lines) != 0 {
		t.Error("expected P2's 3-in-a-row to not be detected on P1's turn")
	}
}

// Two separate 3-in-a-rows at once results in multiple Lines entries.
func TestThreeInARow_MultipleLines(t *testing.T) {
	gs := newP1Turn()
	// Two complete horizontal lines, far apart so no boop interaction
	gs.Board[0][0] = P1Kitten
	gs.Board[0][1] = P1Kitten
	gs.Board[0][2] = P1Kitten
	gs.Board[3][0] = P1Kitten
	gs.Board[3][1] = P1Kitten
	gs.Board[3][2] = P1Kitten

	gs.Board.checkBoardForThreeInARows(gs)

	if len(gs.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(gs.Lines))
	}
}

// Mixed line (kittens + cats belonging to same player) is still detected.
func TestThreeInARow_MixedKittenAndCat(t *testing.T) {
	gs := newP1Turn()
	gs.Board[0][0] = P1Kitten
	gs.Board[0][1] = P1Cat
	gs.Board[0][2] = P1Kitten

	gs.Board.checkBoardForThreeInARows(gs)

	if len(gs.Lines) != 1 {
		t.Errorf("expected mixed line to be detected, got %d lines", len(gs.Lines))
	}
}

// --- Win conditions ---

// Three cats in a row wins the game immediately.
func TestWin_ThreeCatsInARow(t *testing.T) {
	gs := newP1Turn()
	gs.P1.Cats = 1
	gs.P1.Placed = 2
	gs.Board[0][0] = P1Cat
	gs.Board[0][1] = P1Cat

	if err := gs.Board.move(Position{X: 2, Y: 0}, P1Cat, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Winner != 1 {
		t.Errorf("expected P1 to win with 3 cats in a row, got Winner=%d", gs.Winner)
	}
}

// 3 kittens in a row does NOT win — only cats win.
func TestNoWin_ThreeKittensInARow(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 0, 0)
	place(gs, P1Kitten, 1, 0)

	if err := gs.Board.move(Position{X: 2, Y: 0}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gs.Winner != 0 {
		t.Errorf("expected no winner for 3 kittens, got Winner=%d", gs.Winner)
	}
	if len(gs.Lines) != 1 {
		t.Error("expected graduation to be pending (Lines populated)")
	}
}

// Getting 8 cats on the board wins the game.
func TestWin_EightCatsOnBoard(t *testing.T) {
	gs := newP1Turn()
	positions := []Position{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 1}, {1, 1}, {2, 1}, {3, 1}}
	for _, p := range positions {
		gs.Board[p.Y][p.X] = P1Cat
	}
	gs.P1.Cats = 0
	gs.P1.Placed = 8

	won := gs.Board.winCheckMaxCats(gs)

	if !won {
		t.Error("expected win with 8 cats on board")
	}
	if gs.Winner != 1 {
		t.Errorf("expected Winner=1, got %d", gs.Winner)
	}
}

// 7 cats is not enough to win.
func TestNoWin_SevenCatsOnBoard(t *testing.T) {
	gs := newP1Turn()
	positions := []Position{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 1}, {1, 1}, {2, 1}}
	for _, p := range positions {
		gs.Board[p.Y][p.X] = P1Cat
	}
	gs.P1.Cats = 0
	gs.P1.Placed = 7

	won := gs.Board.winCheckMaxCats(gs)

	if won {
		t.Error("expected no win with only 7 cats on board")
	}
	if gs.Winner != 0 {
		t.Errorf("expected Winner=0, got %d", gs.Winner)
	}
}

// --- Graduation ---

// Three kittens in a row are removed from the board and the player receives 3 cats.
func TestGraduation_ThreeKittensBecomeCats(t *testing.T) {
	gs := newP1Turn()
	place(gs, P1Kitten, 0, 0)
	place(gs, P1Kitten, 1, 0)

	if err := gs.Board.move(Position{X: 2, Y: 0}, P1Kitten, gs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gs.Lines) != 1 {
		t.Fatalf("expected 1 line before graduation, got %d", len(gs.Lines))
	}

	catsBefore := gs.P1.Cats
	placedBefore := gs.P1.Placed
	gs.Board.graduatePieces(gs.Lines[0], gs)

	if gs.Board[0][0] != 0 || gs.Board[0][1] != 0 || gs.Board[0][2] != 0 {
		t.Error("expected all three board positions to be cleared after graduation")
	}
	if gs.P1.Cats != catsBefore+3 {
		t.Errorf("expected P1.Cats=%d, got %d", catsBefore+3, gs.P1.Cats)
	}
	if gs.P1.Placed != placedBefore-3 {
		t.Errorf("expected P1.Placed=%d, got %d", placedBefore-3, gs.P1.Placed)
	}
}

// graduatePiece (single) removes one piece and gives one cat.
func TestGraduation_SinglePiece(t *testing.T) {
	gs := newP1Turn()
	gs.Board[2][2] = P1Kitten
	gs.P1.Kittens = 7
	gs.P1.Placed = 1

	gs.Board.graduatePiece(Position{X: 2, Y: 2}, gs)

	if gs.Board[2][2] != 0 {
		t.Error("expected position to be cleared after single graduation")
	}
	if gs.P1.Cats != 1 {
		t.Errorf("expected P1.Cats=1, got %d", gs.P1.Cats)
	}
	if gs.P1.Placed != 0 {
		t.Errorf("expected P1.Placed=0, got %d", gs.P1.Placed)
	}
}

// --- restorePiece ---

func TestRestorePiece_Kitten(t *testing.T) {
	gs := NewGameState()
	gs.P1.Kittens = 6
	gs.P1.Placed = 2

	gs.restorePiece(P1Kitten)

	if gs.P1.Kittens != 7 {
		t.Errorf("expected P1.Kittens=7, got %d", gs.P1.Kittens)
	}
	if gs.P1.Placed != 1 {
		t.Errorf("expected P1.Placed=1, got %d", gs.P1.Placed)
	}
}

func TestRestorePiece_Cat(t *testing.T) {
	gs := NewGameState()
	gs.P1.Cats = 0
	gs.P1.Placed = 1

	gs.restorePiece(P1Cat)

	if gs.P1.Cats != 1 {
		t.Errorf("expected P1.Cats=1, got %d", gs.P1.Cats)
	}
	if gs.P1.Placed != 0 {
		t.Errorf("expected P1.Placed=0, got %d", gs.P1.Placed)
	}
}

func TestRestorePiece_P2Kitten(t *testing.T) {
	gs := NewGameState()
	gs.P2.Kittens = 5
	gs.P2.Placed = 3

	gs.restorePiece(P2Kitten)

	if gs.P2.Kittens != 6 {
		t.Errorf("expected P2.Kittens=6, got %d", gs.P2.Kittens)
	}
	if gs.P2.Placed != 2 {
		t.Errorf("expected P2.Placed=2, got %d", gs.P2.Placed)
	}
}
