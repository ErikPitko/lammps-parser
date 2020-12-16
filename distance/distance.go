package distance

import (
	"fmt"
	"os"
	"sync"

	"github.com/ErikPitko/lammps-parser/utilities"
	"github.com/OGFris/GoStats"
)

func processStep(steps [][]float64) []float64 {
	tsteps := utilities.Transpose(steps)
	xMean := GoStats.Mean(tsteps[0])
	yMean := GoStats.Mean(tsteps[1])
	zMean := GoStats.Mean(tsteps[2])
	sumMean := (xMean + yMean + zMean) / 3
	return []float64{xMean, yMean, zMean, sumMean}
}

func writeStep(stepMean []float64, f *os.File, step int, timestep int) {
	out := fmt.Sprintf("%d,%f,%f,%f,%f\n",
		step*timestep,
		stepMean[0], stepMean[1], stepMean[2], stepMean[3])
	if f != nil {
		f.WriteString(out)
	} else {
		fmt.Print(out)
	}
}

func Consumer(fOut *os.File, wg *sync.WaitGroup, fileMutex *sync.Mutex, initialPosition *map[string][]float64, lineCh chan *utilities.LineT, filter []string, timestep int) {
	for lineT := range lineCh {
		var steps [][]float64
		for _, line := range lineT.Line {
			aName, tmp, err := utilities.FilterParseStep(line, filter)
			if err != nil {
				continue
			}
			ret := utilities.StrToFloat(tmp)
			initial := (*initialPosition)[aName]
			for i := 0; i < 3; i++ {
				ret[i] = (ret[i] - initial[i]) * (ret[i] - initial[i])
			}
			steps = append(steps, ret)
		}

		stepMean := processStep(steps)
		steps = nil
		fileMutex.Lock()
		writeStep(stepMean, fOut, lineT.StepN, timestep)
		fileMutex.Unlock()
		wg.Done()
	}
}
