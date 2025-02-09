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

// findShortestPaths finds multiple shortest paths using BFS.
func findShortestPaths(graph *Graph, start, end string) [][]string {
	var paths [][]string
	queue := [][]string{{start}}
	visited := make(map[string]bool)
	visited[start] = true

	for len(queue) > 0 {
		levelSize := len(queue)
		newVisited := make(map[string]bool)

		for i := 0; i < levelSize; i++ {
			path := queue[0]
			queue = queue[1:]
			last := path[len(path)-1]

			if last == end {
				paths = append(paths, path)
				continue
			}

			for _, neighbor := range graph.Connections[last] {
				if !visited[neighbor] && !newVisited[neighbor] {
					newPath := append([]string{}, path...)
					newPath = append(newPath, neighbor)
					queue = append(queue, newPath)
					newVisited[neighbor] = true
				}
			}
		}

		for node := range newVisited {
			visited[node] = true
		}
	}

	return paths
}

// findMultiplePaths finds additional shortest paths while minimizing overlap.
func findMultiplePaths(graph *Graph, paths [][]string, start, end string) [][]string {
	// Use a set to track rooms that are already part of a path
	usedRooms := make(map[string]bool)
	for _, path := range paths {
		for _, room := range path {
			usedRooms[room] = true
		}
	}

	// Find additional paths that minimize overlap with existing paths
	var additionalPaths [][]string
	queue := [][]string{{start}}
	visited := map[string]bool{start: true}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		last := path[len(path)-1]

		if last == end {
			// Check if the path has minimal overlap with existing paths
			overlap := false
			for _, room := range path {
				if usedRooms[room] {
					overlap = true
					break
				}
			}
			if !overlap {
				additionalPaths = append(additionalPaths, path)
				for _, room := range path {
					usedRooms[room] = true
				}
			}
		}

		for _, neighbor := range graph.Connections[last] {
			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := append([]string{}, path...)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}

	return append(paths, additionalPaths...)
}

// edmondsKarp applies the max-flow algorithm to model ants as flow.
func edmondsKarp(graph *Graph, start, end string) int {
	// Initialize residual capacities
	residual := make(map[string]map[string]int)
	for room := range graph.Rooms {
		residual[room] = make(map[string]int)
		for _, neighbor := range graph.Connections[room] {
			residual[room][neighbor] = 1 // Capacity of 1 for each connection
		}
	}

	// Helper function to perform BFS and find an augmenting path
	bfs := func() []string {
		queue := [][]string{{start}}
		visited := map[string]bool{start: true}
		parent := make(map[string]string)

		for len(queue) > 0 {
			path := queue[0]
			queue = queue[1:]
			last := path[len(path)-1]

			if last == end {
				// Reconstruct the path
				var augmentingPath []string
				for node := end; node != start; node = parent[node] {
					augmentingPath = append([]string{node}, augmentingPath...)
				}
				augmentingPath = append([]string{start}, augmentingPath...)
				return augmentingPath
			}

			for neighbor, capacity := range residual[last] {
				if !visited[neighbor] && capacity > 0 {
					visited[neighbor] = true
					parent[neighbor] = last
					newPath := append([]string{}, path...)
					newPath = append(newPath, neighbor)
					queue = append(queue, newPath)
				}
			}
		}
		return nil
	}

	// Main loop of the Edmonds-Karp algorithm
	maxFlow := 0
	for {
		augmentingPath := bfs()
		if augmentingPath == nil {
			break
		}

		// Find the bottleneck capacity
		bottleneck := 1
		for i := 0; i < len(augmentingPath)-1; i++ {
			u, v := augmentingPath[i], augmentingPath[i+1]
			if residual[u][v] < bottleneck {
				bottleneck = residual[u][v]
			}
		}

		// Update residual capacities
		for i := 0; i < len(augmentingPath)-1; i++ {
			u, v := augmentingPath[i], augmentingPath[i+1]
			residual[u][v] -= bottleneck
			residual[v][u] += bottleneck
		}

		maxFlow += bottleneck
	}

	return maxFlow
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
	antCount := make([]int, len(paths))
	for i := 0; i < ants; i++ {
		// Assign ant to the least loaded path
		minIndex := 0
		for j := 1; j < len(paths); j++ {
			if len(paths[j])+antCount[j] < len(paths[minIndex])+antCount[minIndex] {
				minIndex = j
			}
		}
		antCount[minIndex]++
		assignment[antIndex] = paths[minIndex]
		antIndex++
	}

	return assignment
}

// printMovements prints the movements of ants.
func printMovements(assignment map[int][]string) {
	for step := 0; ; step++ {
		var moveStrings []string
		finished := true

		for antID, path := range assignment {
			if step < len(path)-1 {
				moveStrings = append(moveStrings, fmt.Sprintf("L%d-%s", antID, path[step+1]))
				finished = false
			}
		}

		if finished {
			break
		}

		fmt.Println(strings.Join(moveStrings, " "))
	}
}

// main is the entry point of the program.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <input_file>")
		return
	}

	graph, start, end, ants := readInput(os.Args[1])

	// Step 2: Find Shortest Paths (BFS)
	paths := findShortestPaths(graph, start, end)
	if len(paths) == 0 {
		fmt.Println("ERROR: No valid path found")
		return
	}

	// Step 3: Find Additional Shortest Paths Minimizing Overlap
	paths = findMultiplePaths(graph, paths, start, end)

	// Step 4: Apply Edmonds-Karp for Maximum Flow (Ant Flow Simulation)
	flow := edmondsKarp(graph, start, end)

	// Step 5: Distribute Ants Optimally Across Paths
	assignment := distributeAnts(paths, ants)

	// Step 6: Print Ant Movements
	printMovements(assignment)
}
