// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/workflowexecutiondata"
)

// WorkflowExecutionDataQuery is the builder for querying WorkflowExecutionData entities.
type WorkflowExecutionDataQuery struct {
	config
	ctx                   *QueryContext
	order                 []workflowexecutiondata.OrderOption
	inters                []Interceptor
	predicates            []predicate.WorkflowExecutionData
	withWorkflowExecution *WorkflowExecutionQuery
	withFKs               bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the WorkflowExecutionDataQuery builder.
func (wedq *WorkflowExecutionDataQuery) Where(ps ...predicate.WorkflowExecutionData) *WorkflowExecutionDataQuery {
	wedq.predicates = append(wedq.predicates, ps...)
	return wedq
}

// Limit the number of records to be returned by this query.
func (wedq *WorkflowExecutionDataQuery) Limit(limit int) *WorkflowExecutionDataQuery {
	wedq.ctx.Limit = &limit
	return wedq
}

// Offset to start from.
func (wedq *WorkflowExecutionDataQuery) Offset(offset int) *WorkflowExecutionDataQuery {
	wedq.ctx.Offset = &offset
	return wedq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (wedq *WorkflowExecutionDataQuery) Unique(unique bool) *WorkflowExecutionDataQuery {
	wedq.ctx.Unique = &unique
	return wedq
}

// Order specifies how the records should be ordered.
func (wedq *WorkflowExecutionDataQuery) Order(o ...workflowexecutiondata.OrderOption) *WorkflowExecutionDataQuery {
	wedq.order = append(wedq.order, o...)
	return wedq
}

// QueryWorkflowExecution chains the current query on the "workflow_execution" edge.
func (wedq *WorkflowExecutionDataQuery) QueryWorkflowExecution() *WorkflowExecutionQuery {
	query := (&WorkflowExecutionClient{config: wedq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := wedq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := wedq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(workflowexecutiondata.Table, workflowexecutiondata.FieldID, selector),
			sqlgraph.To(workflowexecution.Table, workflowexecution.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, workflowexecutiondata.WorkflowExecutionTable, workflowexecutiondata.WorkflowExecutionColumn),
		)
		fromU = sqlgraph.SetNeighbors(wedq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first WorkflowExecutionData entity from the query.
// Returns a *NotFoundError when no WorkflowExecutionData was found.
func (wedq *WorkflowExecutionDataQuery) First(ctx context.Context) (*WorkflowExecutionData, error) {
	nodes, err := wedq.Limit(1).All(setContextOp(ctx, wedq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{workflowexecutiondata.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) FirstX(ctx context.Context) *WorkflowExecutionData {
	node, err := wedq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first WorkflowExecutionData ID from the query.
// Returns a *NotFoundError when no WorkflowExecutionData ID was found.
func (wedq *WorkflowExecutionDataQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = wedq.Limit(1).IDs(setContextOp(ctx, wedq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{workflowexecutiondata.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) FirstIDX(ctx context.Context) int {
	id, err := wedq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single WorkflowExecutionData entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one WorkflowExecutionData entity is found.
// Returns a *NotFoundError when no WorkflowExecutionData entities are found.
func (wedq *WorkflowExecutionDataQuery) Only(ctx context.Context) (*WorkflowExecutionData, error) {
	nodes, err := wedq.Limit(2).All(setContextOp(ctx, wedq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{workflowexecutiondata.Label}
	default:
		return nil, &NotSingularError{workflowexecutiondata.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) OnlyX(ctx context.Context) *WorkflowExecutionData {
	node, err := wedq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only WorkflowExecutionData ID in the query.
// Returns a *NotSingularError when more than one WorkflowExecutionData ID is found.
// Returns a *NotFoundError when no entities are found.
func (wedq *WorkflowExecutionDataQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = wedq.Limit(2).IDs(setContextOp(ctx, wedq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{workflowexecutiondata.Label}
	default:
		err = &NotSingularError{workflowexecutiondata.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) OnlyIDX(ctx context.Context) int {
	id, err := wedq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of WorkflowExecutionDataSlice.
func (wedq *WorkflowExecutionDataQuery) All(ctx context.Context) ([]*WorkflowExecutionData, error) {
	ctx = setContextOp(ctx, wedq.ctx, ent.OpQueryAll)
	if err := wedq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*WorkflowExecutionData, *WorkflowExecutionDataQuery]()
	return withInterceptors[[]*WorkflowExecutionData](ctx, wedq, qr, wedq.inters)
}

// AllX is like All, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) AllX(ctx context.Context) []*WorkflowExecutionData {
	nodes, err := wedq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of WorkflowExecutionData IDs.
func (wedq *WorkflowExecutionDataQuery) IDs(ctx context.Context) (ids []int, err error) {
	if wedq.ctx.Unique == nil && wedq.path != nil {
		wedq.Unique(true)
	}
	ctx = setContextOp(ctx, wedq.ctx, ent.OpQueryIDs)
	if err = wedq.Select(workflowexecutiondata.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) IDsX(ctx context.Context) []int {
	ids, err := wedq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (wedq *WorkflowExecutionDataQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, wedq.ctx, ent.OpQueryCount)
	if err := wedq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, wedq, querierCount[*WorkflowExecutionDataQuery](), wedq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) CountX(ctx context.Context) int {
	count, err := wedq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (wedq *WorkflowExecutionDataQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, wedq.ctx, ent.OpQueryExist)
	switch _, err := wedq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (wedq *WorkflowExecutionDataQuery) ExistX(ctx context.Context) bool {
	exist, err := wedq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the WorkflowExecutionDataQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (wedq *WorkflowExecutionDataQuery) Clone() *WorkflowExecutionDataQuery {
	if wedq == nil {
		return nil
	}
	return &WorkflowExecutionDataQuery{
		config:                wedq.config,
		ctx:                   wedq.ctx.Clone(),
		order:                 append([]workflowexecutiondata.OrderOption{}, wedq.order...),
		inters:                append([]Interceptor{}, wedq.inters...),
		predicates:            append([]predicate.WorkflowExecutionData{}, wedq.predicates...),
		withWorkflowExecution: wedq.withWorkflowExecution.Clone(),
		// clone intermediate query.
		sql:  wedq.sql.Clone(),
		path: wedq.path,
	}
}

// WithWorkflowExecution tells the query-builder to eager-load the nodes that are connected to
// the "workflow_execution" edge. The optional arguments are used to configure the query builder of the edge.
func (wedq *WorkflowExecutionDataQuery) WithWorkflowExecution(opts ...func(*WorkflowExecutionQuery)) *WorkflowExecutionDataQuery {
	query := (&WorkflowExecutionClient{config: wedq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	wedq.withWorkflowExecution = query
	return wedq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Checkpoints [][]uint8 `json:"checkpoints,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.WorkflowExecutionData.Query().
//		GroupBy(workflowexecutiondata.FieldCheckpoints).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (wedq *WorkflowExecutionDataQuery) GroupBy(field string, fields ...string) *WorkflowExecutionDataGroupBy {
	wedq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &WorkflowExecutionDataGroupBy{build: wedq}
	grbuild.flds = &wedq.ctx.Fields
	grbuild.label = workflowexecutiondata.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Checkpoints [][]uint8 `json:"checkpoints,omitempty"`
//	}
//
//	client.WorkflowExecutionData.Query().
//		Select(workflowexecutiondata.FieldCheckpoints).
//		Scan(ctx, &v)
func (wedq *WorkflowExecutionDataQuery) Select(fields ...string) *WorkflowExecutionDataSelect {
	wedq.ctx.Fields = append(wedq.ctx.Fields, fields...)
	sbuild := &WorkflowExecutionDataSelect{WorkflowExecutionDataQuery: wedq}
	sbuild.label = workflowexecutiondata.Label
	sbuild.flds, sbuild.scan = &wedq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a WorkflowExecutionDataSelect configured with the given aggregations.
func (wedq *WorkflowExecutionDataQuery) Aggregate(fns ...AggregateFunc) *WorkflowExecutionDataSelect {
	return wedq.Select().Aggregate(fns...)
}

func (wedq *WorkflowExecutionDataQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range wedq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, wedq); err != nil {
				return err
			}
		}
	}
	for _, f := range wedq.ctx.Fields {
		if !workflowexecutiondata.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if wedq.path != nil {
		prev, err := wedq.path(ctx)
		if err != nil {
			return err
		}
		wedq.sql = prev
	}
	return nil
}

func (wedq *WorkflowExecutionDataQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*WorkflowExecutionData, error) {
	var (
		nodes       = []*WorkflowExecutionData{}
		withFKs     = wedq.withFKs
		_spec       = wedq.querySpec()
		loadedTypes = [1]bool{
			wedq.withWorkflowExecution != nil,
		}
	)
	if wedq.withWorkflowExecution != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, workflowexecutiondata.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*WorkflowExecutionData).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &WorkflowExecutionData{config: wedq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, wedq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := wedq.withWorkflowExecution; query != nil {
		if err := wedq.loadWorkflowExecution(ctx, query, nodes, nil,
			func(n *WorkflowExecutionData, e *WorkflowExecution) { n.Edges.WorkflowExecution = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (wedq *WorkflowExecutionDataQuery) loadWorkflowExecution(ctx context.Context, query *WorkflowExecutionQuery, nodes []*WorkflowExecutionData, init func(*WorkflowExecutionData), assign func(*WorkflowExecutionData, *WorkflowExecution)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*WorkflowExecutionData)
	for i := range nodes {
		if nodes[i].workflow_execution_execution_data == nil {
			continue
		}
		fk := *nodes[i].workflow_execution_execution_data
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(workflowexecution.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "workflow_execution_execution_data" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (wedq *WorkflowExecutionDataQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := wedq.querySpec()
	_spec.Node.Columns = wedq.ctx.Fields
	if len(wedq.ctx.Fields) > 0 {
		_spec.Unique = wedq.ctx.Unique != nil && *wedq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, wedq.driver, _spec)
}

func (wedq *WorkflowExecutionDataQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(workflowexecutiondata.Table, workflowexecutiondata.Columns, sqlgraph.NewFieldSpec(workflowexecutiondata.FieldID, field.TypeInt))
	_spec.From = wedq.sql
	if unique := wedq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if wedq.path != nil {
		_spec.Unique = true
	}
	if fields := wedq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, workflowexecutiondata.FieldID)
		for i := range fields {
			if fields[i] != workflowexecutiondata.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := wedq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := wedq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := wedq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := wedq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (wedq *WorkflowExecutionDataQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(wedq.driver.Dialect())
	t1 := builder.Table(workflowexecutiondata.Table)
	columns := wedq.ctx.Fields
	if len(columns) == 0 {
		columns = workflowexecutiondata.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if wedq.sql != nil {
		selector = wedq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if wedq.ctx.Unique != nil && *wedq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range wedq.predicates {
		p(selector)
	}
	for _, p := range wedq.order {
		p(selector)
	}
	if offset := wedq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := wedq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// WorkflowExecutionDataGroupBy is the group-by builder for WorkflowExecutionData entities.
type WorkflowExecutionDataGroupBy struct {
	selector
	build *WorkflowExecutionDataQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (wedgb *WorkflowExecutionDataGroupBy) Aggregate(fns ...AggregateFunc) *WorkflowExecutionDataGroupBy {
	wedgb.fns = append(wedgb.fns, fns...)
	return wedgb
}

// Scan applies the selector query and scans the result into the given value.
func (wedgb *WorkflowExecutionDataGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, wedgb.build.ctx, ent.OpQueryGroupBy)
	if err := wedgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*WorkflowExecutionDataQuery, *WorkflowExecutionDataGroupBy](ctx, wedgb.build, wedgb, wedgb.build.inters, v)
}

func (wedgb *WorkflowExecutionDataGroupBy) sqlScan(ctx context.Context, root *WorkflowExecutionDataQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(wedgb.fns))
	for _, fn := range wedgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*wedgb.flds)+len(wedgb.fns))
		for _, f := range *wedgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*wedgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := wedgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// WorkflowExecutionDataSelect is the builder for selecting fields of WorkflowExecutionData entities.
type WorkflowExecutionDataSelect struct {
	*WorkflowExecutionDataQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (weds *WorkflowExecutionDataSelect) Aggregate(fns ...AggregateFunc) *WorkflowExecutionDataSelect {
	weds.fns = append(weds.fns, fns...)
	return weds
}

// Scan applies the selector query and scans the result into the given value.
func (weds *WorkflowExecutionDataSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, weds.ctx, ent.OpQuerySelect)
	if err := weds.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*WorkflowExecutionDataQuery, *WorkflowExecutionDataSelect](ctx, weds.WorkflowExecutionDataQuery, weds, weds.inters, v)
}

func (weds *WorkflowExecutionDataSelect) sqlScan(ctx context.Context, root *WorkflowExecutionDataQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(weds.fns))
	for _, fn := range weds.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*weds.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := weds.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
