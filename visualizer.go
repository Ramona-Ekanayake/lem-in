package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Room represents a room in the ant farm.
// Each room has a name, coordinates (X, Y), and flags indicating if it is the start or end room.
type Room struct {
	Name    string
	X, Y    int
	IsStart bool
	IsEnd   bool
}

// Graph represents the entire ant farm.
// It contains rooms, connections between rooms, the number of ants, and the start and end rooms.
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
func (graph *Graph) AddRoom(name string, x, y int, isStart, isEnd bool) {
	graph.Rooms[name] = Room{Name: name, X: x, Y: y, IsStart: isStart, IsEnd: isEnd}
	if isStart {
		graph.StartRoom = name
	}
	if isEnd {
		graph.EndRoom = name
	}
}

// AddConnection adds a connection (tunnel) between two rooms.
func (graph *Graph) AddConnection(roomA, roomB string) error {
	_, roomAExists := graph.Rooms[roomA]
	_, roomBExists := graph.Rooms[roomB]
	if !roomAExists || !roomBExists {
		return fmt.Errorf("unknown room: %s or %s", roomA, roomB)
	}
	graph.Connections[roomA] = append(graph.Connections[roomA], roomB)
	graph.Connections[roomB] = append(graph.Connections[roomB], roomA)
	return nil
}

// parseInput reads the input file and constructs the graph.
func parseInput(filename string) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	graph := NewGraph()
	scanner := bufio.NewScanner(file)

	lineNumber := 0
	var start, end string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			if line == "##start" {
				start = "start"
			} else if line == "##end" {
				end = "end"
			}
			continue
		}
		if lineNumber == 0 {
			graph.AntCount, err = strconv.Atoi(line)
			if err != nil {
				return nil, fmt.Errorf("invalid number of ants")
			}
			lineNumber++
			continue
		}
		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid link format: %s", line)
			}
			roomA, roomB := parts[0], parts[1]
			err := graph.AddConnection(roomA, roomB)
			if err != nil {
				return nil, err
			}
		} else {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				return nil, fmt.Errorf("invalid room format: %s", line)
			}
			name := fields[0]
			x, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid room coordinates")
			}
			y, err := strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("invalid room coordinates")
			}
			isStart := start == "start"
			isEnd := end == "end"
			graph.AddRoom(name, x, y, isStart, isEnd)
			start, end = "", ""
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if graph.StartRoom == "" || graph.EndRoom == "" {
		return nil, fmt.Errorf("missing start or end room")
	}
	return graph, nil
}

// findShortestPath performs a breadth-first search to find the shortest path from the start room to the end room.
func findShortestPath(graph *Graph) []string {
	queue := [][]string{{graph.StartRoom}}
	visitedRooms := make(map[string]bool)
	visitedRooms[graph.StartRoom] = true

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]
		currentRoom := currentPath[len(currentPath)-1]

		if currentRoom == graph.EndRoom {
			return currentPath
		}

		for _, neighbor := range graph.Connections[currentRoom] {
			if !visitedRooms[neighbor] {
				visitedRooms[neighbor] = true
				newPath := append([]string{}, currentPath...) // Copy path
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}
	return nil
}

// visualizeAntMovements visually simulates the movement of ants along the shortest path.
func visualizeAntMovements(graph *Graph, path []string, numAnts int) {
	antPositions := make(map[int]int) // Stores ant index in path
	roomOccupancy := make(map[string]int) // Stores the number of ants in each room
	step := 1

	for {
		var moveStrings []string
		finishedAnts := 0

		for antID := 1; antID <= numAnts; antID++ {
			currentPosition := antPositions[antID]
			if currentPosition < len(path)-1 {
				nextPosition := currentPosition + 1
				nextRoom := path[nextPosition]

				// Ensure only one ant occupies a room at a time
				if roomOccupancy[nextRoom] == 0 || nextRoom == graph.EndRoom {
					// Move the ant to the next room
					antPositions[antID] = nextPosition
					moveStrings = append(moveStrings, fmt.Sprintf("L%d-%s", antID, nextRoom))
					roomOccupancy[nextRoom]++
					if currentPosition > 0 {
						roomOccupancy[path[currentPosition]]--
					}
				}
			} else {
				finishedAnts++
			}
		}

		if len(moveStrings) > 0 {
			fmt.Println(strings.Join(moveStrings, " "))
		}

		if finishedAnts == numAnts {
			break
		}
		step++
		time.Sleep(1 * time.Second) // Add a delay to visualize the movements
	}
}

// main is the entry point of the visualizer program.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run visualizer.go <input_file>")
		return
	}

	filename := os.Args[1]
	graph, err := parseInput(filename)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	shortestPath := findShortestPath(graph)
	visualizeAntMovements(graph, shortestPath, graph.AntCount)
}
