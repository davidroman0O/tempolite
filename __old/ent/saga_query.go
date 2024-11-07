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
	"github.com/davidroman0O/tempolite/ent/saga"
	"github.com/davidroman0O/tempolite/ent/sagaexecution"
)

// SagaQuery is the builder for querying Saga entities.
type SagaQuery struct {
	config
	ctx        *QueryContext
	order      []saga.OrderOption
	inters     []Interceptor
	predicates []predicate.Saga
	withSteps  *SagaExecutionQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SagaQuery builder.
func (sq *SagaQuery) Where(ps ...predicate.Saga) *SagaQuery {
	sq.predicates = append(sq.predicates, ps...)
	return sq
}

// Limit the number of records to be returned by this query.
func (sq *SagaQuery) Limit(limit int) *SagaQuery {
	sq.ctx.Limit = &limit
	return sq
}

// Offset to start from.
func (sq *SagaQuery) Offset(offset int) *SagaQuery {
	sq.ctx.Offset = &offset
	return sq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (sq *SagaQuery) Unique(unique bool) *SagaQuery {
	sq.ctx.Unique = &unique
	return sq
}

// Order specifies how the records should be ordered.
func (sq *SagaQuery) Order(o ...saga.OrderOption) *SagaQuery {
	sq.order = append(sq.order, o...)
	return sq
}

// QuerySteps chains the current query on the "steps" edge.
func (sq *SagaQuery) QuerySteps() *SagaExecutionQuery {
	query := (&SagaExecutionClient{config: sq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := sq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := sq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(saga.Table, saga.FieldID, selector),
			sqlgraph.To(sagaexecution.Table, sagaexecution.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, saga.StepsTable, saga.StepsColumn),
		)
		fromU = sqlgraph.SetNeighbors(sq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Saga entity from the query.
// Returns a *NotFoundError when no Saga was found.
func (sq *SagaQuery) First(ctx context.Context) (*Saga, error) {
	nodes, err := sq.Limit(1).All(setContextOp(ctx, sq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{saga.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (sq *SagaQuery) FirstX(ctx context.Context) *Saga {
	node, err := sq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Saga ID from the query.
// Returns a *NotFoundError when no Saga ID was found.
func (sq *SagaQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sq.Limit(1).IDs(setContextOp(ctx, sq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{saga.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (sq *SagaQuery) FirstIDX(ctx context.Context) string {
	id, err := sq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Saga entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Saga entity is found.
// Returns a *NotFoundError when no Saga entities are found.
func (sq *SagaQuery) Only(ctx context.Context) (*Saga, error) {
	nodes, err := sq.Limit(2).All(setContextOp(ctx, sq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{saga.Label}
	default:
		return nil, &NotSingularError{saga.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (sq *SagaQuery) OnlyX(ctx context.Context) *Saga {
	node, err := sq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Saga ID in the query.
// Returns a *NotSingularError when more than one Saga ID is found.
// Returns a *NotFoundError when no entities are found.
func (sq *SagaQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sq.Limit(2).IDs(setContextOp(ctx, sq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{saga.Label}
	default:
		err = &NotSingularError{saga.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (sq *SagaQuery) OnlyIDX(ctx context.Context) string {
	id, err := sq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Sagas.
func (sq *SagaQuery) All(ctx context.Context) ([]*Saga, error) {
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryAll)
	if err := sq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Saga, *SagaQuery]()
	return withInterceptors[[]*Saga](ctx, sq, qr, sq.inters)
}

// AllX is like All, but panics if an error occurs.
func (sq *SagaQuery) AllX(ctx context.Context) []*Saga {
	nodes, err := sq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Saga IDs.
func (sq *SagaQuery) IDs(ctx context.Context) (ids []string, err error) {
	if sq.ctx.Unique == nil && sq.path != nil {
		sq.Unique(true)
	}
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryIDs)
	if err = sq.Select(saga.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (sq *SagaQuery) IDsX(ctx context.Context) []string {
	ids, err := sq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (sq *SagaQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, sq.ctx, ent.OpQueryCount)
	if err := sq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, sq, querierCount[*SagaQuery](), sq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (sq *SagaQuery) CountX(ctx context.Context) int {
	count, err := sq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (sq *SagaQuery) Exist(ctx context.Context) (bool, error) {
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
func (sq *SagaQuery) ExistX(ctx context.Context) bool {
	exist, err := sq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SagaQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (sq *SagaQuery) Clone() *SagaQuery {
	if sq == nil {
		return nil
	}
	return &SagaQuery{
		config:     sq.config,
		ctx:        sq.ctx.Clone(),
		order:      append([]saga.OrderOption{}, sq.order...),
		inters:     append([]Interceptor{}, sq.inters...),
		predicates: append([]predicate.Saga{}, sq.predicates...),
		withSteps:  sq.withSteps.Clone(),
		// clone intermediate query.
		sql:  sq.sql.Clone(),
		path: sq.path,
	}
}

// WithSteps tells the query-builder to eager-load the nodes that are connected to
// the "steps" edge. The optional arguments are used to configure the query builder of the edge.
func (sq *SagaQuery) WithSteps(opts ...func(*SagaExecutionQuery)) *SagaQuery {
	query := (&SagaExecutionClient{config: sq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	sq.withSteps = query
	return sq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		RunID string `json:"run_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Saga.Query().
//		GroupBy(saga.FieldRunID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (sq *SagaQuery) GroupBy(field string, fields ...string) *SagaGroupBy {
	sq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SagaGroupBy{build: sq}
	grbuild.flds = &sq.ctx.Fields
	grbuild.label = saga.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		RunID string `json:"run_id,omitempty"`
//	}
//
//	client.Saga.Query().
//		Select(saga.FieldRunID).
//		Scan(ctx, &v)
func (sq *SagaQuery) Select(fields ...string) *SagaSelect {
	sq.ctx.Fields = append(sq.ctx.Fields, fields...)
	sbuild := &SagaSelect{SagaQuery: sq}
	sbuild.label = saga.Label
	sbuild.flds, sbuild.scan = &sq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SagaSelect configured with the given aggregations.
func (sq *SagaQuery) Aggregate(fns ...AggregateFunc) *SagaSelect {
	return sq.Select().Aggregate(fns...)
}

func (sq *SagaQuery) prepareQuery(ctx context.Context) error {
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
		if !saga.ValidColumn(f) {
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

func (sq *SagaQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Saga, error) {
	var (
		nodes       = []*Saga{}
		_spec       = sq.querySpec()
		loadedTypes = [1]bool{
			sq.withSteps != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Saga).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Saga{config: sq.config}
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
	if query := sq.withSteps; query != nil {
		if err := sq.loadSteps(ctx, query, nodes,
			func(n *Saga) { n.Edges.Steps = []*SagaExecution{} },
			func(n *Saga, e *SagaExecution) { n.Edges.Steps = append(n.Edges.Steps, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (sq *SagaQuery) loadSteps(ctx context.Context, query *SagaExecutionQuery, nodes []*Saga, init func(*Saga), assign func(*Saga, *SagaExecution)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[string]*Saga)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.SagaExecution(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(saga.StepsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.saga_steps
		if fk == nil {
			return fmt.Errorf(`foreign-key "saga_steps" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "saga_steps" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (sq *SagaQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := sq.querySpec()
	_spec.Node.Columns = sq.ctx.Fields
	if len(sq.ctx.Fields) > 0 {
		_spec.Unique = sq.ctx.Unique != nil && *sq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, sq.driver, _spec)
}

func (sq *SagaQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(saga.Table, saga.Columns, sqlgraph.NewFieldSpec(saga.FieldID, field.TypeString))
	_spec.From = sq.sql
	if unique := sq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if sq.path != nil {
		_spec.Unique = true
	}
	if fields := sq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, saga.FieldID)
		for i := range fields {
			if fields[i] != saga.FieldID {
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

func (sq *SagaQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(sq.driver.Dialect())
	t1 := builder.Table(saga.Table)
	columns := sq.ctx.Fields
	if len(columns) == 0 {
		columns = saga.Columns
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

// SagaGroupBy is the group-by builder for Saga entities.
type SagaGroupBy struct {
	selector
	build *SagaQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (sgb *SagaGroupBy) Aggregate(fns ...AggregateFunc) *SagaGroupBy {
	sgb.fns = append(sgb.fns, fns...)
	return sgb
}

// Scan applies the selector query and scans the result into the given value.
func (sgb *SagaGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sgb.build.ctx, ent.OpQueryGroupBy)
	if err := sgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SagaQuery, *SagaGroupBy](ctx, sgb.build, sgb, sgb.build.inters, v)
}

func (sgb *SagaGroupBy) sqlScan(ctx context.Context, root *SagaQuery, v any) error {
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

// SagaSelect is the builder for selecting fields of Saga entities.
type SagaSelect struct {
	*SagaQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ss *SagaSelect) Aggregate(fns ...AggregateFunc) *SagaSelect {
	ss.fns = append(ss.fns, fns...)
	return ss
}

// Scan applies the selector query and scans the result into the given value.
func (ss *SagaSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ss.ctx, ent.OpQuerySelect)
	if err := ss.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SagaQuery, *SagaSelect](ctx, ss.SagaQuery, ss, ss.inters, v)
}

func (ss *SagaSelect) sqlScan(ctx context.Context, root *SagaQuery, v any) error {
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