package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	// "github.com/davecgh/go-spew/spew"
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
					Usage: "The number of routines to be used by the simulation. Best to set <= procs",
				},
				cli.IntFlag{
					Name:  "buffer, b",
					Value: 100,
					Usage: "The buffer size used by the communication channel between goroutines. If not set it is scaled to graph size: size * 10",
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
		//graph = generateRandomGraph(size, seed)
		graph = generateRandomTreeWithRules(4, size, seed)
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
				node.Visited = true
				for _, childId := range node.Children {
					child := graph[childId]
					if child.Rule == nil || Interpret(actives, child.Rule) {
						actives[child.Label] = child
					}
				}
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
	channelBufferSize := c.Int("buffer")

	if c.IsSet("input") {
		graph = load(c.String("input"))
		size = len(graph)
		fmt.Println("Graph Loaded")
	} else {
		seed := time.Now().UnixNano()
		//graph = generateRandomGraph(size, seed)
		graph = generateRandomTreeWithRules(4, size, seed)
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

	if !c.IsSet("buffer") {
		channelBufferSize = size * 10
	}
	collect := make(chan *LabelNode, channelBufferSize)
	sendWorkers := make(chan *LabelNode, channelBufferSize)

	waitGroup := new(sync.WaitGroup)

	// NOTE: No locking is used on the data structures until a more elegant solution is found for
	// concurrent communication on the graph

	// DEPRECATED locks
	// Actives mutex is used for reading the actives map in interpreting rules for
	// nodes with rules.
	// It is unnecessary to have a mutex on graph since at most 1 goroutine can access a
	// node that is not yet in actives. Once a node is labeled as visited and added to the
	// actives, it is sent to the worker routines.
	// activesMutex := new(sync.RWMutex)

	// Assumptions
	// Note: Unnecessary to use mutexes if only one routine interprets
	// Note: Also unnecessary as long as nodes can only be modified in 1 goroutine

	start := time.Now()

	collect <- graph[0]

	// Create the collector goroutine that collects the active nodes
	// and send them to the workers the node received has not been visited.
	// It is the only goroutine allowed to do any mutation on the graph, actives
	// or any nodes
	go func(collect <-chan *LabelNode, sendWorkers chan<- *LabelNode) {
		for {
			select {
			case newActive := <-collect:
				if !newActive.Visited {
					newActive.Visited = true
					if newActive.Rule == nil || Interpret(actives, newActive.Rule) {
						actives[newActive.Label] = newActive
						sendWorkers <- newActive
					}
				}
			}
		}
	}(collect, sendWorkers)

	// The worker routines receive a node on the receive channel and processes it and
	// sends the children to the collect channel.
	// The select statement takes a node from the receive channel if there is a value to be received.
	// If there is none, the select statement blocks until either the timeout channel sends a value
	// and terminates the goroutine or it receives a value from the receive channel.
	// When it receives a value on the receive channel the timeout channel is reset.
	// Timeout is currently 10 Milliseconds, which introduces a lower bound to the processing if more
	// processing resources (goroutines or depth) is allocated to the simulation.
	numGoroutines := c.Int("routines")
	for i := 0; i < numGoroutines; i++ {
		waitGroup.Add(1)
		go func(collect chan<- *LabelNode, receive <-chan *LabelNode) {
			timeout := time.After(10 * time.Millisecond)
			for x := 0; x < depth; x++ {
				select {
				case node := <-receive:
					for _, childId := range node.Children {
						collect <- graph[childId]
					}
					timeout = time.After(10 * time.Millisecond)
				case <-timeout:
					waitGroup.Done()
					return
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

func Interpret(actives map[string]*LabelNode, rule []string) bool {
	for i := range rule {
		if actives[rule[i]] == nil {
			return false
		}
	}
	return true
}

func ConcurrentInterpret(mutex *sync.RWMutex, actives map[string]*LabelNode, rule []string) bool {
	mutex.RLock()
	for i := range rule {
		if actives[rule[i]] == nil {
			mutex.RUnlock()
			return false
		}
	}
	mutex.RUnlock()
	return true

}
