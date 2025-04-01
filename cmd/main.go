package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var filesWithoutExtensions = map[string]bool{
	"license": true,
}

var ignoredFilesAndFolders = map[string]bool{
	".gitignore": true,
	".git":       true,
}

type Node struct {
	name     string
	isDir    bool
	children []*Node
	parent   *Node
	depth    int
}

func main() {
	mode := flag.Int("mode", 0, "0: Create project folders and files\n1: Create project tree structure")
	inputFile := flag.String("input", "", "Input file containing directory structure")
	outputDir := flag.String("output", ".", "Output directory where structure will be created")
	path := flag.String("path", ".", "project path to create structure tree")

	flag.Parse()

	switch *mode {
	case 0:
		if *inputFile == "" {
			fmt.Println("Error: Input file must be specified with -i flag")
			flag.Usage()
			os.Exit(1)
		}

		root, err := parseTree(*inputFile)
		if err != nil {
			fmt.Printf("Error parsing structure: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Creating project structure in: %s\n", *outputDir)
		if err := createFromTree(*outputDir, root); err != nil {
			fmt.Printf("Error creating project structure: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Project structure created successfully!")
	case 1:
		root, err := createTree(*path, 0)
		if err != nil {
			fmt.Printf("Error creating tree: %v\n", err)
			os.Exit(1)
		}

		printTree(root)
	default:
		fmt.Println("invalid mode")
		flag.Usage()
	}
}

func parseTree(filename string) (*Node, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var nodes []*Node
	root := &Node{name: ".", isDir: true}
	currentParent := root
	var currentDepth int = 0

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " ")
		print(line + "\n")
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Calculate depth and name
		depth, name := parseLine(line)
		if name == "" {
			continue
		}

		// Adjust parent based on depth
		if depth > currentDepth {
			// Child of previous node
			currentParent = nodes[len(nodes)-1]
			currentDepth = depth
		} else if depth < currentDepth {
			// Move up the tree
			for currentDepth > depth {
				currentParent = currentParent.parent
				currentDepth--
			}
		}

		node := &Node{
			name:   name,
			isDir:  !strings.Contains(name, ".") && !filesWithoutExtensions[strings.ToLower(name)],
			parent: currentParent,
			depth:  depth,
		}

		currentParent.children = append(currentParent.children, node)
		nodes = append(nodes, node)
		currentDepth = depth
	}

	return root, scanner.Err()
}

func parseLine(line string) (int, string) {
	// Count tree characters to determine depth
	var depth int = 0
	chars := []rune(line)
	for i := 0; i < len(chars); i++ {
		switch chars[i] {
		case '│', '├', '└':
			// Skip tree characters but count depth
			depth++

			// if i+3 < len(chars) && chars[i+1] == '─' && chars[i+2] == '─' && chars[i+3] == ' ' {
			// 	i += 3
			// }
		case ' ', '-', '─':
			continue
		default:
			// Clean up name (remove comments and trim)
			name := string(chars[i:])
			name = strings.Split(name, "#")[0]
			name = strings.Trim(name, " ─│├└")
			return depth, name
		}
	}
	return 0, ""
}

func createFromTree(basePath string, node *Node) error {
	for _, child := range node.children {
		fullPath := filepath.Join(basePath, child.name)

		if child.isDir {
			fmt.Printf("Creating directory: %s\n", fullPath)
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", fullPath, err)
			}
			if err := createFromTree(fullPath, child); err != nil {
				return err
			}
		} else {
			fmt.Printf("Creating file: %s\n", fullPath)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("error creating parent directories for %s: %v", fullPath, err)
			}
			if _, err := os.Create(fullPath); err != nil {
				return fmt.Errorf("error creating file %s: %v", fullPath, err)
			}
		}
	}
	return nil
}

// this function will create a tree structure in the given path and subdirectories
func createTree(path string, depth int) (*Node, error) {

	// start with the root directory and create the tree structure recursively
	directoryName := filepath.Base(path)
	if ignoredFilesAndFolders[directoryName] {
		// skip the ignored directory
		return nil, nil
	}

	parent := &Node{
		name:  directoryName,
		isDir: true,
		depth: depth,
	}

	// list the files and directories in the current directory
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)

		return nil, fmt.Errorf("error reading directory %s: %w", path, err)
	}

	for i := range files {
		if files[i].IsDir() {
			// recursively create the tree for the subdirectory
			subDirPath := filepath.Join(path, files[i].Name())
			dirNode, err := createTree(subDirPath, depth+1)
			if err != nil {
				fmt.Println(err)

				return nil, fmt.Errorf("error creating tree for directory %s: %w", subDirPath, err)
			}

			// add the subdirectory node to the parent node
			if dirNode != nil {
				parent.children = append(parent.children, dirNode)
			}
		} else {
			if ignoredFilesAndFolders[files[i].Name()] {
				// skip the ignored file
				continue
			}

			node := &Node{
				name:   files[i].Name(),
				isDir:  false,
				parent: parent,
				depth:  parent.depth + 1,
			}

			parent.children = append(parent.children, node)
		}
	}

	return parent, nil
}

func printTree(node *Node) {

	for i := range node.depth {
		if i < (node.depth)-1 {
			fmt.Print("│   ")
		} else {
			fmt.Print("│── ")
		}
	}

	if node.isDir {
		fmt.Printf("%s/\n", node.name)
		for i := range node.children {
			printTree(node.children[i])
		}
	} else {
		fmt.Printf("%s\n", node.name)
	}
}
