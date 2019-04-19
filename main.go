package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	. "github.com/kroppt/IntSet"
)

var pstate bool

func init() {
	flag.BoolVar(&pstate, "print", false, "prints DFA construction")
}

func parseEdge(alph Set, trans []map[rune]Set, line string) {
	n := len(trans)
	strs := strings.Split(line, " ")
	if len(strs) != 3 {
		fmt.Fprintf(os.Stderr, "error parsing edge \"%s\"\n", line)
		os.Exit(1)
	}
	n1, err := strconv.Atoi(strs[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing edge index \"%s\"\n", strs[0])
		os.Exit(1)
	}
	if n1 >= n {
		fmt.Fprintf(os.Stderr, "error parsing edge: index %d out of bounds \n", n1)
		os.Exit(1)
	}
	n2, err := strconv.Atoi(strs[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing edge index \"%s\"\n", strs[1])
		os.Exit(1)
	}
	if n2 >= n {
		fmt.Fprintf(os.Stderr, "error parsing edge: index %d out of bounds \n", n2)
		os.Exit(1)
	}
	if len([]rune(strs[2])) != 1 {
		fmt.Fprintf(os.Stderr, "error parsing edge character \"%s\"\n", strs[2])
		os.Exit(1)
	}
	r := []rune(strs[2])[0]
	if _, ok := trans[n1][r]; ok {
		trans[n1][r].Add(n2)
	} else {
		trans[n1][r] = NewSetInit(n2)
	}
	alph.Add(int(r))
}

func εClosure(trans []map[rune]Set, s Set) (ns Set) {
	// iterate until no change
	ns = s.Copy()
	var Δs Set
	iterf := func(i int) bool {
		if s, ok := trans[i]['ε']; ok {
			Δs.Union(s)
		}
		return true
	}
	Δb := true
	for Δb {
		Δs = NewSet()
		ns.Range(iterf)
		Δb = ns.Union(Δs)
	}
	return ns
}

func rmEmpty(strs []string) (out []string) {
	out = make([]string, 0)
	for _, s := range strs {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	flag.Parse()
	// load NFA
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "there must be at least 1 argument for the input file")
		os.Exit(1)
	}
	buf, err := ioutil.ReadFile(args[len(args)-1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fstr := strings.Replace(string(buf), "\r", "", -1)
	// remove carriage returns
	strs := strings.Split(fstr, "\n")
	strs = rmEmpty(strs)
	n, err := strconv.Atoi(strs[0])
	if n < 2 {
		fmt.Fprintln(os.Stderr, "there must be at least 2 nodes")
		os.Exit(1)
	}
	accept := make([]bool, n)
	states := strings.Split(strs[1], " ")
	if len(states) <= 0 {
		fmt.Fprintln(os.Stderr, "there must be at least 1 accepting state")
		os.Exit(1)
	}
	for _, str := range states {
		i, err := strconv.Atoi(str)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading error states")
			os.Exit(1)
		}
		if i >= n {
			fmt.Fprintf(os.Stderr, "accepting state %d is outside bounds\n", i)
			os.Exit(1)
		}
		accept[i] = true
	}
	var trans = make([]map[rune]Set, n)
	for i := range trans {
		trans[i] = make(map[rune]Set)
		trans[i]['ε'] = NewSetInit(i)
	}
	alph := NewSet()
	for _, str := range strs[2:] {
		parseEdge(alph, trans, str)
	}
	// begin algorithm
	initState, _ := trans[0]['ε']
	worklist := []int{0}
	workind := 0
	subset := make([]Set, 1)
	subset[0] = εClosure(trans, initState)
	dfa := map[int](map[rune]Set){}
	if pstate {
		//
	}
	for len(worklist)-workind > 0 {
		d := worklist[workind]
		workind++
		sub := subset[d]
		alph.Range(func(r int) bool {
			if rune(r) == 'ε' {
				return true
			}
			a := NewSet()
			sub.Range(func(s int) bool {
				if t, ok := trans[s][rune(r)]; ok {
					a.Union(t)
				}
				return true
			})
			a = εClosure(trans, a)
			if a.Size() == 0 {
				return true
			}
			nd, ok := func() (Set, bool) {
				for _, m := range dfa {
					for _, v := range m {
						if v.Equals(a) {
							return v, true
						}
					}
				}
				return Set{}, false
			}()
			if !ok {
				nd = a.Copy()
				worklist = append(worklist, len(worklist))
				subset = append(subset, nd)
			}
			if _, ok := dfa[d]; !ok {
				dfa[d] = map[rune]Set{}
			}
			dfa[d][rune(r)] = nd
			return true
		})
	}
	for i, m := range dfa {
		fmt.Printf("state %s\n", subset[i].Print())
		for r, s := range m {
			fmt.Printf("  %s to %s\n", string(r), s.Print())
		}
	}
}
