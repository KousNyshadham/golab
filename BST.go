package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"sync"
	"time"
)

type cst struct{
	x Tree 
	y Tree
	xin int
	yin int
}

type boundedBuffer struct{
	n int
	head int
	N int	
	buffer[] cst	
	// dont have to declare below?
	bbmutex sync.Mutex
	empty *sync.Cond
	full *sync.Cond
}

func NewBoundedBuffer(n int, head int, N int) *boundedBuffer{
	bb := boundedBuffer{n: n, head: head, N: N}
	bb.empty = sync.NewCond(&bb.bbmutex)
	bb.full = sync.NewCond(&bb.bbmutex)
	bb.buffer = make([] cst, N)
	return &bb
}

func (bb *boundedBuffer) get() cst{
	var value cst
	bb.bbmutex.Lock()
	for bb.n == 0 {
		bb.full.Wait()
	}
	value = bb.buffer[bb.head]
	bb.head = (bb.head + 1) % bb.N;
	bb.n = bb.n-1
	bb.empty.Signal()
	bb.bbmutex.Unlock()
	return value
}

func consumeBuffer(){
	for{
		output := globalbb.get()
		if output.xin == -69 {
			globalwgg.Done()
			return
		}
		compare(output.x, output.y, output.xin,output.yin)
	}
}

func (bb *boundedBuffer) producer(inputCst cst){
	bb.bbmutex.Lock()
	for bb.n == bb.N {
		bb.empty.Wait()
	}
	bb.buffer[(bb.head+bb.n)%bb.N] = inputCst
	bb.n = bb.n+1
	bb.full.Signal()
	bb.bbmutex.Unlock()
}

type hashId struct{
	hash int
	id int
}

var dataWorkers int
var mutex = &sync.Mutex{}
var hashComparision int
var hashPtr *int
var dataPtr *int
var globalbb *boundedBuffer
var compPtr *int
var h2i = make(map[int][]int)		
var globalAdjacencyMatrix [][] bool
var globalwgg sync.WaitGroup	

type Tree struct{
	root *Node
}

type Node struct{
	Val int
	left *Node
	right *Node
}


func in_order_channel(node *Node, ch chan int) {
	if node != nil {
		in_order_channel(node.left, ch)
		ch <- node.Val
		in_order_channel(node.right, ch)
	}
}

func createChannel(x Tree, c chan int){
	in_order_channel(x.root, c)
	close(c)
}

func compare(x Tree, y Tree, xin int, yin int) {
	xc := make(chan int)
	yc := make(chan int)
	go createChannel(x, xc)
	go createChannel(y, yc)
	for {
		xv, okx := <- xc
		yv, oky := <- yc
		if !okx || !oky{
			if(okx == oky){
				globalAdjacencyMatrix[xin][yin] = true
				globalAdjacencyMatrix[yin][xin] = true
			}
			return
		}
		if xv != yv {
			return	
		}
	}
}

func parallelCompare(x Tree, y Tree, xin int, yin int, wg *sync.WaitGroup) {
	xc := make(chan int)
	yc := make(chan int)
	go createChannel(x, xc)
	go createChannel(y, yc)
	for {
		xv, okx := <- xc
		yv, oky := <- yc
		if !okx || !oky{
			if(okx == oky){
				globalAdjacencyMatrix[xin][yin] = true
				globalAdjacencyMatrix[yin][xin] = true
			}
			wg.Done()
			return
		}
		if xv != yv {
			wg.Done()
			return	
		}
	}
}

func in_order_traversal(node *Node, traversal* [] int){
	if node != nil{
		in_order_traversal(node.left, traversal)	
		*traversal = append(*traversal, node.Val)
		in_order_traversal(node.right, traversal)
	}
}

func sequentialComputeHash(tree *Tree, id int, c chan hashId){
	hash := 1
	var inorder [] int
	in_order_traversal(tree.root, &inorder)
	for i := 0; i < len(inorder); i++ {
		new_value := inorder[i] + 2
		hash = (hash * new_value + new_value) % 4222234741
	}
	if dataWorkers == 1{
		treeInMap := hashId{hash, id}
		c <- treeInMap
	}
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
	content, err3 := ioutil.ReadFile(*inputPtr)
	if err3 != nil{
		fmt.Println(err3)
	}
	var trees [] Tree
	var splitTrees[] string = strings.Split(string(content), "\n")
	for i := 0; i < len(splitTrees); i++ {
		var splitTree[] string = strings.Split(splitTrees[i], " ")
		atoiRoot, err := strconv.Atoi(splitTree[0])
		if err != nil { fmt.Println(err); continue}
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
	if (*dataPtr == 1){
		go mapInsert(c, race, len(splitTrees))
	}
	if *hashPtr == 1 {
		start := time.Now()
		i := 0
		for i < len(trees){
			sequentialComputeHash(&trees[i], i, c)
			i += 1
		}
		elapsed:=time.Since(start)
		fmt.Println("hashTime: ", elapsed)
		if dataWorkers != 0{
			fmt.Println("hashGroupTime: ", elapsed)
		}
	} else {
		if hashWorkers >= len(trees){
			var wg sync.WaitGroup	
			bruh := time.Now()
			for j := 0; j < hashWorkers; j++ {
				wg.Add(1)
				if j >= len(trees) {
					go parallelComputeHash(trees, 1, 0,  c, &wg)
				} else {
					go parallelComputeHash(trees, j, j+1, c, &wg)
				}
			}
			wg.Wait()
			elapsed:=time.Since(bruh)
			fmt.Println("hashTime: ", elapsed)
			if dataWorkers != 0{
				fmt.Println("hashGroupTime: ", elapsed)
			}
		} else {
			index := 0
			overflow := len(trees) % hashWorkers
			perfect := int(len(trees) / hashWorkers)
			var wg sync.WaitGroup	
			bruh := time.Now()
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
			elapsed:=time.Since(bruh)
			fmt.Println("hashTime: ", elapsed)
			if dataWorkers != 0{
				fmt.Println("hashGroupTime: ", elapsed)
			}
		}
	}
	if *dataPtr == 1{
		close(c)
	}
	if dataWorkers == 0 {
		return
	}
	for key := range h2i {
		sameHashes := h2i[key]
		if len(sameHashes) == 1{
			continue
		}
		fmt.Printf("%d: ",key)
		for bruhh := range sameHashes{
			bruhh = bruhh
			fmt.Printf("%d ", sameHashes[bruhh])
		}
		fmt.Println()
	}
	//part 3
	compWorkers := * compPtr
	if compWorkers == 0 {
		return
	}
	s := len(trees)
	adjacencyMatrix := make([][] bool, s)
	for r := range adjacencyMatrix {
		adjacencyMatrix[r] = make([] bool, s)
		for c := range adjacencyMatrix[r] {
			adjacencyMatrix[r][c] = false
		}
	}
	globalAdjacencyMatrix = adjacencyMatrix
    //sequential: no compWorkers spawned; different from one comp worker being spawned
	if compWorkers == -69{
		//var wg sync.WaitGroup	
		start := time.Now()
		for key := range h2i {
			sameHashes := h2i[key]
			for o := 0; o < len(sameHashes); o++ {
				for i:= o; i < len(sameHashes); i++ {
					compare(trees[sameHashes[o]], trees[sameHashes[i]], sameHashes[o], sameHashes[i])
					/*
					wg.Add(1)
					go parallelCompare(trees[sameHashes[o]], trees[sameHashes[i]], sameHashes[o], sameHashes[i], &wg)
					*/
				}
			}
		}
		elapsed := time.Since(start)
		fmt.Println("compareTreeTime: ", elapsed)
		//wg.Wait()
	} else {
		start := time.Now()
		globalbb = NewBoundedBuffer(0, 0, compWorkers)
		for f := 0; f < compWorkers; f++ {
			globalwgg.Add(1)
			go consumeBuffer()
		}
		for key := range h2i{
			sameHashes := h2i[key]
			for o:=0; o < len(sameHashes); o++ {
				for i:= o; i < len(sameHashes); i++{
					sendCst := cst{trees[sameHashes[o]], trees[sameHashes[i]], sameHashes[o], sameHashes[i]}
					globalbb.producer(sendCst)
				}
			}
		}
		//sending poisoned values to go routines waiting to receive to exit them from infite for loop and wait so print normal 
		for f:= 0; f < compWorkers; f++ {
			sendCst := cst{Tree{&Node{Val: -69}}, Tree{&Node{Val: -69}}, -69, -69}
			globalbb.producer(sendCst)
		}
		globalwgg.Wait()
		elapsed := time.Since(start)
		fmt.Println("compareTreeTime: ", elapsed)
	}
	set := make(map[int]bool)
	for f:=0; f < len(trees); f++ {
		set[f] = true
	}
	groups := 0
	for f:=0; f < len(trees); f++{
		if set[f] == false{
			continue
		}
		count := 0
		for brooo := range globalAdjacencyMatrix[f] {
			if globalAdjacencyMatrix[f][brooo] {
				count = count + 1
			}
		}
		if count <= 1{
			continue
		} else{
			fmt.Printf("group %d: ", groups)
			for brooo := range globalAdjacencyMatrix[f] {
				if globalAdjacencyMatrix[f][brooo] {
					set[brooo] = false			
					fmt.Printf("%d ", brooo)
				}
			}
			fmt.Println()
			groups = groups + 1
		}
	}
}
