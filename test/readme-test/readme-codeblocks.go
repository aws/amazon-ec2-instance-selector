package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// CodeBlock models the rundoc codeblock output
type CodeBlock struct {
	Code        string   `json:"code"`
	Interpreter string   `json:"interpreter"`
	Runs        []string `json:"Runs"`
	Tags        []string `json:"tags"`
}

// RunDoc is the outer model for rundocs output
type RunDoc struct {
	CodeBlocks []CodeBlock `json:"code_blocks"`
}

// main takes a rundoc style report parsed from a README file and compares against actual file contents from the tag.
// This is useful to ensure that code examples in the readme are up-to-date with actual example go source files
// If they are in-sync, the source files are executed to make sure the functionality also works.
func main() {
	currentDir := flag.String("current-dir", "", "The current dir this script is called from")
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	var cb strings.Builder
	cb.Grow(32)
	for scanner.Scan() {
		fmt.Fprintf(&cb, "%s", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
	codeBlocksJSON := cb.String()
	runDoc := RunDoc{}
	if err := json.Unmarshal([]byte(codeBlocksJSON), &runDoc); err != nil {
		log.Fatal(err)
	}
	for _, codeBlock := range runDoc.CodeBlocks {
		code := codeBlock.Code
		tags := removeFromSlice(codeBlock.Tags, []string{codeBlock.Interpreter})
		codeFileDir := fmt.Sprintf("%s/../../%s", *currentDir, tags[0])

		switch codeBlock.Interpreter {
		case "go":
			if !compareBlockWithFile(code, codeFileDir) {
				log.Fatalf("Code Block found in README.md does not match corresponding source file: %s", codeFileDir)
			}
		}
	}
}

func compareBlockWithFile(codeBlock string, codePath string) bool {
	fileContents, err := os.ReadFile(codePath)
	if err != nil {
		log.Fatalf("Unable to read file contents at %s", codePath)
	}
	fileContentStr := removeWhitespace(string(fileContents))
	codeBlock = removeWhitespace(string(codeBlock))
	return fileContentStr == codeBlock
}

func removeFromSlice(original []string, removals []string) []string {
	newSlice := []string{}
	for i, element := range original {
		for _, removal := range removals {
			if removal == element {
				newSlice = append(original[:i], original[i+1:]...)
			}
		}
	}
	return newSlice
}

func removeWhitespace(original string) string {
	removed := strings.ReplaceAll(original, " ", "")
	removed = strings.ReplaceAll(removed, "\t", "")
	return strings.ReplaceAll(removed, "\n", "")
}
