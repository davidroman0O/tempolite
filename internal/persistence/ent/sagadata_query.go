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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/sagadata"
)

// SagaDataQuery is the builder for querying SagaData entities.
type SagaDataQuery struct {
	config
	ctx        *QueryContext
	order      []sagadata.OrderOption
	inters     []Interceptor
	predicates []predicate.SagaData
	withEntity *EntityQuery
	withFKs    bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SagaDataQuery builder.
func (sdq *SagaDataQuery) Where(ps ...predicate.SagaData) *SagaDataQuery {
	sdq.predicates = append(sdq.predicates, ps...)
	return sdq
}

// Limit the number of records to be returned by this query.
func (sdq *SagaDataQuery) Limit(limit int) *SagaDataQuery {
	sdq.ctx.Limit = &limit
	return sdq
}

// Offset to start from.
func (sdq *SagaDataQuery) Offset(offset int) *SagaDataQuery {
	sdq.ctx.Offset = &offset
	return sdq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (sdq *SagaDataQuery) Unique(unique bool) *SagaDataQuery {
	sdq.ctx.Unique = &unique
	return sdq
}

// Order specifies how the records should be ordered.
func (sdq *SagaDataQuery) Order(o ...sagadata.OrderOption) *SagaDataQuery {
	sdq.order = append(sdq.order, o...)
	return sdq
}

// QueryEntity chains the current query on the "entity" edge.
func (sdq *SagaDataQuery) QueryEntity() *EntityQuery {
	query := (&EntityClient{config: sdq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := sdq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := sdq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(sagadata.Table, sagadata.FieldID, selector),
			sqlgraph.To(entity.Table, entity.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, sagadata.EntityTable, sagadata.EntityColumn),
		)
		fromU = sqlgraph.SetNeighbors(sdq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first SagaData entity from the query.
// Returns a *NotFoundError when no SagaData was found.
func (sdq *SagaDataQuery) First(ctx context.Context) (*SagaData, error) {
	nodes, err := sdq.Limit(1).All(setContextOp(ctx, sdq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{sagadata.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (sdq *SagaDataQuery) FirstX(ctx context.Context) *SagaData {
	node, err := sdq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first SagaData ID from the query.
// Returns a *NotFoundError when no SagaData ID was found.
func (sdq *SagaDataQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = sdq.Limit(1).IDs(setContextOp(ctx, sdq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{sagadata.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (sdq *SagaDataQuery) FirstIDX(ctx context.Context) int {
	id, err := sdq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single SagaData entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one SagaData entity is found.
// Returns a *NotFoundError when no SagaData entities are found.
func (sdq *SagaDataQuery) Only(ctx context.Context) (*SagaData, error) {
	nodes, err := sdq.Limit(2).All(setContextOp(ctx, sdq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{sagadata.Label}
	default:
		return nil, &NotSingularError{sagadata.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (sdq *SagaDataQuery) OnlyX(ctx context.Context) *SagaData {
	node, err := sdq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only SagaData ID in the query.
// Returns a *NotSingularError when more than one SagaData ID is found.
// Returns a *NotFoundError when no entities are found.
func (sdq *SagaDataQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = sdq.Limit(2).IDs(setContextOp(ctx, sdq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{sagadata.Label}
	default:
		err = &NotSingularError{sagadata.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (sdq *SagaDataQuery) OnlyIDX(ctx context.Context) int {
	id, err := sdq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of SagaDataSlice.
func (sdq *SagaDataQuery) All(ctx context.Context) ([]*SagaData, error) {
	ctx = setContextOp(ctx, sdq.ctx, ent.OpQueryAll)
	if err := sdq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*SagaData, *SagaDataQuery]()
	return withInterceptors[[]*SagaData](ctx, sdq, qr, sdq.inters)
}

// AllX is like All, but panics if an error occurs.
func (sdq *SagaDataQuery) AllX(ctx context.Context) []*SagaData {
	nodes, err := sdq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of SagaData IDs.
func (sdq *SagaDataQuery) IDs(ctx context.Context) (ids []int, err error) {
	if sdq.ctx.Unique == nil && sdq.path != nil {
		sdq.Unique(true)
	}
	ctx = setContextOp(ctx, sdq.ctx, ent.OpQueryIDs)
	if err = sdq.Select(sagadata.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (sdq *SagaDataQuery) IDsX(ctx context.Context) []int {
	ids, err := sdq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (sdq *SagaDataQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, sdq.ctx, ent.OpQueryCount)
	if err := sdq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, sdq, querierCount[*SagaDataQuery](), sdq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (sdq *SagaDataQuery) CountX(ctx context.Context) int {
	count, err := sdq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (sdq *SagaDataQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, sdq.ctx, ent.OpQueryExist)
	switch _, err := sdq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (sdq *SagaDataQuery) ExistX(ctx context.Context) bool {
	exist, err := sdq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SagaDataQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (sdq *SagaDataQuery) Clone() *SagaDataQuery {
	if sdq == nil {
		return nil
	}
	return &SagaDataQuery{
		config:     sdq.config,
		ctx:        sdq.ctx.Clone(),
		order:      append([]sagadata.OrderOption{}, sdq.order...),
		inters:     append([]Interceptor{}, sdq.inters...),
		predicates: append([]predicate.SagaData{}, sdq.predicates...),
		withEntity: sdq.withEntity.Clone(),
		// clone intermediate query.
		sql:  sdq.sql.Clone(),
		path: sdq.path,
	}
}

// WithEntity tells the query-builder to eager-load the nodes that are connected to
// the "entity" edge. The optional arguments are used to configure the query builder of the edge.
func (sdq *SagaDataQuery) WithEntity(opts ...func(*EntityQuery)) *SagaDataQuery {
	query := (&EntityClient{config: sdq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	sdq.withEntity = query
	return sdq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Compensating bool `json:"compensating,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.SagaData.Query().
//		GroupBy(sagadata.FieldCompensating).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (sdq *SagaDataQuery) GroupBy(field string, fields ...string) *SagaDataGroupBy {
	sdq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SagaDataGroupBy{build: sdq}
	grbuild.flds = &sdq.ctx.Fields
	grbuild.label = sagadata.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Compensating bool `json:"compensating,omitempty"`
//	}
//
//	client.SagaData.Query().
//		Select(sagadata.FieldCompensating).
//		Scan(ctx, &v)
func (sdq *SagaDataQuery) Select(fields ...string) *SagaDataSelect {
	sdq.ctx.Fields = append(sdq.ctx.Fields, fields...)
	sbuild := &SagaDataSelect{SagaDataQuery: sdq}
	sbuild.label = sagadata.Label
	sbuild.flds, sbuild.scan = &sdq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SagaDataSelect configured with the given aggregations.
func (sdq *SagaDataQuery) Aggregate(fns ...AggregateFunc) *SagaDataSelect {
	return sdq.Select().Aggregate(fns...)
}

func (sdq *SagaDataQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range sdq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, sdq); err != nil {
				return err
			}
		}
	}
	for _, f := range sdq.ctx.Fields {
		if !sagadata.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if sdq.path != nil {
		prev, err := sdq.path(ctx)
		if err != nil {
			return err
		}
		sdq.sql = prev
	}
	return nil
}

func (sdq *SagaDataQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*SagaData, error) {
	var (
		nodes       = []*SagaData{}
		withFKs     = sdq.withFKs
		_spec       = sdq.querySpec()
		loadedTypes = [1]bool{
			sdq.withEntity != nil,
		}
	)
	if sdq.withEntity != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, sagadata.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*SagaData).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &SagaData{config: sdq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, sdq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := sdq.withEntity; query != nil {
		if err := sdq.loadEntity(ctx, query, nodes, nil,
			func(n *SagaData, e *Entity) { n.Edges.Entity = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (sdq *SagaDataQuery) loadEntity(ctx context.Context, query *EntityQuery, nodes []*SagaData, init func(*SagaData), assign func(*SagaData, *Entity)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*SagaData)
	for i := range nodes {
		if nodes[i].entity_saga_data == nil {
			continue
		}
		fk := *nodes[i].entity_saga_data
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(entity.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "entity_saga_data" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (sdq *SagaDataQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := sdq.querySpec()
	_spec.Node.Columns = sdq.ctx.Fields
	if len(sdq.ctx.Fields) > 0 {
		_spec.Unique = sdq.ctx.Unique != nil && *sdq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, sdq.driver, _spec)
}

func (sdq *SagaDataQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(sagadata.Table, sagadata.Columns, sqlgraph.NewFieldSpec(sagadata.FieldID, field.TypeInt))
	_spec.From = sdq.sql
	if unique := sdq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if sdq.path != nil {
		_spec.Unique = true
	}
	if fields := sdq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, sagadata.FieldID)
		for i := range fields {
			if fields[i] != sagadata.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := sdq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := sdq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := sdq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := sdq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (sdq *SagaDataQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(sdq.driver.Dialect())
	t1 := builder.Table(sagadata.Table)
	columns := sdq.ctx.Fields
	if len(columns) == 0 {
		columns = sagadata.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if sdq.sql != nil {
		selector = sdq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if sdq.ctx.Unique != nil && *sdq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range sdq.predicates {
		p(selector)
	}
	for _, p := range sdq.order {
		p(selector)
	}
	if offset := sdq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := sdq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// SagaDataGroupBy is the group-by builder for SagaData entities.
type SagaDataGroupBy struct {
	selector
	build *SagaDataQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (sdgb *SagaDataGroupBy) Aggregate(fns ...AggregateFunc) *SagaDataGroupBy {
	sdgb.fns = append(sdgb.fns, fns...)
	return sdgb
}

// Scan applies the selector query and scans the result into the given value.
func (sdgb *SagaDataGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sdgb.build.ctx, ent.OpQueryGroupBy)
	if err := sdgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SagaDataQuery, *SagaDataGroupBy](ctx, sdgb.build, sdgb, sdgb.build.inters, v)
}

func (sdgb *SagaDataGroupBy) sqlScan(ctx context.Context, root *SagaDataQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(sdgb.fns))
	for _, fn := range sdgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*sdgb.flds)+len(sdgb.fns))
		for _, f := range *sdgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*sdgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sdgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SagaDataSelect is the builder for selecting fields of SagaData entities.
type SagaDataSelect struct {
	*SagaDataQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (sds *SagaDataSelect) Aggregate(fns ...AggregateFunc) *SagaDataSelect {
	sds.fns = append(sds.fns, fns...)
	return sds
}

// Scan applies the selector query and scans the result into the given value.
func (sds *SagaDataSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sds.ctx, ent.OpQuerySelect)
	if err := sds.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SagaDataQuery, *SagaDataSelect](ctx, sds.SagaDataQuery, sds, sds.inters, v)
}

func (sds *SagaDataSelect) sqlScan(ctx context.Context, root *SagaDataQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(sds.fns))
	for _, fn := range sds.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*sds.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sds.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
