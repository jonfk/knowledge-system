package main

import (
	"github.com/satori/go.uuid"
	"math/rand"
	"sort"
)

type Data struct {
	Label string
}

type DataList []Data

func (a DataList) Less(i, j int) bool {
	return a[i].Label < a[j].Label
}

func (a DataList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a DataList) Len() int {
	return len(a)
}

func SortData(a DataList) DataList {
	sort.Sort(a)
	return a
}

func GenerateDataSet(size int) []Data {
	var res []Data
	for i := 0; i < size; i++ {
		newData := Data{
			Label: uuid.NewV4().String(),
		}
		res = append(res, newData)
	}
	return res
}

func DataToPointers(x []Data) []*Data {
	var res []*Data
	for i := range x {
		res = append(res, &x[i])
	}
	return res
}

func GetRandomFromDataSet(data []Data, size int, seed int64) []Data {
	set := make(map[string]Data)

	r := rand.New(rand.NewSource(seed))
	length := len(data)

	for i := 0; i < size; i++ {
		rI := r.Intn(length)
		set[data[rI].Label] = data[rI]
	}
	var res []Data
	for _, v := range set {
		res = append(res, v)
	}
	return res
}
