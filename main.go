package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Room represents a room in the ant farm.
type Room struct {
	Name    string
	X, Y    int
	IsStart bool
	IsEnd   bool
}

// Graph represents the entire ant farm.
type Graph struct {
	Rooms       map[string]Room
	Connections map[string][]string
	AntCount    int
	StartRoom   string
	EndRoom     string
}

// NewGraph initializes and returns a new Graph.
func NewGraph() *Graph {
	return &Graph{
		Rooms:       make(map[string]Room),
		Connections: make(map[string][]string),
	}
}

// AddRoom adds a room to the graph.
func (g *Graph) AddRoom(name string, x, y int, isStart, isEnd bool) error {
	if _, exists := g.Rooms[name]; exists {
		return fmt.Errorf("duplicate room: %s", name)
	}
	g.Rooms[name] = Room{Name: name, X: x, Y: y, IsStart: isStart, IsEnd: isEnd}
	if isStart {
		g.StartRoom = name
	}
	if isEnd {
		g.EndRoom = name
	}
	return nil
}

// AddConnection adds a connection (tunnel) between two rooms.
func (g *Graph) AddConnection(roomA, roomB string) error {
	if _, ok := g.Rooms[roomA]; !ok {
		return fmt.Errorf("invalid connection: %s - %s", roomA, roomB)
	}
	if _, ok := g.Rooms[roomB]; !ok {
		return fmt.Errorf("invalid connection: %s - %s", roomA, roomB)
	}
	g.Connections[roomA] = append(g.Connections[roomA], roomB)
	g.Connections[roomB] = append(g.Connections[roomB], roomA)
	return nil

}

// readInput parses the input file and constructs the graph.
func readInput(exampleFile string) (*Graph, string, string, int) {
	file, err := os.Open(exampleFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1) // Exits immediately, skipping the deferred function
	}
	defer file.Close()
	fmt.Println("File opened successfully!")
	fmt.Println()

	graph := NewGraph()
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var start, end bool

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			if line == "##start" {
				start = true
			} else if line == "##end" {
				end = true
			}
			continue
		}

		if lineNumber == 0 {
			graph.AntCount, err = strconv.Atoi(line)
			if err != nil || graph.AntCount <= 0 {
				fmt.Println("ERROR: invalid data format, number of ants must be a positive integer")
				os.Exit(1)
			}
			lineNumber++
			continue
		}

		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				fmt.Println("ERROR: invalid connection:", line)
				os.Exit(1)
			}
			if parts[0] == parts[1] {
				fmt.Println("ERROR: self referencing room:", line)
				os.Exit(0)
			}
			for key, vals := range graph.Connections {
				for _, val := range vals {
					if (key == parts[0] && val == parts[1]) || (key == parts[1] && val == parts[0]) {
						fmt.Println("ERROR: invalid data format, identical connection already exists:", line)
						os.Exit(1)
					}
				}
			}
			graph.AddConnection(parts[0], parts[1])
		} else {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				fmt.Println("ERROR: invalid room format:", line)
				os.Exit(1)
			}
			name, xStr, yStr := fields[0], fields[1], fields[2]
			x, err := strconv.Atoi(xStr)
			if err != nil {
				fmt.Println("ERROR: invalid x coordinate")
				os.Exit(1)
			}
			y, err := strconv.Atoi(yStr)
			if err != nil {
				fmt.Println("ERROR: invalid y coordinate")
				os.Exit(1)
			}
			if err := graph.AddRoom(name, x, y, start, end); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(1)
			}
			start, end = false, false
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(0)
	}
	if graph.StartRoom == "" || graph.EndRoom == "" {
		fmt.Println("ERROR: missing start or end room")
		os.Exit(0)
	}
	return graph, graph.StartRoom, graph.EndRoom, graph.AntCount
}

// debugAntCount prints the number of ants.
func debugAntCount(antCount int) {
	fmt.Printf("Number of ants: %d\n", antCount)
}

// findAllPaths uses DFS to find all paths from the start room to the end room.
func findAllPaths(graph *Graph, currentRoom string, visited map[string]bool, path []string, allPaths *[][]string) {
	visited[currentRoom] = true
	path = append(path, currentRoom)

	if currentRoom == graph.EndRoom {
		pathCopy := make([]string, len(path))
		copy(pathCopy, path)
		*allPaths = append(*allPaths, pathCopy)
	} else {
		for _, neighbor := range graph.Connections[currentRoom] {
			if !visited[neighbor] {
				findAllPaths(graph, neighbor, visited, path, allPaths)
			}
		}
	}

	// Backtracking
	path = path[:len(path)-1]
	visited[currentRoom] = false
}

func findShortestPaths(graph *Graph, start string) [][]string {
	var allPaths [][]string
	visited := make(map[string]bool) // Key:name of the room and if visited
	findAllPaths(graph, start, visited, []string{}, &allPaths)

	// Sort paths by length (shortest first)
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i]) < len(allPaths[j])
	})

	return allPaths
}

func solutionsCompatible(path1, path2 []string, start, end string) bool {
	for _, room1 := range path1 {
		if room1 == start || room1 == end {
			continue
		}
		for _, room2 := range path2 {
			if room1 == room2 {
				return false
			}
		}
	}
	return true
}

// Checks if a given candidate path is compatible with all paths in group
func solutionCompatibleWithGroup(candidate []string, group [][]string, start, end string) bool {
	for _, path := range group {
		if !solutionsCompatible(path, candidate, start, end) {
			return false
		}
	}
	return true
}

// Groups compatible solutions together.
func calculateSolutionGroups(paths [][]string, start, end string) [][][]string {
	var solGroups [][][]string

	if len(paths) <= 1 {
		if len(paths) == 1 {
			solGroups = append(solGroups, paths)
		}
		return solGroups
	}

	for i, path1 := range paths {
		group := [][]string{path1}
		for j, path2 := range paths {
			if i == j {
				continue
			}
			if solutionCompatibleWithGroup(path2, group, start, end) {
				group = append(group, path2)
			}
		}
		solGroups = append(solGroups, group)
	}

	return solGroups
}

func distributeAnts(paths [][]string, ants int) map[int][]string {
	assignment := make(map[int][]string)
	loads := make([]int, len(paths))
	for i, path := range paths {
		loads[i] = len(path)
	}

	// Distribute ants based on the load.
	for antIndex := 1; antIndex <= ants; antIndex++ {
		minLoad := loads[0]
		minIndex := 0
		for i, load := range loads {
			if load < minLoad {
				minLoad = load
				minIndex = i
			}
		}
		assignment[antIndex] = paths[minIndex]
		loads[minIndex]++
	}

	return assignment
}

// getAntMoves prints the movements of ants.
func getAntMoves(originalAssignment map[int][]string, end string) string {
	type AntAssignment struct {
		AntID int
		Path  []string
	}

	// Convert the map into a slice.
	var assignments []AntAssignment
	for antID, path := range originalAssignment {
		assignments = append(assignments, AntAssignment{AntID: antID, Path: path})
	}

	// Sort the slice
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].AntID < assignments[j].AntID
	})

	antMoves := ""
	antPositions := make(map[int]int)
	roomFull := make(map[string]bool)

	for {
		tunnelsUsed := make(map[string]string)
		var moveStrings []string
		finishedAnts := 0

		// Process each ant's movement.
		for i := range assignments {
			currentPosition := antPositions[assignments[i].AntID]

			if currentPosition < len(assignments[i].Path)-1 {
				nextPosition := currentPosition + 1
				currentRoom := assignments[i].Path[currentPosition]
				nextRoom := assignments[i].Path[nextPosition]

				if !roomFull[nextRoom] && tunnelsUsed[currentRoom] != nextRoom {
					antPositions[assignments[i].AntID] = nextPosition
					moveStrings = append(moveStrings, fmt.Sprintf("L%d-%s", assignments[i].AntID, nextRoom))

					if nextRoom != end {
						roomFull[nextRoom] = true
					}
					roomFull[assignments[i].Path[currentPosition]] = false
					tunnelsUsed[currentRoom] = nextRoom
				}
			} else {
				finishedAnts++
			}
		}

		if len(moveStrings) > 0 {
			antMoves += strings.Join(moveStrings, " ") + "\n"
		}

		// When all ants have reached the end of their paths, finish.
		if finishedAnts == len(assignments) {
			break
		}
	}
	return antMoves
}

// debugPaths prints all the paths found.
func debugPaths(paths [][]string) {
	fmt.Println("All paths found:")
	for i, path := range paths {
		fmt.Printf("Path %d: %s", i+1, strings.Join(path, " ->"))
		fmt.Println()
	}
}

// main is the entry point of the program.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <filename.txt>")
		return
	}

	graph, start, end, ants := readInput(os.Args[1])

	// Print all ants on file
	debugAntCount(ants)

	paths := findShortestPaths(graph, start)
	if len(paths) == 0 {
		fmt.Println("ERROR: No valid path found")
		return
	}
	// Print all paths found
	debugPaths(paths)

	solutionGroups := calculateSolutionGroups(paths, start, end)
	if len(solutionGroups) == 0 {
		fmt.Println("ERROR: No compatible solution group found")
		return
	}

	var antMovesPerPath []string
	for _, solutionGroup := range solutionGroups {
		// Step 5: Distribute Ants Optimally Across Paths
		assignment := distributeAnts(solutionGroup, ants)

		// Step 6: Print Ant Movements
		antMovesPerPath = append(antMovesPerPath, getAntMoves(assignment, end))
	}

	shortestSolution := antMovesPerPath[0]
	for _, solution := range antMovesPerPath {
		if strings.Count(solution, "\n") < strings.Count(shortestSolution, "\n") {
			shortestSolution = solution
		}
	}

	steps := strings.Count(shortestSolution, "\n")
	fmt.Println()
	fmt.Println("Shortest path\n" + shortestSolution)
	fmt.Println("Here you go!")
	fmt.Println("Quickest path with", steps, "turns")
}
