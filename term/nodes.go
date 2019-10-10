package main

import (
	"fmt"
	"strings"

	"github.com/donomii/goof"

	"github.com/mattn/go-shellwords"
)

var currentNode *Node
var currentThing []*Node

type Menu []string

type Node struct {
	Name     string
	SubNodes []*Node
	Command  string
	Data     string
}

func makeNodeShort(name string, subNodes []*Node) *Node {
	return &Node{name, subNodes, name, ""}
}

func makeNodeLong(name string, subNodes []*Node, command, data string) *Node {
	return &Node{name, subNodes, name, data}
}

func UberMenu() *Node {
	node := makeNodeLong("Main menu",
		[]*Node{
			appsMenu(),
			fileManagerMenu(),
		},
		"", "")
	return node

}

var menuData = `
[
"!arc list",
"!git status",
"git add",
"!!git commit",
"!ls -gGh"
]`

var myMenu Menu

func NodesToStringArray(ns []*Node) []string {
	var out []string
	for _, v := range ns {
		out = append(out, v.Name)

	}
	return out

}

func fileManagerMenu() *Node {
	return makeNodeShort("File Manager", []*Node{})
}
func appsMenu() *Node {
	node := makeNodeShort("Applications Menu",
		[]*Node{})
	addTextNodesFromStrStr(node, Apps())
	return node
}

func Apps() [][]string {
	lines := strings.Split(goof.QC([]string{"ls", "/Applications"}), "\n")
	out := [][]string{}
	for _, v := range lines {
		name := strings.TrimSuffix(v, ".app")
		command := fmt.Sprintf("!open \"/Applications/%v\"", v)
		out = append(out, []string{name, command})
	}
	return out
}

func configFile() *Node {
	return makeNodeShort("Edit Config", []*Node{})
}

/*
func AddAppNodes(n *Node) *Node {

}
*/

var header string

func makeStartNode() *Node {
	n := makeNodeShort("Command:", []*Node{})

	return n
}

func b(v int32) bool {
	return v == 1
}

func fflag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func (n *Node) String() string {
	return n.Name
}

func (n *Node) ToString() string {
	return n.Name
}

func findNode(n *Node, name string) *Node {
	if n == nil {
		return n
	}
	for _, v := range n.SubNodes {
		if v.Name == name {
			return v
		}
	}
	return nil

}

func controlMenu() *Node {
	node := makeNodeShort("System controls", []*Node{})
	addTextNodesFromStrStr(node,
		[][]string{
			[]string{"pmset sleepnow"},
		})
	return node
}

func historyMenu() *Node {
	return addHistoryNodes()
}

func addHistoryNodes() *Node {
	src := goof.Command("fish", []string{"-c", "history"})
	lines := strings.Split(src, "\n")
	startNode := makeNodeShort("Previous command lines", []*Node{})
	for _, l := range lines {
		currentNode := startNode
		/*
				var s scanner.Scanner
				s.Init(strings.NewReader(l))
				s.Filename = "example"
				for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			        text := s.TokenText()
					fmt.Printf("%s: %s\n", s.Position, text)
			        if findNode(currentNode, text) == nil {
			            newNode := Node{text, []*Node{}}
			            currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
			            currentNode = &newNode
			        } else {
			            currentNode = findNode(currentNode, text)
			        }
		*/
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}

		}
	}
	return startNode
}

func addTextNodesFromString(startNode *Node, src string) *Node {
	lines := strings.Split(src, "\n")
	return addTextNodesFromStringList(startNode, lines)
}

func appendNewNodeShort(text string, aNode *Node) *Node {
	newNode := makeNodeShort(text, []*Node{})
	aNode.SubNodes = append(aNode.SubNodes, newNode)
	return aNode
}

func addTextNodesFromStringList(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		currentNode := startNode
		args, _ := shellwords.Parse(l)
		for _, text := range args {
			if findNode(currentNode, text) == nil {
				newNode := makeNodeShort(text, []*Node{})
				currentNode.SubNodes = append(currentNode.SubNodes, newNode)
				currentNode = newNode
			} else {
				currentNode = findNode(currentNode, text)
			}
		}
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromCommands(startNode *Node, lines []string) *Node {
	for _, l := range lines {
		appendNewNodeShort(l, startNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], ""}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func addTextNodesFromStrStrStr(startNode *Node, lines [][]string) *Node {
	for _, l := range lines {
		currentNode := startNode
		newNode := Node{l[0], []*Node{}, l[1], l[2]}
		currentNode.SubNodes = append(currentNode.SubNodes, &newNode)
	}

	fmt.Println()
	fmt.Printf("%+v\n", startNode)
	dumpTree(startNode, 0)
	return startNode

}

func dumpTree(n *Node, indent int) {
	fmt.Printf("%*s%s\n", indent, "", n.Name)
	for _, v := range n.SubNodes {
		dumpTree(v, indent+1)
	}

}
