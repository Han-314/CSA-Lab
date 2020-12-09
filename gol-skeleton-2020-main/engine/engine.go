package main

import (
	"errors"
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
)

/*var (
	topics  = make(map[string]chan stubs.Board)
	topicmx sync.RWMutex
)*/

//Create a new topic as a channel.
/*func createTopic(topic string) {
	topicmx.Lock()
	defer topicmx.Unlock()
	if _, ok := topics[topic]; !ok {
		topics[topic] = make(chan stubs.Board)
		fmt.Println("Created channel #", topic)
	}
}*/

//The Board is published to the topic.
/*func publish(topic string, b stubs.Board) (err error) {
	topicmx.RLock()
	defer topicmx.RUnlock()
	if ch, ok := topics[topic]; ok {
		ch <- b
	} else {
		return errors.New("No such topic.")
	}
	return
}*/

//The subscriber loops run asynchronously, reading from the topic and sending the err
//'job' pairs to their associated subscriber.
/*func subscriber_loop(topic chan stubs.Board, client *rpc.Client, callback string) {
	for {
		nboard := <-topic
		response := new(stubs.BoardResponse)
		err := client.Call(callback, nboard, response)
		if err != nil {
			fmt.Println("Error")
			fmt.Println(err)
			fmt.Println("Closing subscriber thread.")
			//Place the unfulfilled job back on the topic channel.
			topic <- nboard
			break
		}

	}
}*/

//The subscribe function registers a worker to the topic, creating an RPC client,
//and will use the given callback string as the callback function whenever work
//is available.
/*func subscribe(topic string, workerAddress string, callback string) (err error) {
	fmt.Println("Subscription request")
	topicmx.RLock()
	ch := topics[topic]
	topicmx.RUnlock()
	client, err := rpc.Dial("tcp", workerAddress)
	if err == nil {
		go subscriber_loop(ch, client, callback)
	} else {
		fmt.Println("Error subscribing ", workerAddress)
		fmt.Println(err)
		return err
	}
	return
}*/

/*func (w *Worker) NextState(req stubs.Board, res *stubs.BoardResponse) (err error) {
	height := req.P.ImageHeight
	width := req.P.ImageWidth
	newWorld := make([][]byte, height)
	for i := range newWorld {
		newWorld[i] = make([]byte, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			neighbours := countNeighbours(req.P, x, y, req.World)
			if req.World[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	res.NewWorld = newWorld
	return
}*/

const alive = 0xFF
const dead = 0x00

func mod(x, m int) int {
	return (x + m) % m
}

func countNeighbours(p gol.Params, x, y int, world [][]byte) int {
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

func calculateNextState(p gol.Params, world [][]byte, turn int) [][]byte {
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := countNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

type Engine struct{}

func (e *Engine) NewBoard(req stubs.Board, res *stubs.BoardResponse) (err error) {
	var boardRequest *stubs.Board
	if boardRequest == nil {
		err = errors.New("???")
		return
	}

	turn := 0
	for ; turn < req.P.Turns; turn++ {
		req.World = calculateNextState(req.P, req.World, req.Turn)
	}
	go func() {
		res.NewWorld = req.World
		res.NewTurn = turn
	}()
	return
}

// func (e *Engine) CreateChannl(req stubs.PublishRequest, res *stubs.BoardResponse) (err error) {
// 	createTopic(req.Topic)
// 	return
// }
// func (e *Engine) Subscribe(req stubs.Subscription, res *stubs.StatusReport) (err error) {
// 	err = subscribe(req.Topic, req.WorkerAddress, req.Callback)
// 	if err != nil {
// 		res.Message = "Error during subscription"
// 	}
// 	return
// }
// func (e *Engine) Publish(req stubs.PublishRequest, res *stubs.StatusReport) (err error) {
// 	err = publish(req.Topic, req.Board)
// 	return
// }

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	//flag.StringVar(&nextAddr, "next", "localhost:8040", "IP:Port string for the next member")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&Engine{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
