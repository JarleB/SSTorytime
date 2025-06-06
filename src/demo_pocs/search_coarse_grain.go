//******************************************************************
//
// Find <end|start> transition matrix and calculate symmetries
//
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

const (
	host     = "localhost"
	port     = 5432
	user     = "sstoryline"
	password = "sst_1234"
	dbname   = "sstoryline"
)

//******************************************************************

func main() {

	load_arrows := true
	ctx := SST.Open(load_arrows)

	// Contra colliding wavefronts as path integral solver

	const maxdepth = 5
	var ldepth,rdepth int = 1,1
	var Lnum,Rnum int
	var count int
	var left_paths, right_paths [][]SST.Link

	start_bc := "A1"
	end_bc := "B6"

	leftptrs := SST.GetDBNodePtrMatchingName(ctx,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(ctx,end_bc,"")

	if leftptrs == nil || rightptrs == nil {
		fmt.Println("No paths available from end points")
		return
	}

	// Find the path matrix

	var solutions [][]SST.Link

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = SST.GetEntireConePathsAsLinks(ctx,"fwd",leftptrs[0],ldepth)
		right_paths,Rnum = SST.GetEntireConePathsAsLinks(ctx,"bwd",rightptrs[0],rdepth)		
		
		solutions,_ = WaveFrontsOverlap(ctx,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {
			fmt.Println("-- T R E E ----------------------------------")
			fmt.Println("Path solution",count,"from",start_bc,"to",end_bc,"with lengths",ldepth,-rdepth)

			for s := 0; s < len(solutions); s++ {
				prefix := fmt.Sprintf(" - story %d: ",s)
				SST.PrintLinkPath(ctx,solutions,s,prefix,"",nil)
			}
			count++
			fmt.Println("-------------------------------------------")
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}

	// Calculate the node layer sets S[path][depth]

	var supernodes [][]SST.NodePtr

	for depth := 0; depth < maxdepth*2; depth++ {

		for p_i := 0; p_i < len(solutions); p_i++ {

			if depth == len(solutions[p_i])-1 {
				supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)
			}

			if depth > len(solutions[p_i])-1 {
				continue
			}

			supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)

			for p_j := p_i+1; p_j < len(solutions); p_j++ {

				if depth < 1 || depth > len(solutions[p_j])-2 {
					break
				}

				if solutions[p_i][depth-1].Dst == solutions[p_j][depth-1].Dst && 
				   solutions[p_i][depth+1].Dst == solutions[p_j][depth+1].Dst {
					   supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_j][depth].Dst)
				}
			}
		}		
	}

	fmt.Println("Look for coarse grains, final matroid:",len(supernodes))

	for g := range supernodes {
		fmt.Print("\n    - Super node ",g," = {")
		for n := range supernodes[g] {
			node :=SST.GetDBNodeByNodePtr(ctx,supernodes[g][n])
			fmt.Print(node.S,",")
		}
		fmt.Println("}")
	}
	
}

// **********************************************************

func WaveFrontsOverlap(ctx SST.PoSST,left_paths,right_paths [][]SST.Link,Lnum,Rnum,ldepth,rdepth int) ([][]SST.Link,[][]SST.Link) {

	// The wave front consists of Lnum and Rnum points left_paths[len()-1].
	// Any of the

	var solutions [][]SST.Link
	var loops [][]SST.Link

	// Start expanding the waves from left and right, one step at a time, alternately

	leftfront := WaveFront(left_paths,Lnum)
	rightfront := WaveFront(right_paths,Rnum)

	fmt.Println("\n  Left front radius",ldepth,":",ShowNode(ctx,leftfront))
	fmt.Println("  Right front radius",rdepth,":",ShowNode(ctx,rightfront))

	incidence := NodesOverlap(ctx,leftfront,rightfront)
	
	for lp := range incidence {

		rp := incidence[lp]

		var LRsplice []SST.Link		

		LRsplice = LeftJoin(LRsplice,left_paths[lp])
		adjoint := SST.AdjointLinkPath(right_paths[rp])
		LRsplice = RightComplementJoin(LRsplice,adjoint)

		fmt.Printf("...SPLICE PATHS L%d with R%d.....\n",lp,rp)
		fmt.Println("Left tendril",ShowNodePath(ctx,left_paths[lp]))
		fmt.Println("Right tendril",ShowNodePath(ctx,right_paths[rp]))
		fmt.Println("Right adjoint:",ShowNodePath(ctx,adjoint))
		fmt.Println(".....................\n")

		if IsDAG(LRsplice) {
			solutions = append(solutions,LRsplice)
		} else {
			loops = append(loops,LRsplice)
		}
	}

	fmt.Printf("  (found %d touching solutions)\n",len(incidence))
	return solutions,loops
}

// **********************************************************

func WaveFront(path [][]SST.Link,num int) []SST.NodePtr {

	// assemble the cross cutting nodeptrs of the wavefronts

	var front []SST.NodePtr

	for l := 0; l < num; l++ {
		front = append(front,path[l][len(path[l])-1].Dst)
	}

	return front
}

// **********************************************************

func NodesOverlap(ctx SST.PoSST,left,right []SST.NodePtr) map[int]int {

	var LRsplice = make(map[int]int)
	var list string

	// Return coordinate pairs of partial paths to splice

	for l := 0; l < len(left); l++ {
		for r := 0; r < len(right); r++ {
			if left[l] == right[r] {
				node := SST.GetDBNodeByNodePtr(ctx,left[l])
				list += node.S+", "
				LRsplice[l] = r
			}
		}
	}

	if len(list) > 0 {
		fmt.Println("  (i.e. waves impinge",len(LRsplice),"times at: ",list,")\n")
	}

	return LRsplice
}

// **********************************************************

func LeftJoin(LRsplice,seq []SST.Link) []SST.Link {

	for i := 0; i < len(seq); i++ {

		LRsplice = append(LRsplice,seq[i])
	}

	return LRsplice
}

// **********************************************************

func RightComplementJoin(LRsplice,adjoint []SST.Link) []SST.Link {

	// len(seq)-1 matches the last node of right join
	// when we invert, links and destinations are shifted

	for j := 1; j < len(adjoint); j++ {
		LRsplice = append(LRsplice,adjoint[j])
	}

	return LRsplice
}

// **********************************************************

func IsDAG(seq []SST.Link) bool {

	var freq = make(map[SST.NodePtr]int)

	for i := range seq {
		freq[seq[i].Dst]++
	}

	for n := range freq {
		if freq[n] > 1 {
			return false
		}
	}

	return true
}

// **********************************************************

func ShowNode(ctx SST.PoSST,nptr []SST.NodePtr) string {

	var ret string

	for n := range nptr {
		node := SST.GetDBNodeByNodePtr(ctx,nptr[n])
		ret += node.S + ","
	}

	return ret
}

// **********************************************************

func ShowNodePath(ctx SST.PoSST,lnk []SST.Link) string {

	var ret string

	for n := range lnk {
		node := SST.GetDBNodeByNodePtr(ctx,lnk[n].Dst)
		arrs := SST.GetDBArrowByPtr(ctx,lnk[n].Arr).Long
		ret += fmt.Sprintf("(%s) -> %s ",arrs,node.S)
	}

	return ret
}

// **********************************************************

func Together(matroid [][]SST.NodePtr,n1 SST.NodePtr,n2 SST.NodePtr) [][]SST.NodePtr {

        // matroid [snode][member]

	if len(matroid) == 0 {
		var newsuper []SST.NodePtr
		newsuper = append(newsuper,n1)
		if n1 != n2 {
			newsuper = append(newsuper,n2)
		}
		matroid = append(matroid,newsuper)
		return matroid
	}

	for i := range matroid {
		if In(matroid[i],n1) || In(matroid[i],n2) {
			matroid[i] = IdempAdd(matroid[i],n1)
			matroid[i] = IdempAdd(matroid[i],n2)
			return matroid
		}
	}

	var newsuper []SST.NodePtr

	newsuper = IdempAdd(newsuper,n1)
	newsuper = IdempAdd(newsuper,n2)
	matroid = append(matroid,newsuper)

	return matroid
}

// **********************************************************

func IdempAdd(set []SST.NodePtr, n SST.NodePtr) []SST.NodePtr {

	if !In(set,n) {
		set = append(set,n)
	}
	return set
}

// **********************************************************

func In(list []SST.NodePtr,node SST.NodePtr) bool {

	for n := range list {
		if list[n] == node {
			return true
		}
	}
	return false
}









