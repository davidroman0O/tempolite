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
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/predicate"
)

// ExecutionRelationshipQuery is the builder for querying ExecutionRelationship entities.
type ExecutionRelationshipQuery struct {
	config
	ctx        *QueryContext
	order      []executionrelationship.OrderOption
	inters     []Interceptor
	predicates []predicate.ExecutionRelationship
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ExecutionRelationshipQuery builder.
func (erq *ExecutionRelationshipQuery) Where(ps ...predicate.ExecutionRelationship) *ExecutionRelationshipQuery {
	erq.predicates = append(erq.predicates, ps...)
	return erq
}

// Limit the number of records to be returned by this query.
func (erq *ExecutionRelationshipQuery) Limit(limit int) *ExecutionRelationshipQuery {
	erq.ctx.Limit = &limit
	return erq
}

// Offset to start from.
func (erq *ExecutionRelationshipQuery) Offset(offset int) *ExecutionRelationshipQuery {
	erq.ctx.Offset = &offset
	return erq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (erq *ExecutionRelationshipQuery) Unique(unique bool) *ExecutionRelationshipQuery {
	erq.ctx.Unique = &unique
	return erq
}

// Order specifies how the records should be ordered.
func (erq *ExecutionRelationshipQuery) Order(o ...executionrelationship.OrderOption) *ExecutionRelationshipQuery {
	erq.order = append(erq.order, o...)
	return erq
}

// First returns the first ExecutionRelationship entity from the query.
// Returns a *NotFoundError when no ExecutionRelationship was found.
func (erq *ExecutionRelationshipQuery) First(ctx context.Context) (*ExecutionRelationship, error) {
	nodes, err := erq.Limit(1).All(setContextOp(ctx, erq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{executionrelationship.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) FirstX(ctx context.Context) *ExecutionRelationship {
	node, err := erq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ExecutionRelationship ID from the query.
// Returns a *NotFoundError when no ExecutionRelationship ID was found.
func (erq *ExecutionRelationshipQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = erq.Limit(1).IDs(setContextOp(ctx, erq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{executionrelationship.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) FirstIDX(ctx context.Context) int {
	id, err := erq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ExecutionRelationship entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ExecutionRelationship entity is found.
// Returns a *NotFoundError when no ExecutionRelationship entities are found.
func (erq *ExecutionRelationshipQuery) Only(ctx context.Context) (*ExecutionRelationship, error) {
	nodes, err := erq.Limit(2).All(setContextOp(ctx, erq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{executionrelationship.Label}
	default:
		return nil, &NotSingularError{executionrelationship.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) OnlyX(ctx context.Context) *ExecutionRelationship {
	node, err := erq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ExecutionRelationship ID in the query.
// Returns a *NotSingularError when more than one ExecutionRelationship ID is found.
// Returns a *NotFoundError when no entities are found.
func (erq *ExecutionRelationshipQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = erq.Limit(2).IDs(setContextOp(ctx, erq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{executionrelationship.Label}
	default:
		err = &NotSingularError{executionrelationship.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) OnlyIDX(ctx context.Context) int {
	id, err := erq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ExecutionRelationships.
func (erq *ExecutionRelationshipQuery) All(ctx context.Context) ([]*ExecutionRelationship, error) {
	ctx = setContextOp(ctx, erq.ctx, ent.OpQueryAll)
	if err := erq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ExecutionRelationship, *ExecutionRelationshipQuery]()
	return withInterceptors[[]*ExecutionRelationship](ctx, erq, qr, erq.inters)
}

// AllX is like All, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) AllX(ctx context.Context) []*ExecutionRelationship {
	nodes, err := erq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ExecutionRelationship IDs.
func (erq *ExecutionRelationshipQuery) IDs(ctx context.Context) (ids []int, err error) {
	if erq.ctx.Unique == nil && erq.path != nil {
		erq.Unique(true)
	}
	ctx = setContextOp(ctx, erq.ctx, ent.OpQueryIDs)
	if err = erq.Select(executionrelationship.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) IDsX(ctx context.Context) []int {
	ids, err := erq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (erq *ExecutionRelationshipQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, erq.ctx, ent.OpQueryCount)
	if err := erq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, erq, querierCount[*ExecutionRelationshipQuery](), erq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) CountX(ctx context.Context) int {
	count, err := erq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (erq *ExecutionRelationshipQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, erq.ctx, ent.OpQueryExist)
	switch _, err := erq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (erq *ExecutionRelationshipQuery) ExistX(ctx context.Context) bool {
	exist, err := erq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ExecutionRelationshipQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (erq *ExecutionRelationshipQuery) Clone() *ExecutionRelationshipQuery {
	if erq == nil {
		return nil
	}
	return &ExecutionRelationshipQuery{
		config:     erq.config,
		ctx:        erq.ctx.Clone(),
		order:      append([]executionrelationship.OrderOption{}, erq.order...),
		inters:     append([]Interceptor{}, erq.inters...),
		predicates: append([]predicate.ExecutionRelationship{}, erq.predicates...),
		// clone intermediate query.
		sql:  erq.sql.Clone(),
		path: erq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		ParentID string `json:"parent_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.ExecutionRelationship.Query().
//		GroupBy(executionrelationship.FieldParentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (erq *ExecutionRelationshipQuery) GroupBy(field string, fields ...string) *ExecutionRelationshipGroupBy {
	erq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ExecutionRelationshipGroupBy{build: erq}
	grbuild.flds = &erq.ctx.Fields
	grbuild.label = executionrelationship.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		ParentID string `json:"parent_id,omitempty"`
//	}
//
//	client.ExecutionRelationship.Query().
//		Select(executionrelationship.FieldParentID).
//		Scan(ctx, &v)
func (erq *ExecutionRelationshipQuery) Select(fields ...string) *ExecutionRelationshipSelect {
	erq.ctx.Fields = append(erq.ctx.Fields, fields...)
	sbuild := &ExecutionRelationshipSelect{ExecutionRelationshipQuery: erq}
	sbuild.label = executionrelationship.Label
	sbuild.flds, sbuild.scan = &erq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ExecutionRelationshipSelect configured with the given aggregations.
func (erq *ExecutionRelationshipQuery) Aggregate(fns ...AggregateFunc) *ExecutionRelationshipSelect {
	return erq.Select().Aggregate(fns...)
}

func (erq *ExecutionRelationshipQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range erq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, erq); err != nil {
				return err
			}
		}
	}
	for _, f := range erq.ctx.Fields {
		if !executionrelationship.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if erq.path != nil {
		prev, err := erq.path(ctx)
		if err != nil {
			return err
		}
		erq.sql = prev
	}
	return nil
}

func (erq *ExecutionRelationshipQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ExecutionRelationship, error) {
	var (
		nodes = []*ExecutionRelationship{}
		_spec = erq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ExecutionRelationship).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ExecutionRelationship{config: erq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, erq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (erq *ExecutionRelationshipQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := erq.querySpec()
	_spec.Node.Columns = erq.ctx.Fields
	if len(erq.ctx.Fields) > 0 {
		_spec.Unique = erq.ctx.Unique != nil && *erq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, erq.driver, _spec)
}

func (erq *ExecutionRelationshipQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(executionrelationship.Table, executionrelationship.Columns, sqlgraph.NewFieldSpec(executionrelationship.FieldID, field.TypeInt))
	_spec.From = erq.sql
	if unique := erq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if erq.path != nil {
		_spec.Unique = true
	}
	if fields := erq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, executionrelationship.FieldID)
		for i := range fields {
			if fields[i] != executionrelationship.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := erq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := erq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := erq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := erq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (erq *ExecutionRelationshipQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(erq.driver.Dialect())
	t1 := builder.Table(executionrelationship.Table)
	columns := erq.ctx.Fields
	if len(columns) == 0 {
		columns = executionrelationship.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if erq.sql != nil {
		selector = erq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if erq.ctx.Unique != nil && *erq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range erq.predicates {
		p(selector)
	}
	for _, p := range erq.order {
		p(selector)
	}
	if offset := erq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := erq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// ExecutionRelationshipGroupBy is the group-by builder for ExecutionRelationship entities.
type ExecutionRelationshipGroupBy struct {
	selector
	build *ExecutionRelationshipQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ergb *ExecutionRelationshipGroupBy) Aggregate(fns ...AggregateFunc) *ExecutionRelationshipGroupBy {
	ergb.fns = append(ergb.fns, fns...)
	return ergb
}

// Scan applies the selector query and scans the result into the given value.
func (ergb *ExecutionRelationshipGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ergb.build.ctx, ent.OpQueryGroupBy)
	if err := ergb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ExecutionRelationshipQuery, *ExecutionRelationshipGroupBy](ctx, ergb.build, ergb, ergb.build.inters, v)
}

func (ergb *ExecutionRelationshipGroupBy) sqlScan(ctx context.Context, root *ExecutionRelationshipQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ergb.fns))
	for _, fn := range ergb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ergb.flds)+len(ergb.fns))
		for _, f := range *ergb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ergb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ergb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ExecutionRelationshipSelect is the builder for selecting fields of ExecutionRelationship entities.
type ExecutionRelationshipSelect struct {
	*ExecutionRelationshipQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ers *ExecutionRelationshipSelect) Aggregate(fns ...AggregateFunc) *ExecutionRelationshipSelect {
	ers.fns = append(ers.fns, fns...)
	return ers
}

// Scan applies the selector query and scans the result into the given value.
func (ers *ExecutionRelationshipSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ers.ctx, ent.OpQuerySelect)
	if err := ers.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ExecutionRelationshipQuery, *ExecutionRelationshipSelect](ctx, ers.ExecutionRelationshipQuery, ers, ers.inters, v)
}

func (ers *ExecutionRelationshipSelect) sqlScan(ctx context.Context, root *ExecutionRelationshipQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ers.fns))
	for _, fn := range ers.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ers.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ers.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}