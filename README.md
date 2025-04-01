# Usage <br>
-mode: 0: Create project folders and files 1: Create project tree structure <br>
-input: Input file containing directory structure <br>
-output: output directory where structure will be created <br>
-path: project path to create structure tree <br>

example usage: ```go run cmd/main.go -mode 0 -input example.txt -output ../.```

after running above, you can also print the tree structure using the ```go run cmd/main.go -mode 1 -path ../example```