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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/hierarchy"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/run"
)

// RunQuery is the builder for querying Run entities.
type RunQuery struct {
	config
	ctx             *QueryContext
	order           []run.OrderOption
	inters          []Interceptor
	predicates      []predicate.Run
	withEntities    *EntityQuery
	withHierarchies *HierarchyQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the RunQuery builder.
func (rq *RunQuery) Where(ps ...predicate.Run) *RunQuery {
	rq.predicates = append(rq.predicates, ps...)
	return rq
}

// Limit the number of records to be returned by this query.
func (rq *RunQuery) Limit(limit int) *RunQuery {
	rq.ctx.Limit = &limit
	return rq
}

// Offset to start from.
func (rq *RunQuery) Offset(offset int) *RunQuery {
	rq.ctx.Offset = &offset
	return rq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (rq *RunQuery) Unique(unique bool) *RunQuery {
	rq.ctx.Unique = &unique
	return rq
}

// Order specifies how the records should be ordered.
func (rq *RunQuery) Order(o ...run.OrderOption) *RunQuery {
	rq.order = append(rq.order, o...)
	return rq
}

// QueryEntities chains the current query on the "entities" edge.
func (rq *RunQuery) QueryEntities() *EntityQuery {
	query := (&EntityClient{config: rq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(run.Table, run.FieldID, selector),
			sqlgraph.To(entity.Table, entity.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, run.EntitiesTable, run.EntitiesColumn),
		)
		fromU = sqlgraph.SetNeighbors(rq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryHierarchies chains the current query on the "hierarchies" edge.
func (rq *RunQuery) QueryHierarchies() *HierarchyQuery {
	query := (&HierarchyClient{config: rq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(run.Table, run.FieldID, selector),
			sqlgraph.To(hierarchy.Table, hierarchy.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, run.HierarchiesTable, run.HierarchiesColumn),
		)
		fromU = sqlgraph.SetNeighbors(rq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Run entity from the query.
// Returns a *NotFoundError when no Run was found.
func (rq *RunQuery) First(ctx context.Context) (*Run, error) {
	nodes, err := rq.Limit(1).All(setContextOp(ctx, rq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{run.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (rq *RunQuery) FirstX(ctx context.Context) *Run {
	node, err := rq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Run ID from the query.
// Returns a *NotFoundError when no Run ID was found.
func (rq *RunQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = rq.Limit(1).IDs(setContextOp(ctx, rq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{run.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (rq *RunQuery) FirstIDX(ctx context.Context) int {
	id, err := rq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Run entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Run entity is found.
// Returns a *NotFoundError when no Run entities are found.
func (rq *RunQuery) Only(ctx context.Context) (*Run, error) {
	nodes, err := rq.Limit(2).All(setContextOp(ctx, rq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{run.Label}
	default:
		return nil, &NotSingularError{run.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (rq *RunQuery) OnlyX(ctx context.Context) *Run {
	node, err := rq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Run ID in the query.
// Returns a *NotSingularError when more than one Run ID is found.
// Returns a *NotFoundError when no entities are found.
func (rq *RunQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = rq.Limit(2).IDs(setContextOp(ctx, rq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{run.Label}
	default:
		err = &NotSingularError{run.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (rq *RunQuery) OnlyIDX(ctx context.Context) int {
	id, err := rq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Runs.
func (rq *RunQuery) All(ctx context.Context) ([]*Run, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryAll)
	if err := rq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Run, *RunQuery]()
	return withInterceptors[[]*Run](ctx, rq, qr, rq.inters)
}

// AllX is like All, but panics if an error occurs.
func (rq *RunQuery) AllX(ctx context.Context) []*Run {
	nodes, err := rq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Run IDs.
func (rq *RunQuery) IDs(ctx context.Context) (ids []int, err error) {
	if rq.ctx.Unique == nil && rq.path != nil {
		rq.Unique(true)
	}
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryIDs)
	if err = rq.Select(run.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (rq *RunQuery) IDsX(ctx context.Context) []int {
	ids, err := rq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (rq *RunQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryCount)
	if err := rq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, rq, querierCount[*RunQuery](), rq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (rq *RunQuery) CountX(ctx context.Context) int {
	count, err := rq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (rq *RunQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryExist)
	switch _, err := rq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (rq *RunQuery) ExistX(ctx context.Context) bool {
	exist, err := rq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the RunQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (rq *RunQuery) Clone() *RunQuery {
	if rq == nil {
		return nil
	}
	return &RunQuery{
		config:          rq.config,
		ctx:             rq.ctx.Clone(),
		order:           append([]run.OrderOption{}, rq.order...),
		inters:          append([]Interceptor{}, rq.inters...),
		predicates:      append([]predicate.Run{}, rq.predicates...),
		withEntities:    rq.withEntities.Clone(),
		withHierarchies: rq.withHierarchies.Clone(),
		// clone intermediate query.
		sql:  rq.sql.Clone(),
		path: rq.path,
	}
}

// WithEntities tells the query-builder to eager-load the nodes that are connected to
// the "entities" edge. The optional arguments are used to configure the query builder of the edge.
func (rq *RunQuery) WithEntities(opts ...func(*EntityQuery)) *RunQuery {
	query := (&EntityClient{config: rq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rq.withEntities = query
	return rq
}

// WithHierarchies tells the query-builder to eager-load the nodes that are connected to
// the "hierarchies" edge. The optional arguments are used to configure the query builder of the edge.
func (rq *RunQuery) WithHierarchies(opts ...func(*HierarchyQuery)) *RunQuery {
	query := (&HierarchyClient{config: rq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rq.withHierarchies = query
	return rq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Run.Query().
//		GroupBy(run.FieldCreatedAt).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (rq *RunQuery) GroupBy(field string, fields ...string) *RunGroupBy {
	rq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &RunGroupBy{build: rq}
	grbuild.flds = &rq.ctx.Fields
	grbuild.label = run.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//	}
//
//	client.Run.Query().
//		Select(run.FieldCreatedAt).
//		Scan(ctx, &v)
func (rq *RunQuery) Select(fields ...string) *RunSelect {
	rq.ctx.Fields = append(rq.ctx.Fields, fields...)
	sbuild := &RunSelect{RunQuery: rq}
	sbuild.label = run.Label
	sbuild.flds, sbuild.scan = &rq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a RunSelect configured with the given aggregations.
func (rq *RunQuery) Aggregate(fns ...AggregateFunc) *RunSelect {
	return rq.Select().Aggregate(fns...)
}

func (rq *RunQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range rq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, rq); err != nil {
				return err
			}
		}
	}
	for _, f := range rq.ctx.Fields {
		if !run.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if rq.path != nil {
		prev, err := rq.path(ctx)
		if err != nil {
			return err
		}
		rq.sql = prev
	}
	return nil
}

func (rq *RunQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Run, error) {
	var (
		nodes       = []*Run{}
		_spec       = rq.querySpec()
		loadedTypes = [2]bool{
			rq.withEntities != nil,
			rq.withHierarchies != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Run).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Run{config: rq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, rq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := rq.withEntities; query != nil {
		if err := rq.loadEntities(ctx, query, nodes,
			func(n *Run) { n.Edges.Entities = []*Entity{} },
			func(n *Run, e *Entity) { n.Edges.Entities = append(n.Edges.Entities, e) }); err != nil {
			return nil, err
		}
	}
	if query := rq.withHierarchies; query != nil {
		if err := rq.loadHierarchies(ctx, query, nodes,
			func(n *Run) { n.Edges.Hierarchies = []*Hierarchy{} },
			func(n *Run, e *Hierarchy) { n.Edges.Hierarchies = append(n.Edges.Hierarchies, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (rq *RunQuery) loadEntities(ctx context.Context, query *EntityQuery, nodes []*Run, init func(*Run), assign func(*Run, *Entity)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Run)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Entity(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(run.EntitiesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.run_entities
		if fk == nil {
			return fmt.Errorf(`foreign-key "run_entities" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "run_entities" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (rq *RunQuery) loadHierarchies(ctx context.Context, query *HierarchyQuery, nodes []*Run, init func(*Run), assign func(*Run, *Hierarchy)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Run)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(hierarchy.FieldRunID)
	}
	query.Where(predicate.Hierarchy(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(run.HierarchiesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.RunID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "run_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (rq *RunQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := rq.querySpec()
	_spec.Node.Columns = rq.ctx.Fields
	if len(rq.ctx.Fields) > 0 {
		_spec.Unique = rq.ctx.Unique != nil && *rq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, rq.driver, _spec)
}

func (rq *RunQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(run.Table, run.Columns, sqlgraph.NewFieldSpec(run.FieldID, field.TypeInt))
	_spec.From = rq.sql
	if unique := rq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if rq.path != nil {
		_spec.Unique = true
	}
	if fields := rq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, run.FieldID)
		for i := range fields {
			if fields[i] != run.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := rq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := rq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := rq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := rq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (rq *RunQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(rq.driver.Dialect())
	t1 := builder.Table(run.Table)
	columns := rq.ctx.Fields
	if len(columns) == 0 {
		columns = run.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if rq.sql != nil {
		selector = rq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if rq.ctx.Unique != nil && *rq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range rq.predicates {
		p(selector)
	}
	for _, p := range rq.order {
		p(selector)
	}
	if offset := rq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := rq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// RunGroupBy is the group-by builder for Run entities.
type RunGroupBy struct {
	selector
	build *RunQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (rgb *RunGroupBy) Aggregate(fns ...AggregateFunc) *RunGroupBy {
	rgb.fns = append(rgb.fns, fns...)
	return rgb
}

// Scan applies the selector query and scans the result into the given value.
func (rgb *RunGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rgb.build.ctx, ent.OpQueryGroupBy)
	if err := rgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RunQuery, *RunGroupBy](ctx, rgb.build, rgb, rgb.build.inters, v)
}

func (rgb *RunGroupBy) sqlScan(ctx context.Context, root *RunQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(rgb.fns))
	for _, fn := range rgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*rgb.flds)+len(rgb.fns))
		for _, f := range *rgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*rgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// RunSelect is the builder for selecting fields of Run entities.
type RunSelect struct {
	*RunQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (rs *RunSelect) Aggregate(fns ...AggregateFunc) *RunSelect {
	rs.fns = append(rs.fns, fns...)
	return rs
}

// Scan applies the selector query and scans the result into the given value.
func (rs *RunSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rs.ctx, ent.OpQuerySelect)
	if err := rs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RunQuery, *RunSelect](ctx, rs.RunQuery, rs, rs.inters, v)
}

func (rs *RunSelect) sqlScan(ctx context.Context, root *RunQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(rs.fns))
	for _, fn := range rs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*rs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}