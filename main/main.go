package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/notnil/chess"
	"github.com/notnil/chess/image"
	"github.com/notnil/chess/uci"
)

var game *chess.Game
var eng *uci.Engine

func main() {

	// create file
	f, err := os.Create("game-image.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// create board position
	fenStr := "rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq - 0 1"
	pos := &chess.Position{}
	if err := pos.UnmarshalText([]byte(fenStr)); err != nil {
		log.Fatal(err)
	}

	// set up engine to use stockfish exe
	eng, err = uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	defer eng.Close()
	// initialize uci with new game
	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}
	// have stockfish play speed chess against itself (10 msec per move)
	game = chess.NewGame()

	fmt.Println("starting webserver...")

	r := mux.NewRouter()
	r.HandleFunc("/", makeMoves)
	http.Handle("/", r)
	r.HandleFunc("/help", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "sending help")
	})
	fmt.Println("listening on post 8080")
	http.ListenAndServe(":8080", r)
}

func makeMoves(w http.ResponseWriter, r *http.Request) {
	// player 1 -- player moves
	if len(game.Moves())%2 == 0 {
		// userQuery := r.URL.Query().Get("move")
		// game.Move(&chess.Move{
		// 	s1:    chess.Square{},
		// 	s2:    chess.Square{},
		// 	promo: chess.PieceType{},
		// 	tags:  chess.MoveTag{},
		// })
		// select a random move
		moves := game.ValidMoves()
		move := moves[rand.Intn(len(moves))]
		game.Move(move)
		time.Sleep(time.Second)
	} else { // player 2 -- stockfish moves

		cmdPos := uci.CmdPosition{Position: game.Position()}
		cmdGo := uci.CmdGo{MoveTime: time.Second / 100}
		if err := eng.Run(cmdPos, cmdGo); err != nil {
			panic(err)
		}
		move := eng.SearchResults().BestMove
		if err := game.Move(move); err != nil {
			panic(err)
		}
	}

	f, err := os.Create("game-image.svg")
	if err != nil {
		panic(err)
	}

	if err := image.SVG(f, game.Position().Board()); err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadFile("game-image.svg")
	if err != nil {
		panic(err)
	}

	_, err = w.Write(bytes)
	if err != nil {
		panic(err)
	}

}
