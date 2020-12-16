package utilities

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

type LineT struct {
	Line  []string
	StepN int
}

func Transpose(x [][]float64) [][]float64 {
	out := make([][]float64, len(x[0]))
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(x[0]); j++ {
			out[j] = append(out[j], x[i][j])
		}
	}
	return out
}

func StrToFloat(in []string) []float64 {
	var ret []float64
	for _, line := range in {
		par, _ := strconv.ParseFloat(line, 64)
		ret = append(ret, par)
	}
	return ret
}

func FilterParseStep(line string, filter []string) (string, []string, error) {
	split := strings.Split(line, " ")

	// check if atom is filtered
	if len(filter) > 0 {
		i := sort.Search(len(filter),
			func(i int) bool { return filter[i] >= split[4] })
		if !((i < len(filter)) && filter[i] == split[4]) {
			return "", nil, os.ErrNotExist // return nothing if filtered
		}
	}

	aName := split[0] // name of the atom
	return aName, split[1:4], nil
}
