package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	//"github.com/davecgh/go-spew/spew"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "Knowledge systems: Experiment"
	app.Usage = "Runs a simulation"
	app.Authors = []cli.Author{cli.Author{Name: "Jonathan D Fok", Email: ""}}
	//app.Flags =

	app.Commands = []cli.Command{
		cli.Command{
			Name:        "test",
			Usage:       "Run a test simulation with random data",
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
					Usage:  "The size of the graph for the simulation run",
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
	} else {
		seed := time.Now().UnixNano()
		graph = generateRandom(size, seed)
		fmt.Println("Graph generated")
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

func Simulation(filepath string, depth, size int) {
	//graph := load(filepath)

	seed := time.Now().UnixNano()
	graph := generateRandom(size, seed)
	//spew.Dump(graph)
	fmt.Printf("Depth: %d\n", depth)
	fmt.Println("Graph generated")

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
	//spew.Dump(actives)
	fmt.Printf("actives: %d\n", len(actives))
	fmt.Printf("Time: %s\n", elapsed)
}
