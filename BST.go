package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
)

type Tree struct{
	root *Node
}

type Node struct{
	Val int
	left *Node
	right *Node
}

func in_order_traversal(node *Node, traversal* [] int){
	if node != nil{
		in_order_traversal(node.left, traversal)	
		*traversal = append(*traversal, node.Val)
		in_order_traversal(node.right, traversal)
	}
}

func (tree *Tree) computeHash() int{
	hash := 1
	var inorder [] int
	in_order_traversal(tree.root, &inorder)
	for i := 0; i < len(inorder); i++ {
		new_value := inorder[i] + 2
		hash = (hash * new_value + new_value) % 4222234741
	}
	return hash
}

func (node *Node) insert(val int){
	if node.Val > val{
		if node.left == nil {
			node.left = &Node{Val: val}
		} else {
			node.left.insert(val)
		}
	} else {
		if node.right == nil {
			node.right = &Node{Val: val}
		} else{
			node.right.insert(val)
		}
	}
}

func main() {
	hashPtr := flag.Int("hash-workers", 0, "a int")
	dataPtr := flag.Int("data-workers", 0, "a int")
	compPtr := flag.Int("comp-workers", 0, "a int")
	inputPtr := flag.String("input", "", "a string")
	flag.Parse()
	fmt.Println("hash:", *hashPtr)
	fmt.Println("data:", *dataPtr)
	fmt.Println("comp:", *compPtr)
	fmt.Println("input:", *inputPtr)
	content, err3 := ioutil.ReadFile(*inputPtr)
	if err3 != nil{
		fmt.Println(err3)
	}
	var trees [] Tree
	var splitTrees[] string = strings.Split(string(content), "\n")
	for i := 0; i < len(splitTrees); i++ {
		var splitTree[] string = strings.Split(splitTrees[i], " ")
		atoiRoot, err := strconv.Atoi(splitTree[0])
		if err != nil { fmt.Println(err) }
		tree := Tree{&Node{Val: atoiRoot}}
		for j := 1; j < len(splitTree); j++ {
			atoi, err2 := strconv.Atoi(splitTree[j])	
			if err2 != nil { fmt.Println(err2) }
			tree.root.insert(atoi)
		}
		trees = append(trees, tree)
	}

	var hashes [] int
	for i := 0; i < len(trees); i++ {
		hash :=  trees[i].computeHash()
		fmt.Printf("%d\n", hash)
		hashes = append(hashes, hash)
	}
}
