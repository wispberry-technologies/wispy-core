package wispytail

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// TrieNode represents a node in the trie.
type TrieNode struct {
	children map[rune]*TrieNode
	value    string
	isEnd    bool
	cssRule  string
}

// NewTrieNode creates a new trie node.
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

// Trie is the main structure wrapping the trie root.
type Trie struct {
	root *TrieNode
}

// NewTrie creates a new trie.
func NewTrie() *Trie {
	return &Trie{root: NewTrieNode()}
}

// Insert adds a class and its CSS rule into the trie.
func (t *Trie) Insert(className, cssRule string) {
	node := t.root
	for _, char := range className {
		if _, exists := node.children[char]; !exists {
			node.children[char] = NewTrieNode()
		}
		node = node.children[char]
	}
	node.isEnd = true
	node.cssRule = cssRule
}

// Search looks up a class name and returns its CSS rule if found.
func (t *Trie) Search(className string) (string, bool) {
	node := t.root
	for _, char := range className {
		if child, exists := node.children[char]; exists {
			node = child
		} else {
			return "", false
		}
	}
	if node.isEnd {
		return node.cssRule, true
	}
	return "", false
}

// Dump writes the structure of the trie to the provided io.Writer.
func (t *Trie) Dump(w io.Writer) {
	t.dumpNode(w, t.root, 0)
}

// dumpNode is a helper function to recursively write the trie nodes to the io.Writer.
func (t *Trie) dumpNode(w io.Writer, node *TrieNode, level int) {
	indent := strings.Repeat("  ", level)
	for char, child := range node.children {
		fmt.Fprintf(w, "%s%c", indent, char)
		if child.isEnd {
			fmt.Fprintf(w, " (end: %s)", child.cssRule)
		}
		fmt.Fprintln(w)
		t.dumpNode(w, child, level+1)
	}
}

// ConvertToCSS traverses the trie and generates CSS rules for all classes.
func (t *Trie) ConvertToCSS(w io.Writer, buildSelectorFunc func(className string, prefixes []string) (string, string)) {
	t.convertNodeToCSS(w, t.root, "", buildSelectorFunc, nil)
}

// convertNodeToCSS is a helper function to recursively generate CSS rules.
func (t *Trie) convertNodeToCSS(w io.Writer, node *TrieNode, prefix string, buildSelectorFunc func(className string, prefixes []string) (string, string), prefixes []string) {
	if node.isEnd {
		// Generate the selector and media query for the current class
		selector, mediaQuery := buildSelectorFunc(prefix, prefixes)
		// Write the CSS rule to the writer
		if mediaQuery != "" {
			fmt.Fprintf(w, "%s { %s { %s } }\n", mediaQuery, selector, node.cssRule)
		} else {
			fmt.Fprintf(w, "%s { %s }\n", selector, node.cssRule)
		}
	}

	// Recursively process all child nodes
	for char, child := range node.children {
		t.convertNodeToCSS(w, child, prefix+string(char), buildSelectorFunc, prefixes)
	}
}

// WriteCSSToFile writes the entire trie's CSS rules to a file.
func (t *Trie) WriteCSSToFile(filename string, buildSelectorFunc func(className string, prefixes []string) (string, string)) error {
	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Convert the trie to CSS and write to the file
	t.ConvertToCSS(file, buildSelectorFunc)
	return nil
}

// HasPrefix checks if any key in the trie starts with the given prefix
func (t *Trie) HasPrefix(prefix string) bool {
	node := t.root
	for _, char := range prefix {
		if _, exists := node.children[char]; !exists {
			return false
		}
		node = node.children[char]
	}
	return true
}
