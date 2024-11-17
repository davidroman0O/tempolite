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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/activityexecutiondata"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
)

// ActivityExecutionQuery is the builder for querying ActivityExecution entities.
type ActivityExecutionQuery struct {
	config
	ctx               *QueryContext
	order             []activityexecution.OrderOption
	inters            []Interceptor
	predicates        []predicate.ActivityExecution
	withExecution     *ExecutionQuery
	withExecutionData *ActivityExecutionDataQuery
	withFKs           bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ActivityExecutionQuery builder.
func (aeq *ActivityExecutionQuery) Where(ps ...predicate.ActivityExecution) *ActivityExecutionQuery {
	aeq.predicates = append(aeq.predicates, ps...)
	return aeq
}

// Limit the number of records to be returned by this query.
func (aeq *ActivityExecutionQuery) Limit(limit int) *ActivityExecutionQuery {
	aeq.ctx.Limit = &limit
	return aeq
}

// Offset to start from.
func (aeq *ActivityExecutionQuery) Offset(offset int) *ActivityExecutionQuery {
	aeq.ctx.Offset = &offset
	return aeq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (aeq *ActivityExecutionQuery) Unique(unique bool) *ActivityExecutionQuery {
	aeq.ctx.Unique = &unique
	return aeq
}

// Order specifies how the records should be ordered.
func (aeq *ActivityExecutionQuery) Order(o ...activityexecution.OrderOption) *ActivityExecutionQuery {
	aeq.order = append(aeq.order, o...)
	return aeq
}

// QueryExecution chains the current query on the "execution" edge.
func (aeq *ActivityExecutionQuery) QueryExecution() *ExecutionQuery {
	query := (&ExecutionClient{config: aeq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aeq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aeq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(activityexecution.Table, activityexecution.FieldID, selector),
			sqlgraph.To(execution.Table, execution.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, activityexecution.ExecutionTable, activityexecution.ExecutionColumn),
		)
		fromU = sqlgraph.SetNeighbors(aeq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryExecutionData chains the current query on the "execution_data" edge.
func (aeq *ActivityExecutionQuery) QueryExecutionData() *ActivityExecutionDataQuery {
	query := (&ActivityExecutionDataClient{config: aeq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aeq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aeq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(activityexecution.Table, activityexecution.FieldID, selector),
			sqlgraph.To(activityexecutiondata.Table, activityexecutiondata.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, activityexecution.ExecutionDataTable, activityexecution.ExecutionDataColumn),
		)
		fromU = sqlgraph.SetNeighbors(aeq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first ActivityExecution entity from the query.
// Returns a *NotFoundError when no ActivityExecution was found.
func (aeq *ActivityExecutionQuery) First(ctx context.Context) (*ActivityExecution, error) {
	nodes, err := aeq.Limit(1).All(setContextOp(ctx, aeq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{activityexecution.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) FirstX(ctx context.Context) *ActivityExecution {
	node, err := aeq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ActivityExecution ID from the query.
// Returns a *NotFoundError when no ActivityExecution ID was found.
func (aeq *ActivityExecutionQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = aeq.Limit(1).IDs(setContextOp(ctx, aeq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{activityexecution.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) FirstIDX(ctx context.Context) int {
	id, err := aeq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ActivityExecution entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ActivityExecution entity is found.
// Returns a *NotFoundError when no ActivityExecution entities are found.
func (aeq *ActivityExecutionQuery) Only(ctx context.Context) (*ActivityExecution, error) {
	nodes, err := aeq.Limit(2).All(setContextOp(ctx, aeq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{activityexecution.Label}
	default:
		return nil, &NotSingularError{activityexecution.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) OnlyX(ctx context.Context) *ActivityExecution {
	node, err := aeq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ActivityExecution ID in the query.
// Returns a *NotSingularError when more than one ActivityExecution ID is found.
// Returns a *NotFoundError when no entities are found.
func (aeq *ActivityExecutionQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = aeq.Limit(2).IDs(setContextOp(ctx, aeq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{activityexecution.Label}
	default:
		err = &NotSingularError{activityexecution.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) OnlyIDX(ctx context.Context) int {
	id, err := aeq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ActivityExecutions.
func (aeq *ActivityExecutionQuery) All(ctx context.Context) ([]*ActivityExecution, error) {
	ctx = setContextOp(ctx, aeq.ctx, ent.OpQueryAll)
	if err := aeq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ActivityExecution, *ActivityExecutionQuery]()
	return withInterceptors[[]*ActivityExecution](ctx, aeq, qr, aeq.inters)
}

// AllX is like All, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) AllX(ctx context.Context) []*ActivityExecution {
	nodes, err := aeq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ActivityExecution IDs.
func (aeq *ActivityExecutionQuery) IDs(ctx context.Context) (ids []int, err error) {
	if aeq.ctx.Unique == nil && aeq.path != nil {
		aeq.Unique(true)
	}
	ctx = setContextOp(ctx, aeq.ctx, ent.OpQueryIDs)
	if err = aeq.Select(activityexecution.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) IDsX(ctx context.Context) []int {
	ids, err := aeq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (aeq *ActivityExecutionQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, aeq.ctx, ent.OpQueryCount)
	if err := aeq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, aeq, querierCount[*ActivityExecutionQuery](), aeq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) CountX(ctx context.Context) int {
	count, err := aeq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (aeq *ActivityExecutionQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, aeq.ctx, ent.OpQueryExist)
	switch _, err := aeq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (aeq *ActivityExecutionQuery) ExistX(ctx context.Context) bool {
	exist, err := aeq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ActivityExecutionQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (aeq *ActivityExecutionQuery) Clone() *ActivityExecutionQuery {
	if aeq == nil {
		return nil
	}
	return &ActivityExecutionQuery{
		config:            aeq.config,
		ctx:               aeq.ctx.Clone(),
		order:             append([]activityexecution.OrderOption{}, aeq.order...),
		inters:            append([]Interceptor{}, aeq.inters...),
		predicates:        append([]predicate.ActivityExecution{}, aeq.predicates...),
		withExecution:     aeq.withExecution.Clone(),
		withExecutionData: aeq.withExecutionData.Clone(),
		// clone intermediate query.
		sql:  aeq.sql.Clone(),
		path: aeq.path,
	}
}

// WithExecution tells the query-builder to eager-load the nodes that are connected to
// the "execution" edge. The optional arguments are used to configure the query builder of the edge.
func (aeq *ActivityExecutionQuery) WithExecution(opts ...func(*ExecutionQuery)) *ActivityExecutionQuery {
	query := (&ExecutionClient{config: aeq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aeq.withExecution = query
	return aeq
}

// WithExecutionData tells the query-builder to eager-load the nodes that are connected to
// the "execution_data" edge. The optional arguments are used to configure the query builder of the edge.
func (aeq *ActivityExecutionQuery) WithExecutionData(opts ...func(*ActivityExecutionDataQuery)) *ActivityExecutionQuery {
	query := (&ActivityExecutionDataClient{config: aeq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aeq.withExecutionData = query
	return aeq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Inputs [][]uint8 `json:"inputs,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.ActivityExecution.Query().
//		GroupBy(activityexecution.FieldInputs).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (aeq *ActivityExecutionQuery) GroupBy(field string, fields ...string) *ActivityExecutionGroupBy {
	aeq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ActivityExecutionGroupBy{build: aeq}
	grbuild.flds = &aeq.ctx.Fields
	grbuild.label = activityexecution.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Inputs [][]uint8 `json:"inputs,omitempty"`
//	}
//
//	client.ActivityExecution.Query().
//		Select(activityexecution.FieldInputs).
//		Scan(ctx, &v)
func (aeq *ActivityExecutionQuery) Select(fields ...string) *ActivityExecutionSelect {
	aeq.ctx.Fields = append(aeq.ctx.Fields, fields...)
	sbuild := &ActivityExecutionSelect{ActivityExecutionQuery: aeq}
	sbuild.label = activityexecution.Label
	sbuild.flds, sbuild.scan = &aeq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ActivityExecutionSelect configured with the given aggregations.
func (aeq *ActivityExecutionQuery) Aggregate(fns ...AggregateFunc) *ActivityExecutionSelect {
	return aeq.Select().Aggregate(fns...)
}

func (aeq *ActivityExecutionQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range aeq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, aeq); err != nil {
				return err
			}
		}
	}
	for _, f := range aeq.ctx.Fields {
		if !activityexecution.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if aeq.path != nil {
		prev, err := aeq.path(ctx)
		if err != nil {
			return err
		}
		aeq.sql = prev
	}
	return nil
}

func (aeq *ActivityExecutionQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ActivityExecution, error) {
	var (
		nodes       = []*ActivityExecution{}
		withFKs     = aeq.withFKs
		_spec       = aeq.querySpec()
		loadedTypes = [2]bool{
			aeq.withExecution != nil,
			aeq.withExecutionData != nil,
		}
	)
	if aeq.withExecution != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, activityexecution.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ActivityExecution).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ActivityExecution{config: aeq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, aeq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := aeq.withExecution; query != nil {
		if err := aeq.loadExecution(ctx, query, nodes, nil,
			func(n *ActivityExecution, e *Execution) { n.Edges.Execution = e }); err != nil {
			return nil, err
		}
	}
	if query := aeq.withExecutionData; query != nil {
		if err := aeq.loadExecutionData(ctx, query, nodes, nil,
			func(n *ActivityExecution, e *ActivityExecutionData) { n.Edges.ExecutionData = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (aeq *ActivityExecutionQuery) loadExecution(ctx context.Context, query *ExecutionQuery, nodes []*ActivityExecution, init func(*ActivityExecution), assign func(*ActivityExecution, *Execution)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*ActivityExecution)
	for i := range nodes {
		if nodes[i].execution_activity_execution == nil {
			continue
		}
		fk := *nodes[i].execution_activity_execution
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(execution.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "execution_activity_execution" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (aeq *ActivityExecutionQuery) loadExecutionData(ctx context.Context, query *ActivityExecutionDataQuery, nodes []*ActivityExecution, init func(*ActivityExecution), assign func(*ActivityExecution, *ActivityExecutionData)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*ActivityExecution)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
	}
	query.withFKs = true
	query.Where(predicate.ActivityExecutionData(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(activityexecution.ExecutionDataColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.activity_execution_execution_data
		if fk == nil {
			return fmt.Errorf(`foreign-key "activity_execution_execution_data" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "activity_execution_execution_data" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (aeq *ActivityExecutionQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := aeq.querySpec()
	_spec.Node.Columns = aeq.ctx.Fields
	if len(aeq.ctx.Fields) > 0 {
		_spec.Unique = aeq.ctx.Unique != nil && *aeq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, aeq.driver, _spec)
}

func (aeq *ActivityExecutionQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(activityexecution.Table, activityexecution.Columns, sqlgraph.NewFieldSpec(activityexecution.FieldID, field.TypeInt))
	_spec.From = aeq.sql
	if unique := aeq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if aeq.path != nil {
		_spec.Unique = true
	}
	if fields := aeq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, activityexecution.FieldID)
		for i := range fields {
			if fields[i] != activityexecution.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := aeq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := aeq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := aeq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := aeq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (aeq *ActivityExecutionQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(aeq.driver.Dialect())
	t1 := builder.Table(activityexecution.Table)
	columns := aeq.ctx.Fields
	if len(columns) == 0 {
		columns = activityexecution.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if aeq.sql != nil {
		selector = aeq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if aeq.ctx.Unique != nil && *aeq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range aeq.predicates {
		p(selector)
	}
	for _, p := range aeq.order {
		p(selector)
	}
	if offset := aeq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := aeq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// ActivityExecutionGroupBy is the group-by builder for ActivityExecution entities.
type ActivityExecutionGroupBy struct {
	selector
	build *ActivityExecutionQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (aegb *ActivityExecutionGroupBy) Aggregate(fns ...AggregateFunc) *ActivityExecutionGroupBy {
	aegb.fns = append(aegb.fns, fns...)
	return aegb
}

// Scan applies the selector query and scans the result into the given value.
func (aegb *ActivityExecutionGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, aegb.build.ctx, ent.OpQueryGroupBy)
	if err := aegb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ActivityExecutionQuery, *ActivityExecutionGroupBy](ctx, aegb.build, aegb, aegb.build.inters, v)
}

func (aegb *ActivityExecutionGroupBy) sqlScan(ctx context.Context, root *ActivityExecutionQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(aegb.fns))
	for _, fn := range aegb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*aegb.flds)+len(aegb.fns))
		for _, f := range *aegb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*aegb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := aegb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ActivityExecutionSelect is the builder for selecting fields of ActivityExecution entities.
type ActivityExecutionSelect struct {
	*ActivityExecutionQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (aes *ActivityExecutionSelect) Aggregate(fns ...AggregateFunc) *ActivityExecutionSelect {
	aes.fns = append(aes.fns, fns...)
	return aes
}

// Scan applies the selector query and scans the result into the given value.
func (aes *ActivityExecutionSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, aes.ctx, ent.OpQuerySelect)
	if err := aes.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ActivityExecutionQuery, *ActivityExecutionSelect](ctx, aes.ActivityExecutionQuery, aes, aes.inters, v)
}

func (aes *ActivityExecutionSelect) sqlScan(ctx context.Context, root *ActivityExecutionQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(aes.fns))
	for _, fn := range aes.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*aes.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := aes.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
