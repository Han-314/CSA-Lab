package gol

import (
	"fmt"
	"os"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool

	ioFileName chan<- string //send-only
	output     chan<- uint8
	input      chan uint8
	keyPresses <-chan rune
}

const alive = 0xFF
const dead = 0x00

//This function takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]byte) []util.Cell {
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

func mod(x, m int) int {
	return (x + m) % m
}

func countNeighbours(p Params, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

//This function takes the current state of the world and completes one evolution of the world. It then returns the result.
func calculateNextState(p Params, world [][]byte, c distributorChannels, turn, Y0, Yt, X0, Xt int) [][]byte {
	height := Yt - Y0
	newWorld := make([][]byte, height)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := countNeighbours(p, x, y+Y0, world)
			if world[y+Y0][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
					cell := util.Cell{X: x, Y: y + Y0}
					c.events <- CellFlipped{turn, cell}
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
					cell := util.Cell{X: x, Y: y + Y0}
					c.events <- CellFlipped{turn, cell}
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

func worker(Y0, Yt, X0, Xt, turn int, world [][]byte, out chan<- [][]byte, p Params, c distributorChannels) {
	worldPart := calculateNextState(p, world, c, turn, Y0, Yt, X0, Xt)
	out <- worldPart
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	command := ioInput
	c.ioCommand <- command

	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFileName <- filename

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.input
			world[y][x] = val
		}
	}

	turn := 0

	// TODO: For all initially alive cells send a CellFlipped Event.
	flippedCell := calculateAliveCells(p, world)
	for _, v := range flippedCell {
		c.events <- CellFlipped{0, v}
	}

	//send AliveCellCount event every 2 sec.
	//c.events <- AliveCellsCount{0, 0}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			<-ticker.C
			c.events <- AliveCellsCount{turn, len(calculateAliveCells(p, world))}
		}
	}()
	//control rules' logic
	pause := false
	go func() {
		for {
			keypress := <-c.keyPresses
			switch keypress {
			case 's':
				outputCommand := ioOutput
				c.ioCommand <- outputCommand
				outputFilename := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
				c.ioFileName <- outputFilename
				for y := 0; y < p.ImageHeight; y++ {
					for x := 0; x < p.ImageWidth; x++ {
						c.output <- world[y][x]
					}
				}
				c.events <- ImageOutputComplete{turn, outputFilename}
			case 'q':
				ticker.Stop()
				outputCommand := ioOutput
				c.ioCommand <- outputCommand
				outputFilename := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
				c.ioFileName <- outputFilename
				for y := 0; y < p.ImageHeight; y++ {
					for x := 0; x < p.ImageWidth; x++ {
						c.output <- world[y][x]
					}
				}
				c.events <- ImageOutputComplete{turn, outputFilename}
				c.events <- TurnComplete{turn}
				c.ioCommand <- ioCheckIdle
				<-c.ioIdle
				c.events <- StateChange{turn, Quitting}
				os.Exit(0)

			case 'p':
				ticker.Stop()
				pause = true
				c.events <- StateChange{turn, Paused}
				fmt.Println("Current turn:", turn)
				for {
					key2 := <-c.keyPresses
					if key2 == 'p' {
						c.events <- StateChange{turn, Executing}
						fmt.Println("Continuing")
						pause = false
						ticker.Reset(2 * time.Second)
						break
					} else {
						continue
					}
				}
			default:
				fmt.Println("nothing")
			}
		}
	}()

	//logic execution.
	workerHeight := p.ImageHeight / p.Threads
	for ; turn < p.Turns; turn++ {
		out := make([]chan [][]byte, p.Threads)
		for i := range out {
			out[i] = make(chan [][]byte)
		}
	HERE:
		for pause {
			goto HERE
		}

		for i := 0; i < p.Threads-1; i++ {
			go worker(i*workerHeight, (i+1)*workerHeight, 0, p.ImageWidth, turn, world, out[i], p, c)
		}
		go worker((p.Threads-1)*workerHeight, p.ImageHeight, 0, p.ImageWidth, turn, world, out[p.Threads-1], p, c)

		nWorld := make([][]byte, 0)
		for i := range nWorld {
			nWorld[i] = make([]byte, 0)
		}

		for i := 0; i < p.Threads; i++ {
			part := <-out[i]
			nWorld = append(nWorld, part...)
		}
		world = nWorld
		c.events <- TurnComplete{turn}
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// send finalturncomplete event
	aliveCell := calculateAliveCells(p, world)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: aliveCell}
	ticker.Stop()

	// output the final board state as a PGM image
	outputCommand := ioOutput
	c.ioCommand <- outputCommand

	outputFilename := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, p.Turns)
	c.ioFileName <- outputFilename

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.output <- world[y][x]
		}
	}
	c.events <- ImageOutputComplete{turn, outputFilename}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
