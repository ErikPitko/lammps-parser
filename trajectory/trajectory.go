package trajectory

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ErikPitko/lammps-parser/utilities"

	"golang.org/x/sync/semaphore"
)

func write(fOut *os.File, mu *sync.Mutex, out *map[string]string, step int, timestep int) {
	keys := make([]string, 0, len(*out))
	for k := range *out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	mu.Lock()
	var outBuilder strings.Builder
	outBuilder.WriteString(fmt.Sprintf("%d", step*timestep))
	for _, k := range keys {
		outBuilder.WriteRune(',')
		outBuilder.WriteString((*out)[k])
	}
	outBuilder.WriteRune('\n')
	fOut.WriteString(outBuilder.String())
	mu.Unlock()
}

func Consumer(fOut *os.File, limiter *semaphore.Weighted, wg *sync.WaitGroup, fileMutex *sync.Mutex, lineCh chan *utilities.LineT, filter []string, outCh chan int, timestep int) {
	var stepMap map[string]string
	for lineT := range lineCh {
		stepMap = make(map[string]string)
		for _, line := range lineT.Line {
			aName, ret, err := utilities.FilterParseStep(line, filter)
			if err != nil {
				continue
			}
			stepMap[aName] = strings.Join(ret, ",")
		}
		write(fOut, fileMutex, &stepMap, lineT.StepN, timestep)
		wg.Done()
		limiter.Release(1)
	}
	outCh <- len(stepMap)
}
