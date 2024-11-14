package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/gorilla/websocket"
)

// func auth(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		// Put authentication Credentials into request Context.
// 		// Since we don't have any session backend here we simply
// 		// set user ID as empty string. Users with empty ID called
// 		// anonymous users, in real app you should decide whether
// 		// anonymous users allowed to connect to your server or not.
// 		cred := &centrifuge.Credentials{
// 			UserID: "",
// 		}
// 		newCtx := centrifuge.SetCredentials(ctx, cred)
// 		r = r.WithContext(newCtx)
// 		h.ServeHTTP(w, r)
// 	})
// }

func LoadTestGameState() *GameState {
	gameState := NewGameState()

	gameState.Board = Board{
		{1, 8, 8, 0, 8, 8},
		{0, 1, 2, 0, 1, 0},
		{2, 0, 1, 0, 0, 0},
		{2, 0, 8, 0, 2, 0},
		{0, 0, 8, 0, 8, 8},
		{8, 0, 0, 0, 0, 0},
	}

	gameState.P1.Cats = 8
	gameState.P1.Kittens = 8

	return gameState
}

type webSocketHandler struct {
	upgrader  websocket.Upgrader
	gameState *GameState
	// ch        chan Position
}

type NewMove struct {
	Position Position    `json:"position"`
	Piece    json.Number `json:"piece"`
}

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("error %s when upgrading connection to websocket", err)
		return
	}

	err = conn.WriteJSON(wsh.gameState)
	if err != nil {
		log.Println("initial message to client: writeJSON err:", err)
	}

	go func(conn *websocket.Conn) {
		for {
			newMove := &NewMove{}
			err := conn.ReadJSON(newMove)
			if err != nil {
				log.Printf("Reading New move from client: ws: Error %s when reading msg from client", err)
				conn.Close()
				return
			}
			fmt.Println("new move is: ", *newMove)
			// wsh.ch <- newMove.Position

			// kittenOrCat can either be 0 or 1 for kitten or cat respectively
			kittenOrCat, _ := newMove.Piece.Int64()
			if wsh.gameState.isPlayer1() {
				if kittenOrCat == 0 {
					err = wsh.gameState.Board.move(newMove.Position, 1, wsh.gameState)
				} else {
					err = wsh.gameState.Board.move(newMove.Position, 2, wsh.gameState)
				}
			} else {
				if kittenOrCat == 0 {
					err = wsh.gameState.Board.move(newMove.Position, 8, wsh.gameState)
				} else {
					err = wsh.gameState.Board.move(newMove.Position, 9, wsh.gameState)
				}
			}
			if err != nil {
				log.Printf("ws: Error %s", err)
				continue
			}
			log.Print(*newMove)
			log.Print(wsh.gameState.P1.Placed, wsh.gameState.P2.Placed)

			err = conn.WriteJSON(wsh.gameState)
			if err != nil {
				log.Println("writeJSON err:", err)
				break
			}

			if len(wsh.gameState.Lines) > 1 {
				wsh.gameState.Waiting = true
			} else if len(wsh.gameState.Lines) == 1 {
				wsh.gameState.Board.graduatePieces(wsh.gameState.Lines[0], wsh.gameState)
				wsh.gameState.TurnNumber++
			} else {
				if wsh.gameState.isPlayer1() && wsh.gameState.P1.Placed == 8 || !wsh.gameState.isPlayer1() && wsh.gameState.P2.Placed == 8 {
					if wsh.gameState.Board.winCheckMaxCats(wsh.gameState) {
						err = conn.WriteJSON(wsh.gameState)
						if err != nil {
							log.Println("writeJSON err:", err)
							break
						}
						conn.Close()
						return
					}
					wsh.gameState.Waiting = true
					wsh.waitForMaxedOutGraduationChoice(conn)
				}
				wsh.gameState.TurnNumber++
			}

			for wsh.gameState.Waiting {
				fmt.Println(wsh.gameState.Waiting)
				threeSelection := &NewMove{}
				err := conn.ReadJSON(threeSelection)
				if err != nil {
					log.Printf("Waiting: ws: Error %s when reading msg from client", err)
					continue
				}
				if slices.Contains(wsh.gameState.ThreeChoices, threeSelection.Position) {
					//check which of the Lines the threeSelection is in
					var line = wsh.gameState.getLineContainingPosition(threeSelection.Position)
					if line == nil {
						fmt.Println("line is nil")
						continue
					}
					wsh.gameState.Board.graduatePieces(line, wsh.gameState)
					wsh.gameState.TurnNumber++
					wsh.gameState.Waiting = false
					err = conn.WriteJSON(wsh.gameState)
					if err != nil {
						log.Println("writeJSON err:", err)
						continue
					}
				} else {
					fmt.Println("not a valid selection")
					continue
				}
			}

			err = conn.WriteJSON(wsh.gameState)
			if err != nil {
				log.Println("writeJSON err:", err)
				break
			}

		}
	}(conn)
}

func (wsh webSocketHandler) waitForMaxedOutGraduationChoice(conn *websocket.Conn) {
	for wsh.gameState.Waiting {
		fmt.Println("Waiting for maxed out grad choice", wsh.gameState.Waiting)
		pieceSelection := &NewMove{}
		err := conn.ReadJSON(pieceSelection)
		if err != nil {
			log.Printf("Max Piece Waiting: ws: Error %s when reading msg from client", err)
			conn.Close()
			return
		}
		playerPiecePosition := wsh.gameState.Board.getPlayerPiecePositions(wsh.gameState)
		fmt.Println(playerPiecePosition, pieceSelection.Position)
		if slices.Contains(playerPiecePosition, pieceSelection.Position) {
			wsh.gameState.Board.graduatePiece(pieceSelection.Position, wsh.gameState)
			wsh.gameState.Waiting = false
		} else {
			fmt.Println("Max piece waiting: not a valid selection")
			continue
		}
	}
}

type server struct {
	// subscriberMessageBuffer int
	mux http.ServeMux
	// subscribers             map[*subscriber]struct{}
}

// type subscriber struct {
// 	msgs chan []byte
// }

func NewServer() *server {
	s := &server{
		// subscriberMessageBuffer: 10,
		// subscribers:             make(map[*subscriber]struct{}),
	}

	// my websocket handler has a method that implements the http.Handler interface
	wsHandler := webSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if r.Header.Get("Origin") == "http://localhost:5173" || r.Header.Get("Origin") == "http://127.0.0.1:5173" {
					return true
				} else {
					fmt.Printf("error when upgrading connection to websocket: %s ", "Origin not allowed")
					return false
				}
			},
		},
		gameState: NewGameState(),
		// ch:        make(chan Position, 10),
	}

	s.mux.Handle("/ws", wsHandler)

	return s
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// (*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

}

func main() {

	// node, err := centrifuge.New(centrifuge.Config{})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// node.OnConnect(func(client *centrifuge.Client) {
	// 	transportName := client.Transport().Name()
	// 	// In our example clients connect with JSON protocol but it can also be Protobuf.
	// 	transportProto := client.Transport().Protocol()
	// 	log.Printf("client connected via %s (%s)", transportName, transportProto)

	// 	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
	// 		log.Printf("client subscribes on channel %s", e.Channel)
	// 		cb(centrifuge.SubscribeReply{}, nil)
	// 	})

	// 	client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
	// 		log.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
	// 		cb(centrifuge.PublishReply{}, nil)
	// 	})

	// 	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
	// 		log.Printf("client disconnected")
	// 	})
	// })

	// if err := node.Run(); err != nil {
	// 	log.Fatal(err)
	// }

	// wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{})
	// http.Handle("/connection/websocket", auth(wsHandler))

	// log.Printf("Starting server, visit http://localhost:8000")
	// if err := http.ListenAndServe(":8000", nil); err != nil {
	// 	log.Fatal(err)
	// }

	server := NewServer()

	err := http.ListenAndServe(":8080", &server.mux)

	if err != nil {
		fmt.Println("Error starting server")
		os.Exit(1)
	}
}
