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
	app.Name = "Micro Benchmarks"
	app.Usage = "Runs a set of micro benchmarks to test some assumptions about the choices of data structures."
	app.Authors = []cli.Author{cli.Author{Name: "Jonathan D Fok", Email: ""}}

	app.Commands = []cli.Command{
		cli.Command{
			Name:        "map",
			Usage:       "Run a micro benchmark testing maps vs binary search trees",
			Description: "A Micro benchmark testing the performance difference between maps and BST for existance of value",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "size, s",
					Value: 100,
					Usage: "Size of the sample to be tested.",
				},
			},
			Action: MapVsTreeBenchmark,
		},
	}

	app.Run(os.Args)
}

func MapVsTreeBenchmark(c *cli.Context) {
	size := c.Int("size")

	sample := GenerateDataSet(size)

	sortedSample := SortData(sample)

	testValues := GetRandomFromDataSet(sortedSample, size/10, 100)

	mapTest := make(map[string]Data)

	// Insertion to map
	for i := range sortedSample {
		mapTest[sortedSample[i].Label] = sortedSample[i]
	}

	// Insertion to BST
	bstTest := sortedArrayToBST(sortedSample)

	var mapTestValues, treeTestValues []Data

	startMap := time.Now()

	for i := range testValues {
		mapTestValues = append(mapTestValues, mapTest[testValues[i].Label])
	}
	endMap := time.Since(startMap)

	startTree := time.Now()
	for i := range testValues {
		treeTestValues = append(treeTestValues, bstTest.Get(testValues[i].Label))
	}
	endTree := time.Since(startTree)

	for i := range mapTestValues {
		if mapTestValues[i].Label != testValues[i].Label {
			fmt.Println("MapTest is incorrect")
		}
	}
	for i := range treeTestValues {
		if treeTestValues[i].Label != testValues[i].Label {
			fmt.Println("treeTest is incorrect")
		}
	}

	fmt.Printf("Map: %s\nTree: %s\n", endMap, endTree)
}
