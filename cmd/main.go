package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ErikPitko/lammps-parser/distance"
	"github.com/ErikPitko/lammps-parser/trajectory"
	"github.com/ErikPitko/lammps-parser/utilities"

	"github.com/manifoldco/promptui"
	"golang.org/x/sync/semaphore"
)

type initValsT struct {
	fileName      string
	outfName      string
	distance      bool
	trajectory    bool
	trajectoryOut string
	atomList      map[string]struct{}
}

type internalStateT struct {
	currStep int
	timeStep int
}

var empty struct{}
var initVals initValsT
var iState internalStateT

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readFileAtoms(file *os.File) map[string]struct{} {
	atomList := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	readTimestep := false
	timeStepSet := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ITEM: ATOMS id x y z") {
			iState.currStep++
			continue
		}
		if strings.HasPrefix(line, "ITEM: TIMESTEP") {
			readTimestep = true
			continue
		}
		if readTimestep {
			if !timeStepSet {
				iState.timeStep, _ = strconv.Atoi(line)
			}
			readTimestep = false
			timeStepSet = true
			continue
		}
		if iState.currStep == 1 {
			if strings.HasPrefix(line, "ITEM:") {
				break
			}
			split := strings.Split(line, " ")
			last := split[len(split)-1]
			if last == "" {
				atomList[split[len(split)-2]] = empty
			} else {
				atomList[last] = empty
			}
		}
	}
	err := scanner.Err()
	check(err)
	return atomList
}

func readSteps(file *os.File, filter []string, trajectoryAtom string) {
	// Open file for writing if specified
	var fOut *os.File
	if initVals.outfName != "" {
		var err error
		fOut, err = os.Create(initVals.outfName)
		check(err)
		defer fOut.Close()
	}
	var trajOut *os.File
	if initVals.trajectory {
		var err error
		trajOut, err = os.Create(initVals.trajectoryOut)
		if err != nil {
			panic("Could not create trajectory output file")
		}
		defer trajOut.Close()
	}

	sort.Strings(filter)
	initialPosition := make(map[string][]float64)
	lineChannel := make(chan *utilities.LineT, 12000000) // magic constant - guess the number of steps?
	trajChannel := make(chan *utilities.LineT, 12000000)
	trajColumnCh := make(chan int)

	// upper limit for buffer, otherwise we will run out of memory
	trajectoryLimiter := semaphore.NewWeighted(500)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var trajMu sync.Mutex
	var trajWg sync.WaitGroup

	nextStep := false
	scanner := bufio.NewScanner(file)
	lineBuffer := []string{}
	ctx := context.TODO()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ITEM: ATOMS id x y z") {
			nextStep = false
			if iState.currStep > 1 {
				// every step after reading initial position
				lt := &utilities.LineT{Line: lineBuffer, StepN: iState.currStep - 1}
				if initVals.distance {
					wg.Add(1)
					lineChannel <- lt
				}
				if initVals.trajectory {
					trajWg.Add(1)
					trajectoryLimiter.Acquire(ctx, 1)
					trajChannel <- lt
				}
				lineBuffer = []string{}
			} else if iState.currStep == 1 {
				// reading initial position
				if initVals.distance {
					for w := 0; w < 10; w++ {
						go distance.Consumer(fOut, &wg, &mu, &initialPosition, lineChannel, filter, iState.timeStep)
					}
				}
				if initVals.trajectory {
					trajWg.Add(1)
					trajectoryLimiter.Acquire(ctx, 1)
					trajChannel <- &utilities.LineT{Line: lineBuffer, StepN: iState.currStep - 1}
				}
				lineBuffer = []string{}
			} else if initVals.trajectory {
				// before reading initial position
				// increasing workers probably won't mean better performance, tajectory is generating a lot of data and writing it to file
				for w := 0; w < 5; w++ {
					go trajectory.Consumer(trajOut, trajectoryLimiter, &trajWg, &trajMu, trajChannel, []string{trajectoryAtom}, trajColumnCh, iState.timeStep)
				}
			}
			iState.currStep++
			continue
		}
		if strings.HasPrefix(line, "ITEM:") {
			nextStep = true
		}
		if iState.currStep == 1 && !nextStep {
			aName, tmp, err := utilities.FilterParseStep(line, filter)
			if err == nil {
				ret := utilities.StrToFloat(tmp)
				initialPosition[aName] = ret
			}
		}
		if iState.currStep > 0 && !nextStep {
			lineBuffer = append(lineBuffer, line)
		}
	}
	err := scanner.Err()
	check(err)
	if len(lineBuffer) > 0 {
		if initVals.distance {
			wg.Add(1)
			lineChannel <- &utilities.LineT{Line: lineBuffer, StepN: iState.currStep - 1}
		}
		if initVals.trajectory {
			trajWg.Add(1)
			trajectoryLimiter.Acquire(ctx, 1)
			trajChannel <- &utilities.LineT{Line: lineBuffer, StepN: iState.currStep - 1}
		}
	}
	wg.Wait()
	trajWg.Wait()
	close(lineChannel)
	close(trajChannel)
	if initVals.trajectory {
		columnN := <-trajColumnCh
		fmt.Fprintln(os.Stderr, "Num of columns: ", columnN*3+1)
	}
}

func getFilteredList(label string, multiple bool, keys []string) []string {
	if multiple {
		keys = append(keys, "Done")
	}
	var filter []string
	for len(keys) > 0 {
		prompt := promptui.Select{
			Label: label,
			Items: keys,
		}
		ind, sel, err := prompt.Run()
		if err != nil {
			fmt.Println("Interrupted.")
			os.Exit(1)
		}
		if multiple && ind == len(keys)-1 {
			break
		}
		filter = append(filter, sel)
		if !multiple {
			break
		}
		copy(keys[ind:], keys[ind+1:])
		keys[len(keys)-1] = ""
		keys = keys[:len(keys)-1]
	}
	return filter
}

func readInputFile() {
	f, err := os.Open(initVals.fileName)
	check(err)
	defer f.Close()
	initVals.atomList = readFileAtoms(f)
	keys := make([]string, 0, len(initVals.atomList))
	for k := range initVals.atomList {
		keys = append(keys, k)
	}
	filter := []string{""}
	if initVals.distance {
		filter = getFilteredList("What atoms should I look for?", true, keys)
	}

	promptTrajectory := promptui.Prompt{
		Label:     "Would you like to compute trajectory?",
		IsConfirm: true,
	}
	res, _ := promptTrajectory.Run()
	if res == "y" {
		initVals.trajectory = true
	} else {
		initVals.trajectory = false
	}
	trajectoryAtom := ""
	if initVals.trajectory {
		trajectoryAtom = getFilteredList("What atom should I trace", false, keys)[0]
		promptProm := promptui.Prompt{
			Label: "Output file name for trajectory output",
		}
		initVals.trajectoryOut, _ = promptProm.Run()
	}

	if !initVals.distance && !initVals.trajectory {
		return
	}
	f.Seek(0, 0)
	iState.currStep = 0

	// defer profile.Start().Stop()
	start := time.Now()
	readSteps(f, filter, trajectoryAtom)
	fmt.Fprintf(os.Stderr, "Done in %s\n", time.Since(start))
}

func initQuestions() {
	promptDistance := promptui.Prompt{
		Label:     "Would you like to compute distance?",
		IsConfirm: true,
	}
	res, _ := promptDistance.Run()
	if res == "y" {
		initVals.distance = true
		promptProm := promptui.Prompt{
			Label: "Output file name (leave empty for console output)",
		}
		initVals.outfName, _ = promptProm.Run()
	} else {
		initVals.distance = false
	}
}

func main() {
	iState.currStep = 0
	initVals.fileName = os.Args[1]
	initQuestions()
	readInputFile()
}
