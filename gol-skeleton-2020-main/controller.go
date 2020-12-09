package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func readImage(filename string, p gol.Params) [][]byte {
	data, ioError := ioutil.ReadFile("images/" + filename + ".pgm")
	util.Check(ioError)

	fields := strings.Fields(string(data))

	width, _ := strconv.Atoi(fields[1])
	if width != p.ImageWidth {
		panic("Incorrect width")
	}

	height, _ := strconv.Atoi(fields[2])
	if height != p.ImageHeight {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	input := make(chan byte)
	for _, b := range image {
		input <- b
	}

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-input
			world[y][x] = val
		}
	}
	return world
}

const alive = 0xFF
const dead = 0x00

func calculateAliveCells(p gol.Params, world [][]byte) []util.Cell {
	aliveCell := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCell = append(aliveCell, util.Cell{X: x, Y: y})
			}
		}
	}
	return aliveCell
}

func makeCall(client rpc.Client, p gol.Params, events chan gol.Event, world [][]byte) {
	request := stubs.Board{P: p, World: world}
	response := new(stubs.BoardResponse)
	client.Call(stubs.NewBoard, request, response)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		<-ticker.C
		events <- gol.AliveCellsCount{response.NewTurn, len(calculateAliveCells(p, response.NewWorld))}
	}()
}

func main() {
	//stage 2-main
	engineAddr := flag.String("engine", "127.0.0.1:8030", "Address of enging")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *engineAddr)
	defer client.Close()

	params := gol.Params{
		Turns:       100000000,
		Threads:     8,
		ImageWidth:  512,
		ImageHeight: 512,
	}
	filename := fmt.Sprintf("%dx%d", params.ImageHeight, params.ImageWidth)
	newWorld := readImage(filename, params)

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	gol.Run(params, events, keyPresses)
	sdl.Start(params, events, keyPresses)
	makeCall(*client, params, events, newWorld)

	// if err != nil {
	// 	fmt.Println("RPC client returned error:")
	// 	fmt.Println(err)
	// 	fmt.Println("Shutting down controller.")
	// 	panic(err)
	// }
}
