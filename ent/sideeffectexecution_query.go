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
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/sideeffectentity"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecutiondata"
)

// SideEffectExecutionQuery is the builder for querying SideEffectExecution entities.
type SideEffectExecutionQuery struct {
	config
	ctx               *QueryContext
	order             []sideeffectexecution.OrderOption
	inters            []Interceptor
	predicates        []predicate.SideEffectExecution
	withSideEffect    *SideEffectEntityQuery
	withExecutionData *SideEffectExecutionDataQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SideEffectExecutionQuery builder.
func (seeq *SideEffectExecutionQuery) Where(ps ...predicate.SideEffectExecution) *SideEffectExecutionQuery {
	seeq.predicates = append(seeq.predicates, ps...)
	return seeq
}

// Limit the number of records to be returned by this query.
func (seeq *SideEffectExecutionQuery) Limit(limit int) *SideEffectExecutionQuery {
	seeq.ctx.Limit = &limit
	return seeq
}

// Offset to start from.
func (seeq *SideEffectExecutionQuery) Offset(offset int) *SideEffectExecutionQuery {
	seeq.ctx.Offset = &offset
	return seeq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (seeq *SideEffectExecutionQuery) Unique(unique bool) *SideEffectExecutionQuery {
	seeq.ctx.Unique = &unique
	return seeq
}

// Order specifies how the records should be ordered.
func (seeq *SideEffectExecutionQuery) Order(o ...sideeffectexecution.OrderOption) *SideEffectExecutionQuery {
	seeq.order = append(seeq.order, o...)
	return seeq
}

// QuerySideEffect chains the current query on the "side_effect" edge.
func (seeq *SideEffectExecutionQuery) QuerySideEffect() *SideEffectEntityQuery {
	query := (&SideEffectEntityClient{config: seeq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := seeq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := seeq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(sideeffectexecution.Table, sideeffectexecution.FieldID, selector),
			sqlgraph.To(sideeffectentity.Table, sideeffectentity.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, sideeffectexecution.SideEffectTable, sideeffectexecution.SideEffectColumn),
		)
		fromU = sqlgraph.SetNeighbors(seeq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryExecutionData chains the current query on the "execution_data" edge.
func (seeq *SideEffectExecutionQuery) QueryExecutionData() *SideEffectExecutionDataQuery {
	query := (&SideEffectExecutionDataClient{config: seeq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := seeq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := seeq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(sideeffectexecution.Table, sideeffectexecution.FieldID, selector),
			sqlgraph.To(sideeffectexecutiondata.Table, sideeffectexecutiondata.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, sideeffectexecution.ExecutionDataTable, sideeffectexecution.ExecutionDataColumn),
		)
		fromU = sqlgraph.SetNeighbors(seeq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first SideEffectExecution entity from the query.
// Returns a *NotFoundError when no SideEffectExecution was found.
func (seeq *SideEffectExecutionQuery) First(ctx context.Context) (*SideEffectExecution, error) {
	nodes, err := seeq.Limit(1).All(setContextOp(ctx, seeq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{sideeffectexecution.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) FirstX(ctx context.Context) *SideEffectExecution {
	node, err := seeq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first SideEffectExecution ID from the query.
// Returns a *NotFoundError when no SideEffectExecution ID was found.
func (seeq *SideEffectExecutionQuery) FirstID(ctx context.Context) (id schema.SideEffectExecutionID, err error) {
	var ids []schema.SideEffectExecutionID
	if ids, err = seeq.Limit(1).IDs(setContextOp(ctx, seeq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{sideeffectexecution.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) FirstIDX(ctx context.Context) schema.SideEffectExecutionID {
	id, err := seeq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single SideEffectExecution entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one SideEffectExecution entity is found.
// Returns a *NotFoundError when no SideEffectExecution entities are found.
func (seeq *SideEffectExecutionQuery) Only(ctx context.Context) (*SideEffectExecution, error) {
	nodes, err := seeq.Limit(2).All(setContextOp(ctx, seeq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{sideeffectexecution.Label}
	default:
		return nil, &NotSingularError{sideeffectexecution.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) OnlyX(ctx context.Context) *SideEffectExecution {
	node, err := seeq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only SideEffectExecution ID in the query.
// Returns a *NotSingularError when more than one SideEffectExecution ID is found.
// Returns a *NotFoundError when no entities are found.
func (seeq *SideEffectExecutionQuery) OnlyID(ctx context.Context) (id schema.SideEffectExecutionID, err error) {
	var ids []schema.SideEffectExecutionID
	if ids, err = seeq.Limit(2).IDs(setContextOp(ctx, seeq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{sideeffectexecution.Label}
	default:
		err = &NotSingularError{sideeffectexecution.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) OnlyIDX(ctx context.Context) schema.SideEffectExecutionID {
	id, err := seeq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of SideEffectExecutions.
func (seeq *SideEffectExecutionQuery) All(ctx context.Context) ([]*SideEffectExecution, error) {
	ctx = setContextOp(ctx, seeq.ctx, ent.OpQueryAll)
	if err := seeq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*SideEffectExecution, *SideEffectExecutionQuery]()
	return withInterceptors[[]*SideEffectExecution](ctx, seeq, qr, seeq.inters)
}

// AllX is like All, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) AllX(ctx context.Context) []*SideEffectExecution {
	nodes, err := seeq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of SideEffectExecution IDs.
func (seeq *SideEffectExecutionQuery) IDs(ctx context.Context) (ids []schema.SideEffectExecutionID, err error) {
	if seeq.ctx.Unique == nil && seeq.path != nil {
		seeq.Unique(true)
	}
	ctx = setContextOp(ctx, seeq.ctx, ent.OpQueryIDs)
	if err = seeq.Select(sideeffectexecution.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) IDsX(ctx context.Context) []schema.SideEffectExecutionID {
	ids, err := seeq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (seeq *SideEffectExecutionQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, seeq.ctx, ent.OpQueryCount)
	if err := seeq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, seeq, querierCount[*SideEffectExecutionQuery](), seeq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) CountX(ctx context.Context) int {
	count, err := seeq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (seeq *SideEffectExecutionQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, seeq.ctx, ent.OpQueryExist)
	switch _, err := seeq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (seeq *SideEffectExecutionQuery) ExistX(ctx context.Context) bool {
	exist, err := seeq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SideEffectExecutionQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (seeq *SideEffectExecutionQuery) Clone() *SideEffectExecutionQuery {
	if seeq == nil {
		return nil
	}
	return &SideEffectExecutionQuery{
		config:            seeq.config,
		ctx:               seeq.ctx.Clone(),
		order:             append([]sideeffectexecution.OrderOption{}, seeq.order...),
		inters:            append([]Interceptor{}, seeq.inters...),
		predicates:        append([]predicate.SideEffectExecution{}, seeq.predicates...),
		withSideEffect:    seeq.withSideEffect.Clone(),
		withExecutionData: seeq.withExecutionData.Clone(),
		// clone intermediate query.
		sql:  seeq.sql.Clone(),
		path: seeq.path,
	}
}

// WithSideEffect tells the query-builder to eager-load the nodes that are connected to
// the "side_effect" edge. The optional arguments are used to configure the query builder of the edge.
func (seeq *SideEffectExecutionQuery) WithSideEffect(opts ...func(*SideEffectEntityQuery)) *SideEffectExecutionQuery {
	query := (&SideEffectEntityClient{config: seeq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	seeq.withSideEffect = query
	return seeq
}

// WithExecutionData tells the query-builder to eager-load the nodes that are connected to
// the "execution_data" edge. The optional arguments are used to configure the query builder of the edge.
func (seeq *SideEffectExecutionQuery) WithExecutionData(opts ...func(*SideEffectExecutionDataQuery)) *SideEffectExecutionQuery {
	query := (&SideEffectExecutionDataClient{config: seeq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	seeq.withExecutionData = query
	return seeq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		SideEffectEntityID schema.SideEffectEntityID `json:"side_effect_entity_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.SideEffectExecution.Query().
//		GroupBy(sideeffectexecution.FieldSideEffectEntityID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (seeq *SideEffectExecutionQuery) GroupBy(field string, fields ...string) *SideEffectExecutionGroupBy {
	seeq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SideEffectExecutionGroupBy{build: seeq}
	grbuild.flds = &seeq.ctx.Fields
	grbuild.label = sideeffectexecution.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		SideEffectEntityID schema.SideEffectEntityID `json:"side_effect_entity_id,omitempty"`
//	}
//
//	client.SideEffectExecution.Query().
//		Select(sideeffectexecution.FieldSideEffectEntityID).
//		Scan(ctx, &v)
func (seeq *SideEffectExecutionQuery) Select(fields ...string) *SideEffectExecutionSelect {
	seeq.ctx.Fields = append(seeq.ctx.Fields, fields...)
	sbuild := &SideEffectExecutionSelect{SideEffectExecutionQuery: seeq}
	sbuild.label = sideeffectexecution.Label
	sbuild.flds, sbuild.scan = &seeq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SideEffectExecutionSelect configured with the given aggregations.
func (seeq *SideEffectExecutionQuery) Aggregate(fns ...AggregateFunc) *SideEffectExecutionSelect {
	return seeq.Select().Aggregate(fns...)
}

func (seeq *SideEffectExecutionQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range seeq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, seeq); err != nil {
				return err
			}
		}
	}
	for _, f := range seeq.ctx.Fields {
		if !sideeffectexecution.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if seeq.path != nil {
		prev, err := seeq.path(ctx)
		if err != nil {
			return err
		}
		seeq.sql = prev
	}
	return nil
}

func (seeq *SideEffectExecutionQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*SideEffectExecution, error) {
	var (
		nodes       = []*SideEffectExecution{}
		_spec       = seeq.querySpec()
		loadedTypes = [2]bool{
			seeq.withSideEffect != nil,
			seeq.withExecutionData != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*SideEffectExecution).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &SideEffectExecution{config: seeq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, seeq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := seeq.withSideEffect; query != nil {
		if err := seeq.loadSideEffect(ctx, query, nodes, nil,
			func(n *SideEffectExecution, e *SideEffectEntity) { n.Edges.SideEffect = e }); err != nil {
			return nil, err
		}
	}
	if query := seeq.withExecutionData; query != nil {
		if err := seeq.loadExecutionData(ctx, query, nodes, nil,
			func(n *SideEffectExecution, e *SideEffectExecutionData) { n.Edges.ExecutionData = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (seeq *SideEffectExecutionQuery) loadSideEffect(ctx context.Context, query *SideEffectEntityQuery, nodes []*SideEffectExecution, init func(*SideEffectExecution), assign func(*SideEffectExecution, *SideEffectEntity)) error {
	ids := make([]schema.SideEffectEntityID, 0, len(nodes))
	nodeids := make(map[schema.SideEffectEntityID][]*SideEffectExecution)
	for i := range nodes {
		fk := nodes[i].SideEffectEntityID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(sideeffectentity.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "side_effect_entity_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (seeq *SideEffectExecutionQuery) loadExecutionData(ctx context.Context, query *SideEffectExecutionDataQuery, nodes []*SideEffectExecution, init func(*SideEffectExecution), assign func(*SideEffectExecution, *SideEffectExecutionData)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[schema.SideEffectExecutionID]*SideEffectExecution)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(sideeffectexecutiondata.FieldExecutionID)
	}
	query.Where(predicate.SideEffectExecutionData(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(sideeffectexecution.ExecutionDataColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.ExecutionID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "execution_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (seeq *SideEffectExecutionQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := seeq.querySpec()
	_spec.Node.Columns = seeq.ctx.Fields
	if len(seeq.ctx.Fields) > 0 {
		_spec.Unique = seeq.ctx.Unique != nil && *seeq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, seeq.driver, _spec)
}

func (seeq *SideEffectExecutionQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(sideeffectexecution.Table, sideeffectexecution.Columns, sqlgraph.NewFieldSpec(sideeffectexecution.FieldID, field.TypeInt))
	_spec.From = seeq.sql
	if unique := seeq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if seeq.path != nil {
		_spec.Unique = true
	}
	if fields := seeq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sideeffectexecution.FieldID)
		for i := range fields {
			if fields[i] != sideeffectexecution.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if seeq.withSideEffect != nil {
			_spec.Node.AddColumnOnce(sideeffectexecution.FieldSideEffectEntityID)
		}
	}
	if ps := seeq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := seeq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := seeq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := seeq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (seeq *SideEffectExecutionQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(seeq.driver.Dialect())
	t1 := builder.Table(sideeffectexecution.Table)
	columns := seeq.ctx.Fields
	if len(columns) == 0 {
		columns = sideeffectexecution.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if seeq.sql != nil {
		selector = seeq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if seeq.ctx.Unique != nil && *seeq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range seeq.predicates {
		p(selector)
	}
	for _, p := range seeq.order {
		p(selector)
	}
	if offset := seeq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := seeq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// SideEffectExecutionGroupBy is the group-by builder for SideEffectExecution entities.
type SideEffectExecutionGroupBy struct {
	selector
	build *SideEffectExecutionQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (seegb *SideEffectExecutionGroupBy) Aggregate(fns ...AggregateFunc) *SideEffectExecutionGroupBy {
	seegb.fns = append(seegb.fns, fns...)
	return seegb
}

// Scan applies the selector query and scans the result into the given value.
func (seegb *SideEffectExecutionGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, seegb.build.ctx, ent.OpQueryGroupBy)
	if err := seegb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SideEffectExecutionQuery, *SideEffectExecutionGroupBy](ctx, seegb.build, seegb, seegb.build.inters, v)
}

func (seegb *SideEffectExecutionGroupBy) sqlScan(ctx context.Context, root *SideEffectExecutionQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(seegb.fns))
	for _, fn := range seegb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*seegb.flds)+len(seegb.fns))
		for _, f := range *seegb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*seegb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := seegb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SideEffectExecutionSelect is the builder for selecting fields of SideEffectExecution entities.
type SideEffectExecutionSelect struct {
	*SideEffectExecutionQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (sees *SideEffectExecutionSelect) Aggregate(fns ...AggregateFunc) *SideEffectExecutionSelect {
	sees.fns = append(sees.fns, fns...)
	return sees
}

// Scan applies the selector query and scans the result into the given value.
func (sees *SideEffectExecutionSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sees.ctx, ent.OpQuerySelect)
	if err := sees.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SideEffectExecutionQuery, *SideEffectExecutionSelect](ctx, sees.SideEffectExecutionQuery, sees, sees.inters, v)
}

func (sees *SideEffectExecutionSelect) sqlScan(ctx context.Context, root *SideEffectExecutionQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(sees.fns))
	for _, fn := range sees.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*sees.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sees.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
