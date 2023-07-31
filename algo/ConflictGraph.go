package algo

/*
import (
	"fmt"
	"os"
	"time"
)
*/

type ConflictGraph struct {
	nodes map[int][]int
}

func (c *ConflictGraph) Init() {
	c.nodes = make(map[int][]int)
}

func (c *ConflictGraph) Add(i, j int) {
	/*
	fmt.Fprintf(os.Stderr, "graph before adding %d and %d\n", i, j)
	for key, value := range c.nodes {
		fmt.Fprintf(os.Stderr, "c[%d] : %v\n", key, value)
	}
	*/
	c.nodes[i] = append(c.nodes[i], j)
	c.nodes[j] = append(c.nodes[j], i)
	/*
	fmt.Fprintf(os.Stderr, "graph before after adding %d and %d\n", i, j)
	for key, value := range c.nodes {
		fmt.Fprintf(os.Stderr, "c[%d] : %v\n", key, value)
	}
	*/
}

func (c *ConflictGraph) GetHighestDegreeIndex() (index int) {
	index = -1
	max := -1
	for key, value := range c.nodes {
		if currLen := len(value); currLen > max {
			max, index = currLen, key
		}
	}
	return index
}

func (c *ConflictGraph) PopAndCount() (count int) {
	for ; len(c.nodes) > 0; {
		/*
		fmt.Fprintf(os.Stderr, "graph before poping (len is %d)\n", len(c.nodes))
		for key, value := range c.nodes {
			fmt.Fprintf(os.Stderr, "c[%d] : %v\n", key, value)
		}
		*/
		highest := c.GetHighestDegreeIndex()
		/*
		fmt.Fprintf(os.Stderr, "highest index is %d\n", highest)
		time.Sleep(1000 * time.Millisecond)
		*/
		for key, value := range c.nodes {
			if toDelete := Index(value, highest); toDelete != -1 {
				c.nodes[key] = append(c.nodes[key][:toDelete], c.nodes[key][toDelete+1:]...)
			}
		}
		delete(c.nodes, highest)
		for key := range c.nodes {
			if len(c.nodes[key]) == 0 {
				delete(c.nodes, key)
			}
		}
		count++
	}
	return
}
