// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/go-tempolite/ent/predicate"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

// WorkflowQuery is the builder for querying Workflow entities.
type WorkflowQuery struct {
	config
	ctx            *QueryContext
	order          []workflow.OrderOption
	inters         []Interceptor
	predicates     []predicate.Workflow
	withExecutions *WorkflowExecutionQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the WorkflowQuery builder.
func (wq *WorkflowQuery) Where(ps ...predicate.Workflow) *WorkflowQuery {
	wq.predicates = append(wq.predicates, ps...)
	return wq
}

// Limit the number of records to be returned by this query.
func (wq *WorkflowQuery) Limit(limit int) *WorkflowQuery {
	wq.ctx.Limit = &limit
	return wq
}

// Offset to start from.
func (wq *WorkflowQuery) Offset(offset int) *WorkflowQuery {
	wq.ctx.Offset = &offset
	return wq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (wq *WorkflowQuery) Unique(unique bool) *WorkflowQuery {
	wq.ctx.Unique = &unique
	return wq
}

// Order specifies how the records should be ordered.
func (wq *WorkflowQuery) Order(o ...workflow.OrderOption) *WorkflowQuery {
	wq.order = append(wq.order, o...)
	return wq
}

// QueryExecutions chains the current query on the "executions" edge.
func (wq *WorkflowQuery) QueryExecutions() *WorkflowExecutionQuery {
	query := (&WorkflowExecutionClient{config: wq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := wq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := wq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(workflow.Table, workflow.FieldID, selector),
			sqlgraph.To(workflowexecution.Table, workflowexecution.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, workflow.ExecutionsTable, workflow.ExecutionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(wq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Workflow entity from the query.
// Returns a *NotFoundError when no Workflow was found.
func (wq *WorkflowQuery) First(ctx context.Context) (*Workflow, error) {
	nodes, err := wq.Limit(1).All(setContextOp(ctx, wq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{workflow.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (wq *WorkflowQuery) FirstX(ctx context.Context) *Workflow {
	node, err := wq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Workflow ID from the query.
// Returns a *NotFoundError when no Workflow ID was found.
func (wq *WorkflowQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = wq.Limit(1).IDs(setContextOp(ctx, wq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{workflow.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (wq *WorkflowQuery) FirstIDX(ctx context.Context) string {
	id, err := wq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Workflow entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Workflow entity is found.
// Returns a *NotFoundError when no Workflow entities are found.
func (wq *WorkflowQuery) Only(ctx context.Context) (*Workflow, error) {
	nodes, err := wq.Limit(2).All(setContextOp(ctx, wq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{workflow.Label}
	default:
		return nil, &NotSingularError{workflow.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (wq *WorkflowQuery) OnlyX(ctx context.Context) *Workflow {
	node, err := wq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Workflow ID in the query.
// Returns a *NotSingularError when more than one Workflow ID is found.
// Returns a *NotFoundError when no entities are found.
func (wq *WorkflowQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = wq.Limit(2).IDs(setContextOp(ctx, wq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{workflow.Label}
	default:
		err = &NotSingularError{workflow.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (wq *WorkflowQuery) OnlyIDX(ctx context.Context) string {
	id, err := wq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Workflows.
func (wq *WorkflowQuery) All(ctx context.Context) ([]*Workflow, error) {
	ctx = setContextOp(ctx, wq.ctx, ent.OpQueryAll)
	if err := wq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Workflow, *WorkflowQuery]()
	return withInterceptors[[]*Workflow](ctx, wq, qr, wq.inters)
}

// AllX is like All, but panics if an error occurs.
func (wq *WorkflowQuery) AllX(ctx context.Context) []*Workflow {
	nodes, err := wq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Workflow IDs.
func (wq *WorkflowQuery) IDs(ctx context.Context) (ids []string, err error) {
	if wq.ctx.Unique == nil && wq.path != nil {
		wq.Unique(true)
	}
	ctx = setContextOp(ctx, wq.ctx, ent.OpQueryIDs)
	if err = wq.Select(workflow.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (wq *WorkflowQuery) IDsX(ctx context.Context) []string {
	ids, err := wq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (wq *WorkflowQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, wq.ctx, ent.OpQueryCount)
	if err := wq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, wq, querierCount[*WorkflowQuery](), wq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (wq *WorkflowQuery) CountX(ctx context.Context) int {
	count, err := wq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (wq *WorkflowQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, wq.ctx, ent.OpQueryExist)
	switch _, err := wq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (wq *WorkflowQuery) ExistX(ctx context.Context) bool {
	exist, err := wq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the WorkflowQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (wq *WorkflowQuery) Clone() *WorkflowQuery {
	if wq == nil {
		return nil
	}
	return &WorkflowQuery{
		config:         wq.config,
		ctx:            wq.ctx.Clone(),
		order:          append([]workflow.OrderOption{}, wq.order...),
		inters:         append([]Interceptor{}, wq.inters...),
		predicates:     append([]predicate.Workflow{}, wq.predicates...),
		withExecutions: wq.withExecutions.Clone(),
		// clone intermediate query.
		sql:  wq.sql.Clone(),
		path: wq.path,
	}
}

// WithExecutions tells the query-builder to eager-load the nodes that are connected to
// the "executions" edge. The optional arguments are used to configure the query builder of the edge.
func (wq *WorkflowQuery) WithExecutions(opts ...func(*WorkflowExecutionQuery)) *WorkflowQuery {
	query := (&WorkflowExecutionClient{config: wq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	wq.withExecutions = query
	return wq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Status workflow.Status `json:"status,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Workflow.Query().
//		GroupBy(workflow.FieldStatus).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (wq *WorkflowQuery) GroupBy(field string, fields ...string) *WorkflowGroupBy {
	wq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &WorkflowGroupBy{build: wq}
	grbuild.flds = &wq.ctx.Fields
	grbuild.label = workflow.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Status workflow.Status `json:"status,omitempty"`
//	}
//
//	client.Workflow.Query().
//		Select(workflow.FieldStatus).
//		Scan(ctx, &v)
func (wq *WorkflowQuery) Select(fields ...string) *WorkflowSelect {
	wq.ctx.Fields = append(wq.ctx.Fields, fields...)
	sbuild := &WorkflowSelect{WorkflowQuery: wq}
	sbuild.label = workflow.Label
	sbuild.flds, sbuild.scan = &wq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a WorkflowSelect configured with the given aggregations.
func (wq *WorkflowQuery) Aggregate(fns ...AggregateFunc) *WorkflowSelect {
	return wq.Select().Aggregate(fns...)
}

func (wq *WorkflowQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range wq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, wq); err != nil {
				return err
			}
		}
	}
	for _, f := range wq.ctx.Fields {
		if !workflow.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if wq.path != nil {
		prev, err := wq.path(ctx)
		if err != nil {
			return err
		}
		wq.sql = prev
	}
	return nil
}

func (wq *WorkflowQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Workflow, error) {
	var (
		nodes       = []*Workflow{}
		_spec       = wq.querySpec()
		loadedTypes = [1]bool{
			wq.withExecutions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Workflow).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Workflow{config: wq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, wq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := wq.withExecutions; query != nil {
		if err := wq.loadExecutions(ctx, query, nodes,
			func(n *Workflow) { n.Edges.Executions = []*WorkflowExecution{} },
			func(n *Workflow, e *WorkflowExecution) { n.Edges.Executions = append(n.Edges.Executions, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (wq *WorkflowQuery) loadExecutions(ctx context.Context, query *WorkflowExecutionQuery, nodes []*Workflow, init func(*Workflow), assign func(*Workflow, *WorkflowExecution)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[string]*Workflow)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.WorkflowExecution(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(workflow.ExecutionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.workflow_executions
		if fk == nil {
			return fmt.Errorf(`foreign-key "workflow_executions" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "workflow_executions" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (wq *WorkflowQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := wq.querySpec()
	_spec.Node.Columns = wq.ctx.Fields
	if len(wq.ctx.Fields) > 0 {
		_spec.Unique = wq.ctx.Unique != nil && *wq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, wq.driver, _spec)
}

func (wq *WorkflowQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(workflow.Table, workflow.Columns, sqlgraph.NewFieldSpec(workflow.FieldID, field.TypeString))
	_spec.From = wq.sql
	if unique := wq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if wq.path != nil {
		_spec.Unique = true
	}
	if fields := wq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, workflow.FieldID)
		for i := range fields {
			if fields[i] != workflow.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := wq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := wq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := wq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := wq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (wq *WorkflowQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(wq.driver.Dialect())
	t1 := builder.Table(workflow.Table)
	columns := wq.ctx.Fields
	if len(columns) == 0 {
		columns = workflow.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if wq.sql != nil {
		selector = wq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if wq.ctx.Unique != nil && *wq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range wq.predicates {
		p(selector)
	}
	for _, p := range wq.order {
		p(selector)
	}
	if offset := wq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := wq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// WorkflowGroupBy is the group-by builder for Workflow entities.
type WorkflowGroupBy struct {
	selector
	build *WorkflowQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (wgb *WorkflowGroupBy) Aggregate(fns ...AggregateFunc) *WorkflowGroupBy {
	wgb.fns = append(wgb.fns, fns...)
	return wgb
}

// Scan applies the selector query and scans the result into the given value.
func (wgb *WorkflowGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, wgb.build.ctx, ent.OpQueryGroupBy)
	if err := wgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*WorkflowQuery, *WorkflowGroupBy](ctx, wgb.build, wgb, wgb.build.inters, v)
}

func (wgb *WorkflowGroupBy) sqlScan(ctx context.Context, root *WorkflowQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(wgb.fns))
	for _, fn := range wgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*wgb.flds)+len(wgb.fns))
		for _, f := range *wgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*wgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := wgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// WorkflowSelect is the builder for selecting fields of Workflow entities.
type WorkflowSelect struct {
	*WorkflowQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ws *WorkflowSelect) Aggregate(fns ...AggregateFunc) *WorkflowSelect {
	ws.fns = append(ws.fns, fns...)
	return ws
}

// Scan applies the selector query and scans the result into the given value.
func (ws *WorkflowSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ws.ctx, ent.OpQuerySelect)
	if err := ws.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*WorkflowQuery, *WorkflowSelect](ctx, ws.WorkflowQuery, ws, ws.inters, v)
}

func (ws *WorkflowSelect) sqlScan(ctx context.Context, root *WorkflowQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ws.fns))
	for _, fn := range ws.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ws.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ws.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
