package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"sync"
)

type hashId struct{
	hash int
	id int
}

var dataWorkers int
var mutex = &sync.Mutex{}
var hashComparision int
var hashPtr *int
var dataPtr *int
var compPtr *int
var h2i = make(map[int][]int)		

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


func sequentialComputeHash(tree *Tree) int{
	hash := 1
	var inorder [] int
	in_order_traversal(tree.root, &inorder)
	for i := 0; i < len(inorder); i++ {
		new_value := inorder[i] + 2
		hash = (hash * new_value + new_value) % 4222234741
	}
	return hash
}

func parallelComputeHash(trees[] Tree, l int, r int, c chan hashId, wg *sync.WaitGroup){
	for i:= l; i < r; i++ {
		hash := 1
		var inorder [] int
		in_order_traversal(trees[i].root, &inorder)
		for j := 0; j < len(inorder); j++ {
			new_value := inorder[j] + 2
			hash = (hash * new_value + new_value) % 4222234741
		}
		if dataWorkers == 1 {
			treeInMap := hashId{hash, i}
			c <- treeInMap
		} else if dataWorkers == hashComparision{
			mutex.Lock()
			h2i[hash] = append(h2i[hash], i)
			mutex.Unlock()
		}
	}
	wg.Done()
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

func mapInsert(c chan hashId, race chan int, lenTree int){
	for pair:=range c {
		h2i[pair.hash] = append(h2i[pair.hash], pair.id)
	}
}

func main() {
	hashPtr := flag.Int("hash-workers", 0, "a int")
	dataPtr := flag.Int("data-workers", 0, "a int")
	compPtr := flag.Int("comp-workers", 0, "a int")
	inputPtr := flag.String("input", "", "a string")
	flag.Parse()
	fmt.Println("comp:", *compPtr)
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
	var hashWorkers = *hashPtr
	hashComparision = *hashPtr
	dataWorkers = *dataPtr
	c := make(chan hashId)
	race := make(chan int)
	if hashWorkers != 1 && *dataPtr == 1 {
		go mapInsert(c, race, len(splitTrees))
	}
	if *hashPtr == 1 {
		i := 0
		for i < len(trees){
			hash := sequentialComputeHash(&trees[i])
			h2i[hash] = append(h2i[hash], i)
			i += 1
		}
	} else {
		if hashWorkers >= len(trees){
			var wg sync.WaitGroup	
			for j := 0; j < hashWorkers; j++ {
				wg.Add(1)
				if j >= len(trees) {
					go parallelComputeHash(trees, 1, 0,  c, &wg)
				} else {
					go parallelComputeHash(trees, j, j+1, c, &wg)
				}
			}
			wg.Wait()
		} else {
			index := 0
			overflow := len(trees) % hashWorkers
			perfect := int(len(trees) / hashWorkers)
			var wg sync.WaitGroup	
			for j:= 0; j < hashWorkers; j++ {
				wg.Add(1)
				if overflow > 0 {
					oldIndex := index
					index += (perfect + 1)
					go parallelComputeHash(trees, oldIndex, index, c, &wg)
					overflow--
				} else {
					oldIndex := index
					index += perfect
					go parallelComputeHash(trees, oldIndex, index, c, &wg)
				}
			}
			wg.Wait()
		}
	}
	if hashWorkers != 1 && *dataPtr == 1{
		close(c)
	}
	fmt.Println(h2i)
}