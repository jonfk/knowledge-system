package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	//"github.com/davecgh/go-spew/spew"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "Knowledge systems: Experiment"
	app.Usage = "Runs a simulation"
	app.Authors = []cli.Author{cli.Author{Name: "Jonathan D Fok", Email: ""}}

	app.Commands = []cli.Command{
		cli.Command{
			Name:        "test",
			Usage:       "Run a test simulation",
			Description: "Runs a test simulation with specified parameters and prints out the amount of time taken to run, excluding setup.",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:   "depth, d",
					Value:  100,
					Usage:  "The depth for each simulation run",
					EnvVar: "SIM_DEPTH",
				},
				cli.IntFlag{
					Name:   "size, s",
					Value:  100,
					Usage:  "The size of the graph for the simulation run. If input is set, this setting is ignored.",
					EnvVar: "SIM_SIZE",
				},
				cli.StringFlag{
					Name:  "input, i",
					Value: "./data/data.json",
					Usage: "Path to json file containing data set. If not set, a random data set is used.",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "./data/{SEED_Value}.json",
					Usage: "Path to output json of data set used.",
				},
			},
			Action: TestSimulation,
		},
		cli.Command{
			Name:        "concurrent",
			Usage:       "Run a concurrent version of the test simulation.",
			Description: `Runs a concurrent version of the test simulation with specified parameters and prints out the amount of time taken to run, excluding setup.`,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:   "depth, d",
					Value:  100,
					Usage:  "The depth for each simulation run",
					EnvVar: "SIM_DEPTH",
				},
				cli.IntFlag{
					Name:   "size, s",
					Value:  100,
					Usage:  "The size of the graph for the simulation run. If input is set, this setting is ignored.",
					EnvVar: "SIM_SIZE",
				},
				cli.IntFlag{
					Name:  "procs, p",
					Value: 1,
					Usage: "The number of processors to be used by the runtime. Can also be set using env var GOMAXPROCS",
				},
				cli.IntFlag{
					Name:  "routines, r",
					Value: 5,
					Usage: "The number of routines to be used by the simulation. Best to set >= procs",
				},
				cli.IntFlag{
					Name:  "buffer, b",
					Value: 100,
					Usage: "The buffer size used by the communication channel between goroutines.",
				},
				cli.StringFlag{
					Name:  "input, i",
					Value: "./data/data.json",
					Usage: "Path to json file containing data set. If not set, a random data set is used.",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "./data/{SEED_Value}.json",
					Usage: "Path to output json of data set used.",
				},
			},
			Action: ConcurrentTestSimulation,
		},
	}

	app.Action = func(c *cli.Context) {
		//Simulation(c.String("input"), c.Int("depth"), c.Int("size"))
		cli.ShowAppHelp(c)
	}

	app.Run(os.Args)
}

func TestSimulation(c *cli.Context) {
	var graph []*LabelNode

	depth := c.Int("depth")
	size := c.Int("size")

	if c.IsSet("input") {
		graph = load(c.String("input"))
		size = len(graph)
		fmt.Println("Graph Loaded")
	} else {
		seed := time.Now().UnixNano()
		graph = generateRandom(size, seed)
		size = len(graph)
		fmt.Println("Graph generated")
	}
	// reset Visited to false
	for i := range graph {
		graph[i].Visited = false
	}

	fmt.Printf("Simulation Info:\nDepth: %d\nGraph Size: %d\n", depth, size)

	var actives map[string]*LabelNode = map[string]*LabelNode{
		graph[0].Label: graph[0],
	}

	start := time.Now()

	for i := 0; i < depth; i++ {
		//spew.Dump(actives)
		for _, node := range actives {
			if !node.Visited {
				for _, childId := range node.Children {
					child := graph[childId]
					actives[child.Label] = child
				}
				node.Visited = true
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Num actives: %d\n", len(actives))
	fmt.Printf("Time taken: %s\n", elapsed)
	if c.IsSet("output") {
		output(graph, c.String("output"))
	}
}

// This version is non deterministic because of race conditions between goroutines to process
// the nodes received.
// When giving the right result, it typically process the same result as non-concurrent version faster.
func ConcurrentTestSimulation(c *cli.Context) {
	var graph []*LabelNode

	depth := c.Int("depth")
	size := c.Int("size")
	var processors int = -1

	if c.IsSet("input") {
		graph = load(c.String("input"))
		size = len(graph)
		fmt.Println("Graph Loaded")
	} else {
		seed := time.Now().UnixNano()
		graph = generateRandom(size, seed)
		size = len(graph)
		fmt.Println("Graph generated")
	}
	// reset Visited to false
	for i := range graph {
		graph[i].Visited = false
	}

	if c.IsSet("procs") {
		processors = c.Int("procs")
		runtime.GOMAXPROCS(processors)
	}

	fmt.Printf("Simulation Info:\nDepth: %d\nGraph Size: %d\nNum of Cores: %d\nGOMAXPROCS: %d\nConcurrent Routines: %d\n", depth, size, runtime.NumCPU(), runtime.GOMAXPROCS(-1), c.Int("routines"))

	var actives map[string]*LabelNode = make(map[string]*LabelNode)

	channelBufferSize := c.Int("buffer")
	collect := make(chan *LabelNode, channelBufferSize)
	sendWorkers := make(chan *LabelNode, channelBufferSize)

	waitGroup := new(sync.WaitGroup)

	start := time.Now()

	collect <- graph[0]

	// Create the collector goroutine that collects the active nodes
	// and send them to the workers the node received has not been visited.
	// It is the only goroutine allowed to do any mutation on the graph, actives
	// or any nodes
	go func(collect <-chan *LabelNode, sendWorkers chan<- *LabelNode) {
		for {
			newActive := <-collect
			if !newActive.Visited {
				newActive.Visited = true
				actives[newActive.Label] = newActive
				sendWorkers <- newActive
			}
		}
	}(collect, sendWorkers)

	// The worker routines receive a node on the receive channel and processes it and
	// sends the children to the collect channel.
	// If nothing is received on the receive channel it busy eats one cycle without processing
	// anything. This is to prevent deadlocks when there is nothing to process left and the goroutines
	// are left waiting for a never coming node.
	numGoroutines := c.Int("routines")
	for i := 0; i < numGoroutines; i++ {
		waitGroup.Add(1)
		go func(collect chan<- *LabelNode, receive <-chan *LabelNode) {
			for x := 0; x < depth; x++ {
				select {
				case node := <-receive:
					for _, childId := range node.Children {
						collect <- graph[childId]
					}
				default:
					//fmt.Printf("x: %d", x)
				}
			}
			waitGroup.Done()
			return
		}(collect, sendWorkers)
	}

	waitGroup.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Num actives: %d\n", len(actives))
	fmt.Printf("Time taken: %s\n", elapsed)
	if c.IsSet("output") {
		output(graph, c.String("output"))
	}
}
