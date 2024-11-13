package dag

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// AcyclicGraph is a specialization of Graph that cannot have cycles.
type AcyclicGraph struct {
	Graph
}

// WalkFunc is the callback used for walking the graph.
type WalkFunc func(Vertex) Diagnostics

// DepthWalkFunc is a walk function that also receives the current depth of the
// walk as an argument
type DepthWalkFunc func(Vertex, int) error

func (g *AcyclicGraph) DirectedGraph() Grapher {
	return g
}

// Returns a Set that includes every Vertex yielded by walking down from the
// provided starting Vertex v.
func (g *AcyclicGraph) Ancestors(v Vertex) (Set, error) {
	s := make(Set)
	memoFunc := func(v Vertex, d int) error {
		s.Add(v)
		return nil
	}

	if err := g.DepthFirstWalk(g.downEdgesNoCopy(v), memoFunc); err != nil {
		return nil, err
	}

	return s, nil
}

// Returns a Set that includes every Vertex yielded by walking up from the
// provided starting Vertex v.
func (g *AcyclicGraph) Descendents(v Vertex) (Set, error) {
	s := make(Set)
	memoFunc := func(v Vertex, d int) error {
		s.Add(v)
		return nil
	}

	if err := g.ReverseDepthFirstWalk(g.upEdgesNoCopy(v), memoFunc); err != nil {
		return nil, err
	}

	return s, nil
}

// Root returns the root of the DAG, or an error.
//
// Complexity: O(V)
func (g *AcyclicGraph) Root() (Vertex, error) {
	roots := make([]Vertex, 0, 1)
	for _, v := range g.Vertices() {
		if g.upEdgesNoCopy(v).Len() == 0 {
			roots = append(roots, v)
		}
	}

	if len(roots) > 1 {
		// TODO(mitchellh): make this error message a lot better
		return nil, fmt.Errorf("multiple roots: %#v", roots)
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no roots found")
	}

	return roots[0], nil
}

// TransitiveReduction performs the transitive reduction of graph g in place.
// The transitive reduction of a graph is a graph with as few edges as
// possible with the same reachability as the original graph. This means
// that if there are three nodes A => B => C, and A connects to both
// B and C, and B connects to C, then the transitive reduction is the
// same graph with only a single edge between A and B, and a single edge
// between B and C.
//
// The graph must be free of cycles for this operation to behave properly.
//
// Complexity: O(V(V+E)), or asymptotically O(VE)
func (g *AcyclicGraph) TransitiveReduction() {
	// For each vertex u in graph g, do a DFS starting from each vertex
	// v such that the edge (u,v) exists (v is a direct descendant of u).
	//
	// For each v-prime reachable from v, remove the edge (u, v-prime).
	for _, u := range g.Vertices() {
		uTargets := g.downEdgesNoCopy(u)

		g.DepthFirstWalk(g.downEdgesNoCopy(u), func(v Vertex, d int) error {
			shared := uTargets.Intersection(g.downEdgesNoCopy(v))
			for _, vPrime := range shared {
				g.RemoveEdge(BasicEdge(u, vPrime))
			}

			return nil
		})
	}
}

// Validate validates the DAG. A DAG is valid if it has a single root
// with no cycles.
func (g *AcyclicGraph) Validate() error {
	if _, err := g.Root(); err != nil {
		return err
	}

	// Look for cycles of more than 1 component
	var diags Diagnostics
	cycles := g.Cycles()
	if len(cycles) > 0 {
		for _, cycle := range cycles {
			cycleStr := make([]string, len(cycle))
			for j, vertex := range cycle {
				cycleStr[j] = VertexName(vertex)
			}

			diags = diags.Append(fmt.Errorf(
				"Cycle: %s", strings.Join(cycleStr, ", ")))
		}
	}

	// Look for cycles to self
	for _, e := range g.Edges() {
		if e.Source() == e.Target() {
			diags = diags.Append(fmt.Errorf(
				"Self reference: %s", VertexName(e.Source())))
		}
	}

	return diags.Err()
}

// Cycles reports any cycles between graph nodes.
// Self-referencing nodes are not reported, and must be detected separately.
func (g *AcyclicGraph) Cycles() [][]Vertex {
	var cycles [][]Vertex
	for _, cycle := range StronglyConnected(&g.Graph) {
		if len(cycle) > 1 {
			cycles = append(cycles, cycle)
		}
	}
	return cycles
}

// Walk walks the graph, calling your callback as each node is visited.
// This will walk nodes in parallel if it can. The resulting diagnostics
// contains problems from all graphs visited, in no particular order.
func (g *AcyclicGraph) Walk(cb WalkFunc) Diagnostics {
	w := &Walker{Callback: cb, Reverse: true}
	w.Update(g)
	return w.Wait()
}

// simple convenience helper for converting a dag.Set to a []Vertex
func AsVertexList(s Set) []Vertex {
	vertexList := make([]Vertex, 0, len(s))
	for _, raw := range s {
		vertexList = append(vertexList, raw.(Vertex))
	}
	return vertexList
}

type vertexAtDepth struct {
	Vertex Vertex
	Depth  int
}

// DepthFirstWalk does a depth-first walk of the graph starting from
// the vertices in start.
// The algorithm used here does not do a complete topological sort. To ensure
// correct overall ordering run TransitiveReduction first.
func (g *AcyclicGraph) DepthFirstWalk(start Set, f DepthWalkFunc) error {
	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, 0, len(start))
	for _, v := range start {
		frontier = append(frontier, &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		})
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[current.Vertex]; ok {
			continue
		}
		seen[current.Vertex] = struct{}{}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}

		for _, v := range g.downEdgesNoCopy(current.Vertex) {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: v,
				Depth:  current.Depth + 1,
			})
		}
	}

	return nil
}

// SortedDepthFirstWalk does a depth-first walk of the graph starting from
// the vertices in start, always iterating the nodes in a consistent order.
func (g *AcyclicGraph) SortedDepthFirstWalk(start []Vertex, f DepthWalkFunc) error {
	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, len(start))
	for i, v := range start {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[current.Vertex]; ok {
			continue
		}
		seen[current.Vertex] = struct{}{}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}

		// Visit targets of this in a consistent order.
		targets := AsVertexList(g.downEdgesNoCopy(current.Vertex))
		sort.Sort(byVertexName(targets))

		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}
	}

	return nil
}

// ReverseDepthFirstWalk does a depth-first walk _up_ the graph starting from
// the vertices in start.
// The algorithm used here does not do a complete topological sort. To ensure
// correct overall ordering run TransitiveReduction first.
func (g *AcyclicGraph) ReverseDepthFirstWalk(start Set, f DepthWalkFunc) error {
	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, 0, len(start))
	for _, v := range start {
		frontier = append(frontier, &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		})
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[current.Vertex]; ok {
			continue
		}
		seen[current.Vertex] = struct{}{}

		for _, t := range g.upEdgesNoCopy(current.Vertex) {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}
	}

	return nil
}

// SortedReverseDepthFirstWalk does a depth-first walk _up_ the graph starting from
// the vertices in start, always iterating the nodes in a consistent order.
func (g *AcyclicGraph) SortedReverseDepthFirstWalk(start []Vertex, f DepthWalkFunc) error {
	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, len(start))
	for i, v := range start {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[current.Vertex]; ok {
			continue
		}
		seen[current.Vertex] = struct{}{}

		// Add next set of targets in a consistent order.
		targets := AsVertexList(g.upEdgesNoCopy(current.Vertex))
		sort.Sort(byVertexName(targets))
		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}
	}

	return nil
}

// byVertexName implements sort.Interface so a list of Vertices can be sorted
// consistently by their VertexName
type byVertexName []Vertex

func (b byVertexName) Len() int      { return len(b) }
func (b byVertexName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byVertexName) Less(i, j int) bool {
	return VertexName(b[i]) < VertexName(b[j])
}

type Diagnostic interface {
	Severity() Severity
	Description() Description
}

type Severity rune

//go:generate stringer -type=Severity

const (
	Error   Severity = 'E'
	Warning Severity = 'W'
)

type Description struct {
	Summary string
	Detail  string
}

// Diagnostics is a list of diagnostics. Diagnostics is intended to be used
// where a Go "error" might normally be used, allowing richer information
// to be conveyed (more context, support for warnings).
//
// A nil Diagnostics is a valid, empty diagnostics list, thus allowing
// heap allocation to be avoided in the common case where there are no
// diagnostics to report at all.
type Diagnostics []Diagnostic

// Append is the main interface for constructing Diagnostics lists, taking
// an existing list (which may be nil) and appending the new objects to it
// after normalizing them to be implementations of Diagnostic.
//
// The usual pattern for a function that natively "speaks" diagnostics is:
//
//	// Create a nil Diagnostics at the start of the function
//	var diags diag.Diagnostics
//
//	// At later points, build on it if errors / warnings occur:
//	foo, err := DoSomethingRisky()
//	if err != nil {
//	    diags = diags.Append(err)
//	}
//
//	// Eventually return the result and diagnostics in place of error
//	return result, diags
//
// Append accepts a variety of different diagnostic-like types, including
// native Go errors and HCL diagnostics. It also knows how to unwrap
// a multierror.Error into separate error diagnostics. It can be passed
// another Diagnostics to concatenate the two lists. If given something
// it cannot handle, this function will panic.
func (diags Diagnostics) Append(new ...interface{}) Diagnostics {
	for _, item := range new {
		if item == nil {
			continue
		}

		switch ti := item.(type) {
		case Diagnostic:
			diags = append(diags, ti)
		case Diagnostics:
			diags = append(diags, ti...) // flatten
		case diagnosticsAsError:
			diags = diags.Append(ti.Diagnostics) // unwrap
		case NonFatalError:
			diags = diags.Append(ti.Diagnostics) // unwrap
		case error:
			diags = append(diags, nativeError{ti})
		default:
			panic(fmt.Errorf("can't construct diagnostic(s) from %T", item))
		}
	}

	// Given the above, we should never end up with a non-nil empty slice
	// here, but we'll make sure of that so callers can rely on empty == nil
	if len(diags) == 0 {
		return nil
	}

	return diags
}

// HasErrors returns true if any of the diagnostics in the list have
// a severity of Error.
func (diags Diagnostics) HasErrors() bool {
	for _, diag := range diags {
		if diag.Severity() == Error {
			return true
		}
	}
	return false
}

// Err flattens a diagnostics list into a single Go error, or to nil
// if the diagnostics list does not include any error-level diagnostics.
//
// This can be used to smuggle diagnostics through an API that deals in
// native errors, but unfortunately it will lose naked warnings (warnings
// that aren't accompanied by at least one error) since such APIs have no
// mechanism through which to report these.
//
//	return result, diags.Error()
func (diags Diagnostics) Err() error {
	if !diags.HasErrors() {
		return nil
	}
	return diagnosticsAsError{diags}
}

// ErrWithWarnings is similar to Err except that it will also return a non-nil
// error if the receiver contains only warnings.
//
// In the warnings-only situation, the result is guaranteed to be of dynamic
// type NonFatalError, allowing diagnostics-aware callers to type-assert
// and unwrap it, treating it as non-fatal.
//
// This should be used only in contexts where the caller is able to recognize
// and handle NonFatalError. For normal callers that expect a lack of errors
// to be signaled by nil, use just Diagnostics.Err.
func (diags Diagnostics) ErrWithWarnings() error {
	if len(diags) == 0 {
		return nil
	}
	if diags.HasErrors() {
		return diags.Err()
	}
	return NonFatalError{diags}
}

// NonFatalErr is similar to Err except that it always returns either nil
// (if there are no diagnostics at all) or NonFatalError.
//
// This allows diagnostics to be returned over an error return channel while
// being explicit that the diagnostics should not halt processing.
//
// This should be used only in contexts where the caller is able to recognize
// and handle NonFatalError. For normal callers that expect a lack of errors
// to be signaled by nil, use just Diagnostics.Err.
func (diags Diagnostics) NonFatalErr() error {
	if len(diags) == 0 {
		return nil
	}
	return NonFatalError{diags}
}

// Sort applies an ordering to the diagnostics in the receiver in-place.
//
// The ordering is: warnings before errors, sourceless before sourced,
// short source paths before long source paths, and then ordering by
// position within each file.
//
// Diagnostics that do not differ by any of these sortable characteristics
// will remain in the same relative order after this method returns.
func (diags Diagnostics) Sort() {
	sort.Stable(sortDiagnostics(diags))
}

type diagnosticsAsError struct {
	Diagnostics
}

func (dae diagnosticsAsError) Error() string {
	diags := dae.Diagnostics
	switch {
	case len(diags) == 0:
		// should never happen, since we don't create this wrapper if
		// there are no diagnostics in the list.
		return "no errors"
	case len(diags) == 1:
		desc := diags[0].Description()
		if desc.Detail == "" {
			return desc.Summary
		}
		return fmt.Sprintf("%s: %s", desc.Summary, desc.Detail)
	default:
		var ret bytes.Buffer
		fmt.Fprintf(&ret, "%d problems:\n", len(diags))
		for _, diag := range dae.Diagnostics {
			desc := diag.Description()
			if desc.Detail == "" {
				fmt.Fprintf(&ret, "\n- %s", desc.Summary)
			} else {
				fmt.Fprintf(&ret, "\n- %s: %s", desc.Summary, desc.Detail)
			}
		}
		return ret.String()
	}
}

// WrappedErrors is an implementation of errwrap.Wrapper so that an error-wrapped
// diagnostics object can be picked apart by errwrap-aware code.
func (dae diagnosticsAsError) WrappedErrors() []error {
	var errs []error
	for _, diag := range dae.Diagnostics {
		if wrapper, isErr := diag.(nativeError); isErr {
			errs = append(errs, wrapper.err)
		}
	}
	return errs
}

// NonFatalError is a special error type, returned by
// Diagnostics.ErrWithWarnings and Diagnostics.NonFatalErr,
// that indicates that the wrapped diagnostics should be treated as non-fatal.
// Callers can conditionally type-assert an error to this type in order to
// detect the non-fatal scenario and handle it in a different way.
type NonFatalError struct {
	Diagnostics
}

func (woe NonFatalError) Error() string {
	diags := woe.Diagnostics
	switch {
	case len(diags) == 0:
		// should never happen, since we don't create this wrapper if
		// there are no diagnostics in the list.
		return "no errors or warnings"
	case len(diags) == 1:
		desc := diags[0].Description()
		if desc.Detail == "" {
			return desc.Summary
		}
		return fmt.Sprintf("%s: %s", desc.Summary, desc.Detail)
	default:
		var ret bytes.Buffer
		if diags.HasErrors() {
			fmt.Fprintf(&ret, "%d problems:\n", len(diags))
		} else {
			fmt.Fprintf(&ret, "%d warnings:\n", len(diags))
		}
		for _, diag := range woe.Diagnostics {
			desc := diag.Description()
			if desc.Detail == "" {
				fmt.Fprintf(&ret, "\n- %s", desc.Summary)
			} else {
				fmt.Fprintf(&ret, "\n- %s: %s", desc.Summary, desc.Detail)
			}
		}
		return ret.String()
	}
}

// sortDiagnostics is an implementation of sort.Interface
type sortDiagnostics []Diagnostic

var _ sort.Interface = sortDiagnostics(nil)

func (sd sortDiagnostics) Len() int {
	return len(sd)
}

func (sd sortDiagnostics) Less(i, j int) bool {
	iD, jD := sd[i], sd[j]
	iSev, jSev := iD.Severity(), jD.Severity()

	switch {

	case iSev != jSev:
		return iSev == Warning

	default:
		// The remaining properties do not have a defined ordering, so
		// we'll leave it unspecified. Since we use sort.Stable in
		// the caller of this, the ordering of remaining items will
		// be preserved.
		return false
	}
}

func (sd sortDiagnostics) Swap(i, j int) {
	sd[i], sd[j] = sd[j], sd[i]
}

// DotOpts are the options for generating a dot formatted Graph.
type DotOpts struct {
	// Allows some nodes to decide to only show themselves when the user has
	// requested the "verbose" graph.
	Verbose bool

	// Highlight Cycles
	DrawCycles bool

	// How many levels to expand modules as we draw
	MaxDepth int

	// use this to keep the cluster_ naming convention from the previous dot writer
	cluster bool
}

// GraphNodeDotter can be implemented by a node to cause it to be included
// in the dot graph. The Dot method will be called which is expected to
// return a representation of this node.
type GraphNodeDotter interface {
	// Dot is called to return the dot formatting for the node.
	// The first parameter is the title of the node.
	// The second parameter includes user-specified options that affect the dot
	// graph. See GraphDotOpts below for details.
	DotNode(string, *DotOpts) *DotNode
}

// DotNode provides a structure for Vertices to return in order to specify their
// dot format.
type DotNode struct {
	Name  string
	Attrs map[string]string
}

// Returns the DOT representation of this Graph.
func (g *marshalGraph) Dot(opts *DotOpts) []byte {
	if opts == nil {
		opts = &DotOpts{
			DrawCycles: true,
			MaxDepth:   -1,
			Verbose:    true,
		}
	}

	var w indentWriter
	w.WriteString("digraph {\n")
	w.Indent()

	// some dot defaults
	w.WriteString(`compound = "true"` + "\n")
	w.WriteString(`newrank = "true"` + "\n")

	// the top level graph is written as the first subgraph
	w.WriteString(`subgraph "root" {` + "\n")
	g.writeBody(opts, &w)

	// cluster isn't really used other than for naming purposes in some graphs
	opts.cluster = opts.MaxDepth != 0
	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = -1
	}

	for _, s := range g.Subgraphs {
		g.writeSubgraph(s, opts, maxDepth, &w)
	}

	w.Unindent()
	w.WriteString("}\n")
	return w.Bytes()
}

func (v *marshalVertex) dot(g *marshalGraph, opts *DotOpts) []byte {
	var buf bytes.Buffer
	graphName := g.Name
	if graphName == "" {
		graphName = "root"
	}

	name := v.Name
	attrs := v.Attrs
	if v.graphNodeDotter != nil {
		node := v.graphNodeDotter.DotNode(name, opts)
		if node == nil {
			return []byte{}
		}

		newAttrs := make(map[string]string)
		for k, v := range attrs {
			newAttrs[k] = v
		}
		for k, v := range node.Attrs {
			newAttrs[k] = v
		}

		name = node.Name
		attrs = newAttrs
	}

	buf.WriteString(fmt.Sprintf(`"[%s] %s"`, graphName, name))
	writeAttrs(&buf, attrs)
	buf.WriteByte('\n')

	return buf.Bytes()
}

func (e *marshalEdge) dot(g *marshalGraph) string {
	var buf bytes.Buffer
	graphName := g.Name
	if graphName == "" {
		graphName = "root"
	}

	sourceName := g.vertexByID(e.Source).Name
	targetName := g.vertexByID(e.Target).Name
	s := fmt.Sprintf(`"[%s] %s" -> "[%s] %s"`, graphName, sourceName, graphName, targetName)
	buf.WriteString(s)
	writeAttrs(&buf, e.Attrs)

	return buf.String()
}

func cycleDot(e *marshalEdge, g *marshalGraph) string {
	return e.dot(g) + ` [color = "red", penwidth = "2.0"]`
}

// Write the subgraph body. The is recursive, and the depth argument is used to
// record the current depth of iteration.
func (g *marshalGraph) writeSubgraph(sg *marshalGraph, opts *DotOpts, depth int, w *indentWriter) {
	if depth == 0 {
		return
	}
	depth--

	name := sg.Name
	if opts.cluster {
		// we prefix with cluster_ to match the old dot output
		name = "cluster_" + name
		sg.Attrs["label"] = sg.Name
	}
	w.WriteString(fmt.Sprintf("subgraph %q {\n", name))
	sg.writeBody(opts, w)

	for _, sg := range sg.Subgraphs {
		g.writeSubgraph(sg, opts, depth, w)
	}
}

func (g *marshalGraph) writeBody(opts *DotOpts, w *indentWriter) {
	w.Indent()

	for _, as := range attrStrings(g.Attrs) {
		w.WriteString(as + "\n")
	}

	// list of Vertices that aren't to be included in the dot output
	skip := map[string]bool{}

	for _, v := range g.Vertices {
		if v.graphNodeDotter == nil {
			skip[v.ID] = true
			continue
		}

		w.Write(v.dot(g, opts))
	}

	var dotEdges []string

	if opts.DrawCycles {
		for _, c := range g.Cycles {
			if len(c) < 2 {
				continue
			}

			for i, j := 0, 1; i < len(c); i, j = i+1, j+1 {
				if j >= len(c) {
					j = 0
				}
				src := c[i]
				tgt := c[j]

				if skip[src.ID] || skip[tgt.ID] {
					continue
				}

				e := &marshalEdge{
					Name:   fmt.Sprintf("%s|%s", src.Name, tgt.Name),
					Source: src.ID,
					Target: tgt.ID,
					Attrs:  make(map[string]string),
				}

				dotEdges = append(dotEdges, cycleDot(e, g))
				src = tgt
			}
		}
	}

	for _, e := range g.Edges {
		dotEdges = append(dotEdges, e.dot(g))
	}

	// srot these again to match the old output
	sort.Strings(dotEdges)

	for _, e := range dotEdges {
		w.WriteString(e + "\n")
	}

	w.Unindent()
	w.WriteString("}\n")
}

func writeAttrs(buf *bytes.Buffer, attrs map[string]string) {
	if len(attrs) > 0 {
		buf.WriteString(" [")
		buf.WriteString(strings.Join(attrStrings(attrs), ", "))
		buf.WriteString("]")
	}
}

func attrStrings(attrs map[string]string) []string {
	strings := make([]string, 0, len(attrs))
	for k, v := range attrs {
		strings = append(strings, fmt.Sprintf("%s = %q", k, v))
	}
	sort.Strings(strings)
	return strings
}

// Provide a bytes.Buffer like structure, which will indent when starting a
// newline.
type indentWriter struct {
	bytes.Buffer
	level int
}

func (w *indentWriter) indent() {
	newline := []byte("\n")
	if !bytes.HasSuffix(w.Bytes(), newline) {
		return
	}
	for i := 0; i < w.level; i++ {
		w.Buffer.WriteString("\t")
	}
}

// Indent increases indentation by 1
func (w *indentWriter) Indent() { w.level++ }

// Unindent decreases indentation by 1
func (w *indentWriter) Unindent() { w.level-- }

// the following methods intercecpt the byte.Buffer writes and insert the
// indentation when starting a new line.
func (w *indentWriter) Write(b []byte) (int, error) {
	w.indent()
	return w.Buffer.Write(b)
}

func (w *indentWriter) WriteString(s string) (int, error) {
	w.indent()
	return w.Buffer.WriteString(s)
}
func (w *indentWriter) WriteByte(b byte) error {
	w.indent()
	return w.Buffer.WriteByte(b)
}
func (w *indentWriter) WriteRune(r rune) (int, error) {
	w.indent()
	return w.Buffer.WriteRune(r)
}

// Edge represents an edge in the graph, with a source and target vertex.
type Edge interface {
	Source() Vertex
	Target() Vertex

	Hashable
}

// BasicEdge returns an Edge implementation that simply tracks the source
// and target given as-is.
func BasicEdge(source, target Vertex) Edge {
	return &basicEdge{S: source, T: target}
}

// basicEdge is a basic implementation of Edge that has the source and
// target vertex.
type basicEdge struct {
	S, T Vertex
}

func (e *basicEdge) Hashcode() interface{} {
	return [...]interface{}{e.S, e.T}
}

func (e *basicEdge) Source() Vertex {
	return e.S
}

func (e *basicEdge) Target() Vertex {
	return e.T
}

// nativeError is a Diagnostic implementation that wraps a normal Go error
type nativeError struct {
	err error
}

var _ Diagnostic = nativeError{}

func (e nativeError) Severity() Severity {
	return Error
}

func (e nativeError) Description() Description {
	return Description{
		Summary: e.err.Error(),
	}
}

func (e nativeError) Unwrap() error {
	return e.err
}

// Graph is used to represent a dependency graph.
type Graph struct {
	vertices  Set
	edges     Set
	downEdges map[interface{}]Set
	upEdges   map[interface{}]Set
}

// Subgrapher allows a Vertex to be a Graph itself, by returning a Grapher.
type Subgrapher interface {
	Subgraph() Grapher
}

// A Grapher is any type that returns a Grapher, mainly used to identify
// dag.Graph and dag.AcyclicGraph.  In the case of Graph and AcyclicGraph, they
// return themselves.
type Grapher interface {
	DirectedGraph() Grapher
}

// Vertex of the graph.
type Vertex interface{}

// NamedVertex is an optional interface that can be implemented by Vertex
// to give it a human-friendly name that is used for outputting the graph.
type NamedVertex interface {
	Vertex
	Name() string
}

func (g *Graph) DirectedGraph() Grapher {
	return g
}

// Vertices returns the list of all the vertices in the graph.
func (g *Graph) Vertices() []Vertex {
	result := make([]Vertex, 0, len(g.vertices))
	for _, v := range g.vertices {
		result = append(result, v.(Vertex))
	}

	return result
}

// Edges returns the list of all the edges in the graph.
func (g *Graph) Edges() []Edge {
	result := make([]Edge, 0, len(g.edges))
	for _, v := range g.edges {
		result = append(result, v.(Edge))
	}

	return result
}

// EdgesFrom returns the list of edges from the given source.
func (g *Graph) EdgesFrom(v Vertex) []Edge {
	var result []Edge
	from := hashcode(v)
	for _, e := range g.Edges() {
		if hashcode(e.Source()) == from {
			result = append(result, e)
		}
	}

	return result
}

// EdgesTo returns the list of edges to the given target.
func (g *Graph) EdgesTo(v Vertex) []Edge {
	var result []Edge
	search := hashcode(v)
	for _, e := range g.Edges() {
		if hashcode(e.Target()) == search {
			result = append(result, e)
		}
	}

	return result
}

// HasVertex checks if the given Vertex is present in the graph.
func (g *Graph) HasVertex(v Vertex) bool {
	return g.vertices.Include(v)
}

// HasEdge checks if the given Edge is present in the graph.
func (g *Graph) HasEdge(e Edge) bool {
	return g.edges.Include(e)
}

// Add adds a vertex to the graph. This is safe to call multiple time with
// the same Vertex.
func (g *Graph) Add(v Vertex) Vertex {
	g.init()
	g.vertices.Add(v)
	return v
}

// Remove removes a vertex from the graph. This will also remove any
// edges with this vertex as a source or target.
func (g *Graph) Remove(v Vertex) Vertex {
	// Delete the vertex itself
	g.vertices.Delete(v)

	// Delete the edges to non-existent things
	for _, target := range g.downEdgesNoCopy(v) {
		g.RemoveEdge(BasicEdge(v, target))
	}
	for _, source := range g.upEdgesNoCopy(v) {
		g.RemoveEdge(BasicEdge(source, v))
	}

	return nil
}

// Replace replaces the original Vertex with replacement. If the original
// does not exist within the graph, then false is returned. Otherwise, true
// is returned.
func (g *Graph) Replace(original, replacement Vertex) bool {
	// If we don't have the original, we can't do anything
	if !g.vertices.Include(original) {
		return false
	}

	// If they're the same, then don't do anything
	if original == replacement {
		return true
	}

	// Add our new vertex, then copy all the edges
	g.Add(replacement)
	for _, target := range g.downEdgesNoCopy(original) {
		g.Connect(BasicEdge(replacement, target))
	}
	for _, source := range g.upEdgesNoCopy(original) {
		g.Connect(BasicEdge(source, replacement))
	}

	// Remove our old vertex, which will also remove all the edges
	g.Remove(original)

	return true
}

// RemoveEdge removes an edge from the graph.
func (g *Graph) RemoveEdge(edge Edge) {
	g.init()

	// Delete the edge from the set
	g.edges.Delete(edge)

	// Delete the up/down edges
	if s, ok := g.downEdges[hashcode(edge.Source())]; ok {
		s.Delete(edge.Target())
	}
	if s, ok := g.upEdges[hashcode(edge.Target())]; ok {
		s.Delete(edge.Source())
	}
}

// UpEdges returns the vertices connected to the outward edges from the source
// Vertex v.
func (g *Graph) UpEdges(v Vertex) Set {
	return g.upEdgesNoCopy(v).Copy()
}

// DownEdges returns the vertices connected from the inward edges to Vertex v.
func (g *Graph) DownEdges(v Vertex) Set {
	return g.downEdgesNoCopy(v).Copy()
}

// downEdgesNoCopy returns the outward edges from the source Vertex v as a Set.
// This Set is the same as used internally bu the Graph to prevent a copy, and
// must not be modified by the caller.
func (g *Graph) downEdgesNoCopy(v Vertex) Set {
	g.init()
	return g.downEdges[hashcode(v)]
}

// upEdgesNoCopy returns the inward edges to the destination Vertex v as a Set.
// This Set is the same as used internally bu the Graph to prevent a copy, and
// must not be modified by the caller.
func (g *Graph) upEdgesNoCopy(v Vertex) Set {
	g.init()
	return g.upEdges[hashcode(v)]
}

// Connect adds an edge with the given source and target. This is safe to
// call multiple times with the same value. Note that the same value is
// verified through pointer equality of the vertices, not through the
// value of the edge itself.
func (g *Graph) Connect(edge Edge) {
	g.init()

	source := edge.Source()
	target := edge.Target()
	sourceCode := hashcode(source)
	targetCode := hashcode(target)

	// Do we have this already? If so, don't add it again.
	if s, ok := g.downEdges[sourceCode]; ok && s.Include(target) {
		return
	}

	// Add the edge to the set
	g.edges.Add(edge)

	// Add the down edge
	s, ok := g.downEdges[sourceCode]
	if !ok {
		s = make(Set)
		g.downEdges[sourceCode] = s
	}
	s.Add(target)

	// Add the up edge
	s, ok = g.upEdges[targetCode]
	if !ok {
		s = make(Set)
		g.upEdges[targetCode] = s
	}
	s.Add(source)
}

// String outputs some human-friendly output for the graph structure.
func (g *Graph) StringWithNodeTypes() string {
	var buf bytes.Buffer

	// Build the list of node names and a mapping so that we can more
	// easily alphabetize the output to remain deterministic.
	vertices := g.Vertices()
	names := make([]string, 0, len(vertices))
	mapping := make(map[string]Vertex, len(vertices))
	for _, v := range vertices {
		name := VertexName(v)
		names = append(names, name)
		mapping[name] = v
	}
	sort.Strings(names)

	// Write each node in order...
	for _, name := range names {
		v := mapping[name]
		targets := g.downEdges[hashcode(v)]

		buf.WriteString(fmt.Sprintf("%s - %T\n", name, v))

		// Alphabetize dependencies
		deps := make([]string, 0, targets.Len())
		targetNodes := make(map[string]Vertex)
		for _, target := range targets {
			dep := VertexName(target)
			deps = append(deps, dep)
			targetNodes[dep] = target
		}
		sort.Strings(deps)

		// Write dependencies
		for _, d := range deps {
			buf.WriteString(fmt.Sprintf("  %s - %T\n", d, targetNodes[d]))
		}
	}

	return buf.String()
}

// String outputs some human-friendly output for the graph structure.
func (g *Graph) String() string {
	var buf bytes.Buffer

	// Build the list of node names and a mapping so that we can more
	// easily alphabetize the output to remain deterministic.
	vertices := g.Vertices()
	names := make([]string, 0, len(vertices))
	mapping := make(map[string]Vertex, len(vertices))
	for _, v := range vertices {
		name := VertexName(v)
		names = append(names, name)
		mapping[name] = v
	}
	sort.Strings(names)

	// Write each node in order...
	for _, name := range names {
		v := mapping[name]
		targets := g.downEdges[hashcode(v)]

		buf.WriteString(fmt.Sprintf("%s\n", name))

		// Alphabetize dependencies
		deps := make([]string, 0, targets.Len())
		for _, target := range targets {
			deps = append(deps, VertexName(target))
		}
		sort.Strings(deps)

		// Write dependencies
		for _, d := range deps {
			buf.WriteString(fmt.Sprintf("  %s\n", d))
		}
	}

	return buf.String()
}

func (g *Graph) init() {
	if g.vertices == nil {
		g.vertices = make(Set)
	}
	if g.edges == nil {
		g.edges = make(Set)
	}
	if g.downEdges == nil {
		g.downEdges = make(map[interface{}]Set)
	}
	if g.upEdges == nil {
		g.upEdges = make(map[interface{}]Set)
	}
}

// Dot returns a dot-formatted representation of the Graph.
func (g *Graph) Dot(opts *DotOpts) []byte {
	return newMarshalGraph("", g).Dot(opts)
}

// VertexName returns the name of a vertex.
func VertexName(raw Vertex) string {
	switch v := raw.(type) {
	case NamedVertex:
		return v.Name()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// the marshal* structs are for serialization of the graph data.
type marshalGraph struct {
	// Type is always "Graph", for identification as a top level object in the
	// JSON stream.
	Type string

	// Each marshal structure requires a unique ID so that it can be referenced
	// by other structures.
	ID string `json:",omitempty"`

	// Human readable name for this graph.
	Name string `json:",omitempty"`

	// Arbitrary attributes that can be added to the output.
	Attrs map[string]string `json:",omitempty"`

	// List of graph vertices, sorted by ID.
	Vertices []*marshalVertex `json:",omitempty"`

	// List of edges, sorted by Source ID.
	Edges []*marshalEdge `json:",omitempty"`

	// Any number of subgraphs. A subgraph itself is considered a vertex, and
	// may be referenced by either end of an edge.
	Subgraphs []*marshalGraph `json:",omitempty"`

	// Any lists of vertices that are included in cycles.
	Cycles [][]*marshalVertex `json:",omitempty"`
}

func (g *marshalGraph) vertexByID(id string) *marshalVertex {
	for _, v := range g.Vertices {
		if id == v.ID {
			return v
		}
	}
	return nil
}

type marshalVertex struct {
	// Unique ID, used to reference this vertex from other structures.
	ID string

	// Human readable name
	Name string `json:",omitempty"`

	Attrs map[string]string `json:",omitempty"`

	// This is to help transition from the old Dot interfaces. We record if the
	// node was a GraphNodeDotter here, so we can call it to get attributes.
	graphNodeDotter GraphNodeDotter
}

func newMarshalVertex(v Vertex) *marshalVertex {
	dn, ok := v.(GraphNodeDotter)
	if !ok {
		dn = nil
	}

	// the name will be quoted again later, so we need to ensure it's properly
	// escaped without quotes.
	name := strconv.Quote(VertexName(v))
	name = name[1 : len(name)-1]

	return &marshalVertex{
		ID:              marshalVertexID(v),
		Name:            name,
		Attrs:           make(map[string]string),
		graphNodeDotter: dn,
	}
}

// vertices is a sort.Interface implementation for sorting vertices by ID
type vertices []*marshalVertex

func (v vertices) Less(i, j int) bool { return v[i].Name < v[j].Name }
func (v vertices) Len() int           { return len(v) }
func (v vertices) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

type marshalEdge struct {
	// Human readable name
	Name string

	// Source and Target Vertices by ID
	Source string
	Target string

	Attrs map[string]string `json:",omitempty"`
}

func newMarshalEdge(e Edge) *marshalEdge {
	return &marshalEdge{
		Name:   fmt.Sprintf("%s|%s", VertexName(e.Source()), VertexName(e.Target())),
		Source: marshalVertexID(e.Source()),
		Target: marshalVertexID(e.Target()),
		Attrs:  make(map[string]string),
	}
}

// edges is a sort.Interface implementation for sorting edges by Source ID
type edges []*marshalEdge

func (e edges) Less(i, j int) bool { return e[i].Name < e[j].Name }
func (e edges) Len() int           { return len(e) }
func (e edges) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

// build a marshalGraph structure from a *Graph
func newMarshalGraph(name string, g *Graph) *marshalGraph {
	mg := &marshalGraph{
		Type:  "Graph",
		Name:  name,
		Attrs: make(map[string]string),
	}

	for _, v := range g.Vertices() {
		id := marshalVertexID(v)
		if sg, ok := marshalSubgrapher(v); ok {
			smg := newMarshalGraph(VertexName(v), sg)
			smg.ID = id
			mg.Subgraphs = append(mg.Subgraphs, smg)
		}

		mv := newMarshalVertex(v)
		mg.Vertices = append(mg.Vertices, mv)
	}

	sort.Sort(vertices(mg.Vertices))

	for _, e := range g.Edges() {
		mg.Edges = append(mg.Edges, newMarshalEdge(e))
	}

	sort.Sort(edges(mg.Edges))

	for _, c := range (&AcyclicGraph{*g}).Cycles() {
		var cycle []*marshalVertex
		for _, v := range c {
			mv := newMarshalVertex(v)
			cycle = append(cycle, mv)
		}
		mg.Cycles = append(mg.Cycles, cycle)
	}

	return mg
}

// Attempt to return a unique ID for any vertex.
func marshalVertexID(v Vertex) string {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return strconv.Itoa(int(val.Pointer()))
	case reflect.Interface:
		// A vertex shouldn't contain another layer of interface, but handle
		// this just in case.
		return fmt.Sprintf("%#v", val.Interface())
	}

	if v, ok := v.(Hashable); ok {
		h := v.Hashcode()
		if h, ok := h.(string); ok {
			return h
		}
	}

	// fallback to a name, which we hope is unique.
	return VertexName(v)

	// we could try harder by attempting to read the arbitrary value from the
	// interface, but we shouldn't get here from terraform right now.
}

// check for a Subgrapher, and return the underlying *Graph.
func marshalSubgrapher(v Vertex) (*Graph, bool) {
	sg, ok := v.(Subgrapher)
	if !ok {
		return nil, false
	}

	switch g := sg.Subgraph().DirectedGraph().(type) {
	case *Graph:
		return g, true
	case *AcyclicGraph:
		return &g.Graph, true
	}

	return nil, false
}

// Set is a set data structure.
type Set map[interface{}]interface{}

// Hashable is the interface used by set to get the hash code of a value.
// If this isn't given, then the value of the item being added to the set
// itself is used as the comparison value.
type Hashable interface {
	Hashcode() interface{}
}

// hashcode returns the hashcode used for set elements.
func hashcode(v interface{}) interface{} {
	if h, ok := v.(Hashable); ok {
		return h.Hashcode()
	}

	return v
}

// Add adds an item to the set
func (s Set) Add(v interface{}) {
	s[hashcode(v)] = v
}

// Delete removes an item from the set.
func (s Set) Delete(v interface{}) {
	delete(s, hashcode(v))
}

// Include returns true/false of whether a value is in the set.
func (s Set) Include(v interface{}) bool {
	_, ok := s[hashcode(v)]
	return ok
}

// Intersection computes the set intersection with other.
func (s Set) Intersection(other Set) Set {
	result := make(Set)
	if s == nil || other == nil {
		return result
	}
	// Iteration over a smaller set has better performance.
	if other.Len() < s.Len() {
		s, other = other, s
	}
	for _, v := range s {
		if other.Include(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a set with the elements that s has but
// other doesn't.
func (s Set) Difference(other Set) Set {
	if other == nil || other.Len() == 0 {
		return s.Copy()
	}

	result := make(Set)
	for k, v := range s {
		if _, ok := other[k]; !ok {
			result.Add(v)
		}
	}

	return result
}

// Filter returns a set that contains the elements from the receiver
// where the given callback returns true.
func (s Set) Filter(cb func(interface{}) bool) Set {
	result := make(Set)

	for _, v := range s {
		if cb(v) {
			result.Add(v)
		}
	}

	return result
}

// Len is the number of items in the set.
func (s Set) Len() int {
	return len(s)
}

// List returns the list of set elements.
func (s Set) List() []interface{} {
	if s == nil {
		return nil
	}

	r := make([]interface{}, 0, len(s))
	for _, v := range s {
		r = append(r, v)
	}

	return r
}

// Copy returns a shallow copy of the set.
func (s Set) Copy() Set {
	c := make(Set, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}

const (
	_Severity_name_0 = "Error"
	_Severity_name_1 = "Warning"
)

func (i Severity) String() string {
	switch {
	case i == 69:
		return _Severity_name_0
	case i == 87:
		return _Severity_name_1
	default:
		return "Severity(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

// StronglyConnected returns the list of strongly connected components
// within the Graph g. This information is primarily used by this package
// for cycle detection, but strongly connected components have widespread
// use.
func StronglyConnected(g *Graph) [][]Vertex {
	vs := g.Vertices()
	acct := sccAcct{
		NextIndex:   1,
		VertexIndex: make(map[Vertex]int, len(vs)),
	}
	for _, v := range vs {
		// Recurse on any non-visited nodes
		if acct.VertexIndex[v] == 0 {
			stronglyConnected(&acct, g, v)
		}
	}
	return acct.SCC
}

func stronglyConnected(acct *sccAcct, g *Graph, v Vertex) int {
	// Initial vertex visit
	index := acct.visit(v)
	minIdx := index

	for _, raw := range g.downEdgesNoCopy(v) {
		target := raw.(Vertex)
		targetIdx := acct.VertexIndex[target]

		// Recurse on successor if not yet visited
		if targetIdx == 0 {
			minIdx = min(minIdx, stronglyConnected(acct, g, target))
		} else if acct.inStack(target) {
			// Check if the vertex is in the stack
			minIdx = min(minIdx, targetIdx)
		}
	}

	// Pop the strongly connected components off the stack if
	// this is a root vertex
	if index == minIdx {
		var scc []Vertex
		for {
			v2 := acct.pop()
			scc = append(scc, v2)
			if v2 == v {
				break
			}
		}

		acct.SCC = append(acct.SCC, scc)
	}

	return minIdx
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// sccAcct is used ot pass around accounting information for
// the StronglyConnectedComponents algorithm
type sccAcct struct {
	NextIndex   int
	VertexIndex map[Vertex]int
	Stack       []Vertex
	SCC         [][]Vertex
}

// visit assigns an index and pushes a vertex onto the stack
func (s *sccAcct) visit(v Vertex) int {
	idx := s.NextIndex
	s.VertexIndex[v] = idx
	s.NextIndex++
	s.push(v)
	return idx
}

// push adds a vertex to the stack
func (s *sccAcct) push(n Vertex) {
	s.Stack = append(s.Stack, n)
}

// pop removes a vertex from the stack
func (s *sccAcct) pop() Vertex {
	n := len(s.Stack)
	if n == 0 {
		return nil
	}
	vertex := s.Stack[n-1]
	s.Stack = s.Stack[:n-1]
	return vertex
}

// inStack checks if a vertex is in the stack
func (s *sccAcct) inStack(needle Vertex) bool {
	for _, n := range s.Stack {
		if n == needle {
			return true
		}
	}
	return false
}

// Walker is used to walk every vertex of a graph in parallel.
//
// A vertex will only be walked when the dependencies of that vertex have
// been walked. If two vertices can be walked at the same time, they will be.
//
// Update can be called to update the graph. This can be called even during
// a walk, changing vertices/edges mid-walk. This should be done carefully.
// If a vertex is removed but has already been executed, the result of that
// execution (any error) is still returned by Wait. Changing or re-adding
// a vertex that has already executed has no effect. Changing edges of
// a vertex that has already executed has no effect.
//
// Non-parallelism can be enforced by introducing a lock in your callback
// function. However, the goroutine overhead of a walk will remain.
// Walker will create V*2 goroutines (one for each vertex, and dependency
// waiter for each vertex). In general this should be of no concern unless
// there are a huge number of vertices.
//
// The walk is depth first by default. This can be changed with the Reverse
// option.
//
// A single walker is only valid for one graph walk. After the walk is complete
// you must construct a new walker to walk again. State for the walk is never
// deleted in case vertices or edges are changed.
type Walker struct {
	// Callback is what is called for each vertex
	Callback WalkFunc

	// Reverse, if true, causes the source of an edge to depend on a target.
	// When false (default), the target depends on the source.
	Reverse bool

	// changeLock must be held to modify any of the fields below. Only Update
	// should modify these fields. Modifying them outside of Update can cause
	// serious problems.
	changeLock sync.Mutex
	vertices   Set
	edges      Set
	vertexMap  map[Vertex]*walkerVertex

	// wait is done when all vertices have executed. It may become "undone"
	// if new vertices are added.
	wait sync.WaitGroup

	// diagsMap contains the diagnostics recorded so far for execution,
	// and upstreamFailed contains all the vertices whose problems were
	// caused by upstream failures, and thus whose diagnostics should be
	// excluded from the final set.
	//
	// Readers and writers of either map must hold diagsLock.
	diagsMap       map[Vertex]Diagnostics
	upstreamFailed map[Vertex]struct{}
	diagsLock      sync.Mutex
}

func (w *Walker) init() {
	if w.vertices == nil {
		w.vertices = make(Set)
	}
	if w.edges == nil {
		w.edges = make(Set)
	}
}

type walkerVertex struct {
	// These should only be set once on initialization and never written again.
	// They are not protected by a lock since they don't need to be since
	// they are write-once.

	// DoneCh is closed when this vertex has completed execution, regardless
	// of success.
	//
	// CancelCh is closed when the vertex should cancel execution. If execution
	// is already complete (DoneCh is closed), this has no effect. Otherwise,
	// execution is cancelled as quickly as possible.
	DoneCh   chan struct{}
	CancelCh chan struct{}

	// Dependency information. Any changes to any of these fields requires
	// holding DepsLock.
	//
	// DepsCh is sent a single value that denotes whether the upstream deps
	// were successful (no errors). Any value sent means that the upstream
	// dependencies are complete. No other values will ever be sent again.
	//
	// DepsUpdateCh is closed when there is a new DepsCh set.
	DepsCh       chan bool
	DepsUpdateCh chan struct{}
	DepsLock     sync.Mutex

	// Below is not safe to read/write in parallel. This behavior is
	// enforced by changes only happening in Update. Nothing else should
	// ever modify these.
	deps         map[Vertex]chan struct{}
	depsCancelCh chan struct{}
}

// Wait waits for the completion of the walk and returns diagnostics describing
// any problems that arose. Update should be called to populate the walk with
// vertices and edges prior to calling this.
//
// Wait will return as soon as all currently known vertices are complete.
// If you plan on calling Update with more vertices in the future, you
// should not call Wait until after this is done.
func (w *Walker) Wait() Diagnostics {
	// Wait for completion
	w.wait.Wait()

	var diags Diagnostics
	w.diagsLock.Lock()
	for v, vDiags := range w.diagsMap {
		if _, upstream := w.upstreamFailed[v]; upstream {
			// Ignore diagnostics for nodes that had failed upstreams, since
			// the downstream diagnostics are likely to be redundant.
			continue
		}
		diags = diags.Append(vDiags)
	}
	w.diagsLock.Unlock()

	return diags
}

// Update updates the currently executing walk with the given graph.
// This will perform a diff of the vertices and edges and update the walker.
// Already completed vertices remain completed (including any errors during
// their execution).
//
// This returns immediately once the walker is updated; it does not wait
// for completion of the walk.
//
// Multiple Updates can be called in parallel. Update can be called at any
// time during a walk.
func (w *Walker) Update(g *AcyclicGraph) {
	w.init()
	v := make(Set)
	e := make(Set)
	if g != nil {
		v, e = g.vertices, g.edges
	}

	// Grab the change lock so no more updates happen but also so that
	// no new vertices are executed during this time since we may be
	// removing them.
	w.changeLock.Lock()
	defer w.changeLock.Unlock()

	// Initialize fields
	if w.vertexMap == nil {
		w.vertexMap = make(map[Vertex]*walkerVertex)
	}

	// Calculate all our sets
	newEdges := e.Difference(w.edges)
	oldEdges := w.edges.Difference(e)
	newVerts := v.Difference(w.vertices)
	oldVerts := w.vertices.Difference(v)

	// Add the new vertices
	for _, raw := range newVerts {
		v := raw.(Vertex)

		// Add to the waitgroup so our walk is not done until everything finishes
		w.wait.Add(1)

		// Add to our own set so we know about it already
		w.vertices.Add(raw)

		// Initialize the vertex info
		info := &walkerVertex{
			DoneCh:   make(chan struct{}),
			CancelCh: make(chan struct{}),
			deps:     make(map[Vertex]chan struct{}),
		}

		// Add it to the map and kick off the walk
		w.vertexMap[v] = info
	}

	// Remove the old vertices
	for _, raw := range oldVerts {
		v := raw.(Vertex)

		// Get the vertex info so we can cancel it
		info, ok := w.vertexMap[v]
		if !ok {
			// This vertex for some reason was never in our map. This
			// shouldn't be possible.
			continue
		}

		// Cancel the vertex
		close(info.CancelCh)

		// Delete it out of the map
		delete(w.vertexMap, v)
		w.vertices.Delete(raw)
	}

	// Add the new edges
	changedDeps := make(Set)
	for _, raw := range newEdges {
		edge := raw.(Edge)
		waiter, dep := w.edgeParts(edge)

		// Get the info for the waiter
		waiterInfo, ok := w.vertexMap[waiter]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Get the info for the dep
		depInfo, ok := w.vertexMap[dep]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Add the dependency to our waiter
		waiterInfo.deps[dep] = depInfo.DoneCh

		// Record that the deps changed for this waiter
		changedDeps.Add(waiter)
		w.edges.Add(raw)
	}

	// Process removed edges
	for _, raw := range oldEdges {
		edge := raw.(Edge)
		waiter, dep := w.edgeParts(edge)

		// Get the info for the waiter
		waiterInfo, ok := w.vertexMap[waiter]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Delete the dependency from the waiter
		delete(waiterInfo.deps, dep)

		// Record that the deps changed for this waiter
		changedDeps.Add(waiter)
		w.edges.Delete(raw)
	}

	// For each vertex with changed dependencies, we need to kick off
	// a new waiter and notify the vertex of the changes.
	for _, raw := range changedDeps {
		v := raw.(Vertex)
		info, ok := w.vertexMap[v]
		if !ok {
			// Vertex doesn't exist... shouldn't be possible but ignore.
			continue
		}

		// Create a new done channel
		doneCh := make(chan bool, 1)

		// Create the channel we close for cancellation
		cancelCh := make(chan struct{})

		// Build a new deps copy
		deps := make(map[Vertex]<-chan struct{})
		for k, v := range info.deps {
			deps[k] = v
		}

		// Update the update channel
		info.DepsLock.Lock()
		if info.DepsUpdateCh != nil {
			close(info.DepsUpdateCh)
		}
		info.DepsCh = doneCh
		info.DepsUpdateCh = make(chan struct{})
		info.DepsLock.Unlock()

		// Cancel the older waiter
		if info.depsCancelCh != nil {
			close(info.depsCancelCh)
		}
		info.depsCancelCh = cancelCh

		// Start the waiter
		go w.waitDeps(v, deps, doneCh, cancelCh)
	}

	// Start all the new vertices. We do this at the end so that all
	// the edge waiters and changes are set up above.
	for _, raw := range newVerts {
		v := raw.(Vertex)
		go w.walkVertex(v, w.vertexMap[v])
	}
}

// edgeParts returns the waiter and the dependency, in that order.
// The waiter is waiting on the dependency.
func (w *Walker) edgeParts(e Edge) (Vertex, Vertex) {
	if w.Reverse {
		return e.Source(), e.Target()
	}

	return e.Target(), e.Source()
}

// walkVertex walks a single vertex, waiting for any dependencies before
// executing the callback.
func (w *Walker) walkVertex(v Vertex, info *walkerVertex) {
	// When we're done executing, lower the waitgroup count
	defer w.wait.Done()

	// When we're done, always close our done channel
	defer close(info.DoneCh)

	// Wait for our dependencies. We create a [closed] deps channel so
	// that we can immediately fall through to load our actual DepsCh.
	var depsSuccess bool
	var depsUpdateCh chan struct{}
	depsCh := make(chan bool, 1)
	depsCh <- true
	close(depsCh)
	for {
		select {
		case <-info.CancelCh:
			// Cancel
			return

		case depsSuccess = <-depsCh:
			// Deps complete! Mark as nil to trigger completion handling.
			depsCh = nil

		case <-depsUpdateCh:
			// New deps, reloop
		}

		// Check if we have updated dependencies. This can happen if the
		// dependencies were satisfied exactly prior to an Update occurring.
		// In that case, we'd like to take into account new dependencies
		// if possible.
		info.DepsLock.Lock()
		if info.DepsCh != nil {
			depsCh = info.DepsCh
			info.DepsCh = nil
		}
		if info.DepsUpdateCh != nil {
			depsUpdateCh = info.DepsUpdateCh
		}
		info.DepsLock.Unlock()

		// If we still have no deps channel set, then we're done!
		if depsCh == nil {
			break
		}
	}

	// If we passed dependencies, we just want to check once more that
	// we're not cancelled, since this can happen just as dependencies pass.
	select {
	case <-info.CancelCh:
		// Cancelled during an update while dependencies completed.
		return
	default:
	}

	// Run our callback or note that our upstream failed
	var diags Diagnostics
	var upstreamFailed bool
	if depsSuccess {
		diags = w.Callback(v)
	} else {
		// This won't be displayed to the user because we'll set upstreamFailed,
		// but we need to ensure there's at least one error in here so that
		// the failures will cascade downstream.
		diags = diags.Append(errors.New("upstream dependencies failed"))
		upstreamFailed = true
	}

	// Record the result (we must do this after execution because we mustn't
	// hold diagsLock while visiting a vertex.)
	w.diagsLock.Lock()
	if w.diagsMap == nil {
		w.diagsMap = make(map[Vertex]Diagnostics)
	}
	w.diagsMap[v] = diags
	if w.upstreamFailed == nil {
		w.upstreamFailed = make(map[Vertex]struct{})
	}
	if upstreamFailed {
		w.upstreamFailed[v] = struct{}{}
	}
	w.diagsLock.Unlock()
}

func (w *Walker) waitDeps(
	v Vertex,
	deps map[Vertex]<-chan struct{},
	doneCh chan<- bool,
	cancelCh <-chan struct{}) {

	// For each dependency given to us, wait for it to complete
	for _, depCh := range deps {
	DepSatisfied:
		for {
			select {
			case <-depCh:
				// Dependency satisfied!
				break DepSatisfied

			case <-cancelCh:
				// Wait cancelled. Note that we didn't satisfy dependencies
				// so that anything waiting on us also doesn't run.
				doneCh <- false
				return
			}
		}
	}

	// Dependencies satisfied! We need to check if any errored
	w.diagsLock.Lock()
	defer w.diagsLock.Unlock()
	for dep := range deps {
		if w.diagsMap[dep].HasErrors() {
			// One of our dependencies failed, so return false
			doneCh <- false
			return
		}
	}

	// All dependencies satisfied and successful
	doneCh <- true
}
