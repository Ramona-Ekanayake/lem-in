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
func (g *Graph) AddRoom(name string, x, y int, isStart, isEnd bool) {
	g.Rooms[name] = Room{Name: name, X: x, Y: y, IsStart: isStart, IsEnd: isEnd}
	if isStart {
		g.StartRoom = name
	}
	if isEnd {
		g.EndRoom = name
	}
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
func readInput(filename string) (*Graph, string, string, int) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("ERROR:", err)
		return nil, "", "", 0
	}
	defer file.Close()

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
			if err != nil {
				fmt.Println("ERROR: invalid number of ants")
				return nil, "", "", 0
			}
			lineNumber++
			continue
		}

		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				fmt.Println("ERROR: invalid connection:", line)
				return nil, "", "", 0
			}
			graph.AddConnection(parts[0], parts[1])
		} else {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				fmt.Println("ERROR: invalid room format:", line)
				return nil, "", "", 0
			}
			name, xStr, yStr := fields[0], fields[1], fields[2]
			x, err := strconv.Atoi(xStr)
			if err != nil {
				fmt.Println("ERROR: invalid x coordinate")
				return nil, "", "", 0
			}
			y, err := strconv.Atoi(yStr)
			if err != nil {
				fmt.Println("ERROR: invalid y coordinate")
				return nil, "", "", 0
			}
			graph.AddRoom(name, x, y, start, end)
			start, end = false, false
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR:", err)
		return nil, "", "", 0
	}
	if graph.StartRoom == "" || graph.EndRoom == "" {
		fmt.Println("ERROR: missing start or end room")
		return nil, "", "", 0
	}
	return graph, graph.StartRoom, graph.EndRoom, graph.AntCount
}

// findAllPaths uses DFS to find all paths from the start room to the end room.
func findAllPaths(graph *Graph, currentRoom string, visited map[string]bool, path []string, allPaths *[][]string) {
	visited[currentRoom] = true
	path = append(path, currentRoom)

	if currentRoom == graph.EndRoom {
		// Make a copy of the path and add it to allPaths
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

	// Backtrack
	path = path[:len(path)-1]
	visited[currentRoom] = false
}

// findShortestPaths finds the shortest paths using BFS.
func findShortestPaths(graph *Graph, start, end string) [][]string {
	var allPaths [][]string
	visited := make(map[string]bool)
	findAllPaths(graph, start, visited, []string{}, &allPaths)

	// Sort paths by length (shortest first)
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i]) < len(allPaths[j])
	})

	return allPaths
}

// distributeAnts assigns ants optimally across paths.
func distributeAnts(paths [][]string, ants int) map[int][]string {
	assignment := make(map[int][]string)
	antIndex := 1

	// Sort paths by length (shortest first)
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})

	// Distribute ants based on path lengths
	for i := 0; i < ants; i++ {
		assignment[antIndex] = paths[i%len(paths)]
		antIndex++
	}

	return assignment
}

// printMovements prints the movements of ants.
func printMovements(assignment map[int][]string, paths [][]string) {
	antPositions := make(map[int]int) // Stores ant index in path
	roomOccupancy := make(map[string]int) // Stores the number of ants in each room
	step := 1

	for {
		var moveStrings []string
		finishedAnts := 0

		for antID := 1; antID <= len(assignment); antID++ {
			currentPosition := antPositions[antID]
			if currentPosition < len(assignment[antID])-1 {
				nextPosition := currentPosition + 1
				nextRoom := assignment[antID][nextPosition]

				// Ensure only one ant occupies a room at a time
				if roomOccupancy[nextRoom] == 0 || nextRoom == paths[0][len(paths[0])-1] {
					// Move the ant to the next room
					antPositions[antID] = nextPosition
					moveStrings = append(moveStrings, fmt.Sprintf("L%d-%s", antID, nextRoom))
					roomOccupancy[nextRoom]++
					if currentPosition > 0 {
						roomOccupancy[assignment[antID][currentPosition]]--
					}
				}
			} else {
				finishedAnts++
			}
		}

		if len(moveStrings) > 0 {
			fmt.Println(strings.Join(moveStrings, " "))
		}

		if finishedAnts == len(assignment) {
			fmt.Println("All ants have reached the end.")
			break
		}
		step++
	}
}

// debugPaths prints all the paths found.
func debugPaths(paths [][]string) {
	fmt.Println("All paths found:")
	for i, path := range paths {
		fmt.Printf("Path %d: %s\n", i+1, strings.Join(path, " -> "))
	}
}

// debugAntCount prints the number of ants.
func debugAntCount(antCount int) {
	fmt.Printf("Number of ants: %d\n", antCount)
}

// main is the entry point of the program.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <input_file>")
		return
	}

	graph, start, end, ants := readInput(os.Args[1])

	// Debug: Print the number of ants
	debugAntCount(ants)

	// Step 2: Find Shortest Paths (BFS)
	paths := findShortestPaths(graph, start, end)
	if len(paths) == 0 {
		fmt.Println("ERROR: No valid path found")
		return
	}

	// Debug: Print all paths found
	debugPaths(paths)

	// Step 5: Distribute Ants Optimally Across Paths
	assignment := distributeAnts(paths, ants)

	// Step 6: Print Ant Movements
	printMovements(assignment, paths)

	// Ensure the program exits properly
	fmt.Println("Program completed.")
	os.Exit(0)
}


