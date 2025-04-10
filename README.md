# LAMMPS Parser (bestScript)

[![Go Report Card](https://goreportcard.com/badge/github.com/ErikPitko/lammps-parser)](https://goreportcard.com/report/github.com/ErikPitko/lammps-parser)

A Go tool for analyzing LAMMPS molecular dynamics simulation output files. This utility parses LAMMPS trajectory data to calculate atomic distances and track atom trajectories over time with parallel processing for improved performance.

## Overview

This tool provides an interactive command-line interface for processing LAMMPS output files. It allows scientists and researchers to extract specific information about atomic movements and interactions from simulation data.

### Key Features

- **Distance Calculation**: Calculate distances between selected atoms across simulation timesteps
- **Trajectory Tracking**: Follow the movement of specific atoms throughout the simulation
- **Parallel Processing**: Utilizes Go's concurrency features for efficient data handling
- **Interactive Selection**: User-friendly interface to select atoms of interest
- **Flexible Output**: Results can be written to files or displayed in the console

## Installation

### Prerequisites

- Go 1.13 or higher
- Git
- (Optional) UPX for executable compression

### Clone repository
```bash
git clone https://github.com/ErikPitko/lammps-parser.git
cd lammps-parser
```

### Build for current platform
```bash
make build
```

### Build Windows executable
```bash
make build-win
```

#### Build Targets
- `build`: Creates binary for Linux/Unix systems in `bin/` directory
- `build-win`: Cross-compiles Windows executable with optimized CGO flags
- Note: Remove `#` before upx commands in Makefile to enable executable compression

## Usage

Run the binary with a LAMMPS trajectory file as the input:

```bash
./bin/lammps-parser path/to/your/lammps_output.lammpstrj
```

The program will guide you through several prompts:

1. Whether you want to compute distances between atoms (Y/n)
2. If yes, the output file name (leave empty for console output)
3. Selection of atoms to track (multiple selections possible)
4. Whether you want to compute atom trajectories
5. If yes, which atom to trace and the trajectory output file name

## Project Structure

The project is organized into several packages:

- **distance**: Handles calculations of distances between atoms
- **trajectory**: Tracks atom positions over time
- **utilities**: Contains helper functions for parsing and data conversion

## Implementation Details

### Concurrency Model

The parser implements a highly concurrent processing model:

- Uses goroutines to process timesteps in parallel
- Implements channels for communication between processors
- Uses semaphores to limit memory consumption during trajectory processing
- Synchronizes file access with mutexes

### File Processing

The program processes LAMMPS trajectory files in a streaming fashion:

1. First pass identifies all atom types
2. Second pass processes the data for selected atoms
3. Data is buffered in memory efficiently to handle large files

## Dependencies

- [promptui](https://github.com/manifoldco/promptui): Interactive command-line prompts
- [golang.org/x/sync/semaphore](https://golang.org/x/sync/semaphore): Semaphore implementation for Go
- [GoStats](github.com/OGFris/GoStats)

## Performance Considerations

- The program uses buffered channels with a capacity based on the estimated number of steps
- A semaphore limits the number of trajectory calculations in flight to prevent memory exhaustion
- Parallel workers process distance calculations, improving throughput on multi-core systems

## Contributing

Contributions to improve the parser are welcome. Please feel free to submit issues or pull requests to the repository.

## Author

[Erik Pitko](https://github.com/ErikPitko)