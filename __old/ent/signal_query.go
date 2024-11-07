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
	"github.com/davidroman0O/tempolite/ent/predicate"
	"github.com/davidroman0O/tempolite/ent/signal"
	"github.com/davidroman0O/tempolite/ent/signalexecution"
)

// SignalQuery is the builder for querying Signal entities.
type SignalQuery struct {
	config
	ctx            *QueryContext
	order          []signal.OrderOption
	inters         []Interceptor
	predicates     []predicate.Signal
	withExecutions *SignalExecutionQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SignalQuery builder.
func (sq *SignalQuery) Where(ps ...predicate.Signal) *SignalQuery {
	sq.predicates = append(sq.predicates, ps...)
	return sq
}

// Limit the number of records to be returned by this query.
func (sq *SignalQuery) Limit(limit int) *SignalQuery {
	sq.ctx.Limit = &limit
	return sq
}

// Offset to start from.
func (sq *SignalQuery) Offset(offset int) *SignalQuery {
	sq.ctx.Offset = &offset
	return sq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (sq *SignalQuery) Unique(unique bool) *SignalQuery {
	sq.ctx.Unique = &unique
	return sq
}

// Order specifies how the records should be ordered.
func (sq *SignalQuery) Order(o ...signal.OrderOption) *SignalQuery {
	sq.order = append(sq.order, o...)
	return sq
}

// QueryExecutions chains the current query on the "executions" edge.
func (sq *SignalQuery) QueryExecutions() *SignalExecutionQuery {
	query := (&SignalExecutionClient{config: sq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := sq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := sq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(signal.Table, signal.FieldID, selector),
			sqlgraph.To(signalexecution.Table, signalexecution.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, signal.ExecutionsTable, signal.ExecutionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(sq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Signal entity from the query.
// Returns a *NotFoundError when no Signal was found.
func (sq *SignalQuery) First(ctx context.Context) (*Signal, error) {
	nodes, err := sq.Limit(1).All(setContextOp(ctx, sq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{signal.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (sq *SignalQuery) FirstX(ctx context.Context) *Signal {
	node, err := sq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Signal ID from the query.
// Returns a *NotFoundError when no Signal ID was found.
func (sq *SignalQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sq.Limit(1).IDs(setContextOp(ctx, sq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{signal.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (sq *SignalQuery) FirstIDX(ctx context.Context) string {
	id, err := sq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Signal entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Signal entity is found.
// Returns a *NotFoundError when no Signal entities are found.
func (sq *SignalQuery) Only(ctx context.Context) (*Signal, error) {
	nodes, err := sq.Limit(2).All(setContextOp(ctx, sq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{signal.Label}
	default:
		return nil, &NotSingularError{signal.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (sq *SignalQuery) OnlyX(ctx context.Context) *Signal {
	node, err := sq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Signal ID in the query.
// Returns a *NotSingularError when more than one Signal ID is found.
// Returns a *NotFoundError when no entities are found.
func (sq *SignalQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sq.Limit(2).IDs(setContextOp(ctx, sq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{signal.Label}
	default:
		err = &NotSingularError{signal.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (sq *SignalQuery) OnlyIDX(ctx context.Context) string {
	id, err := sq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Signals.
func (sq *SignalQuery) All(ctx context.Context) ([]*Signal, error) {
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryAll)
	if err := sq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Signal, *SignalQuery]()
	return withInterceptors[[]*Signal](ctx, sq, qr, sq.inters)
}

// AllX is like All, but panics if an error occurs.
func (sq *SignalQuery) AllX(ctx context.Context) []*Signal {
	nodes, err := sq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Signal IDs.
func (sq *SignalQuery) IDs(ctx context.Context) (ids []string, err error) {
	if sq.ctx.Unique == nil && sq.path != nil {
		sq.Unique(true)
	}
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryIDs)
	if err = sq.Select(signal.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (sq *SignalQuery) IDsX(ctx context.Context) []string {
	ids, err := sq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (sq *SignalQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryCount)
	if err := sq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, sq, querierCount[*SignalQuery](), sq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (sq *SignalQuery) CountX(ctx context.Context) int {
	count, err := sq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (sq *SignalQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryExist)
	switch _, err := sq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (sq *SignalQuery) ExistX(ctx context.Context) bool {
	exist, err := sq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SignalQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (sq *SignalQuery) Clone() *SignalQuery {
	if sq == nil {
		return nil
	}
	return &SignalQuery{
		config:         sq.config,
		ctx:            sq.ctx.Clone(),
		order:          append([]signal.OrderOption{}, sq.order...),
		inters:         append([]Interceptor{}, sq.inters...),
		predicates:     append([]predicate.Signal{}, sq.predicates...),
		withExecutions: sq.withExecutions.Clone(),
		// clone intermediate query.
		sql:  sq.sql.Clone(),
		path: sq.path,
	}
}

// WithExecutions tells the query-builder to eager-load the nodes that are connected to
// the "executions" edge. The optional arguments are used to configure the query builder of the edge.
func (sq *SignalQuery) WithExecutions(opts ...func(*SignalExecutionQuery)) *SignalQuery {
	query := (&SignalExecutionClient{config: sq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	sq.withExecutions = query
	return sq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		StepID string `json:"step_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Signal.Query().
//		GroupBy(signal.FieldStepID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (sq *SignalQuery) GroupBy(field string, fields ...string) *SignalGroupBy {
	sq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SignalGroupBy{build: sq}
	grbuild.flds = &sq.ctx.Fields
	grbuild.label = signal.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		StepID string `json:"step_id,omitempty"`
//	}
//
//	client.Signal.Query().
//		Select(signal.FieldStepID).
//		Scan(ctx, &v)
func (sq *SignalQuery) Select(fields ...string) *SignalSelect {
	sq.ctx.Fields = append(sq.ctx.Fields, fields...)
	sbuild := &SignalSelect{SignalQuery: sq}
	sbuild.label = signal.Label
	sbuild.flds, sbuild.scan = &sq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SignalSelect configured with the given aggregations.
func (sq *SignalQuery) Aggregate(fns ...AggregateFunc) *SignalSelect {
	return sq.Select().Aggregate(fns...)
}

func (sq *SignalQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range sq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, sq); err != nil {
				return err
			}
		}
	}
	for _, f := range sq.ctx.Fields {
		if !signal.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if sq.path != nil {
		prev, err := sq.path(ctx)
		if err != nil {
			return err
		}
		sq.sql = prev
	}
	return nil
}

func (sq *SignalQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Signal, error) {
	var (
		nodes       = []*Signal{}
		_spec       = sq.querySpec()
		loadedTypes = [1]bool{
			sq.withExecutions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Signal).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Signal{config: sq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, sq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := sq.withExecutions; query != nil {
		if err := sq.loadExecutions(ctx, query, nodes,
			func(n *Signal) { n.Edges.Executions = []*SignalExecution{} },
			func(n *Signal, e *SignalExecution) { n.Edges.Executions = append(n.Edges.Executions, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (sq *SignalQuery) loadExecutions(ctx context.Context, query *SignalExecutionQuery, nodes []*Signal, init func(*Signal), assign func(*Signal, *SignalExecution)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[string]*Signal)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.SignalExecution(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(signal.ExecutionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.signal_executions
		if fk == nil {
			return fmt.Errorf(`foreign-key "signal_executions" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "signal_executions" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (sq *SignalQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := sq.querySpec()
	_spec.Node.Columns = sq.ctx.Fields
	if len(sq.ctx.Fields) > 0 {
		_spec.Unique = sq.ctx.Unique != nil && *sq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, sq.driver, _spec)
}

func (sq *SignalQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(signal.Table, signal.Columns, sqlgraph.NewFieldSpec(signal.FieldID, field.TypeString))
	_spec.From = sq.sql
	if unique := sq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if sq.path != nil {
		_spec.Unique = true
	}
	if fields := sq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, signal.FieldID)
		for i := range fields {
			if fields[i] != signal.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := sq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := sq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := sq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := sq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (sq *SignalQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(sq.driver.Dialect())
	t1 := builder.Table(signal.Table)
	columns := sq.ctx.Fields
	if len(columns) == 0 {
		columns = signal.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if sq.sql != nil {
		selector = sq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if sq.ctx.Unique != nil && *sq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range sq.predicates {
		p(selector)
	}
	for _, p := range sq.order {
		p(selector)
	}
	if offset := sq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := sq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// SignalGroupBy is the group-by builder for Signal entities.
type SignalGroupBy struct {
	selector
	build *SignalQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (sgb *SignalGroupBy) Aggregate(fns ...AggregateFunc) *SignalGroupBy {
	sgb.fns = append(sgb.fns, fns...)
	return sgb
}

// Scan applies the selector query and scans the result into the given value.
func (sgb *SignalGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sgb.build.ctx, ent.OpQueryGroupBy)
	if err := sgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SignalQuery, *SignalGroupBy](ctx, sgb.build, sgb, sgb.build.inters, v)
}

func (sgb *SignalGroupBy) sqlScan(ctx context.Context, root *SignalQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(sgb.fns))
	for _, fn := range sgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*sgb.flds)+len(sgb.fns))
		for _, f := range *sgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*sgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SignalSelect is the builder for selecting fields of Signal entities.
type SignalSelect struct {
	*SignalQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ss *SignalSelect) Aggregate(fns ...AggregateFunc) *SignalSelect {
	ss.fns = append(ss.fns, fns...)
	return ss
}

// Scan applies the selector query and scans the result into the given value.
func (ss *SignalSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ss.ctx, ent.OpQuerySelect)
	if err := ss.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SignalQuery, *SignalSelect](ctx, ss.SignalQuery, ss, ss.inters, v)
}

func (ss *SignalSelect) sqlScan(ctx context.Context, root *SignalQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ss.fns))
	for _, fn := range ss.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ss.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ss.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}