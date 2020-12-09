package stubs

import (
	//"uk.ac.bris.cs/gameoflife/Game-of-Life/gol"
	"uk.ac.bris.cs/gameoflife/gol"
)

var NewBoard = "Engine.NewBaord"

// var CreateChannel = "Engine.CreatChannel"
// var Publish = "Engine.Publish"
// var Subscribe = "Engine.Subscribe"

//PGM image
type Board struct {
	World [][]byte
	Turn  int
	P     gol.Params
}
type BoardResponse struct {
	NewWorld [][]byte
	NewTurn  int
}

// type PublishRequest struct {
// 	Topic string
// 	Board Board
// }

// type ChannelRequest struct {
// 	Topic string
// }

// type BoardPart struct {
// 	Y0, Yt, X0, Xt int
// }

//Response results
// type Subscription struct {
// 	Topic         string
// 	WorkerAddress string
// 	Callback      string
// }

// type StatusReport struct {
// 	Message string
// }
