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
	"github.com/davidroman0O/tempolite/internal/persistence/ent/predicate"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/queue"
)

// QueueQuery is the builder for querying Queue entities.
type QueueQuery struct {
	config
	ctx          *QueryContext
	order        []queue.OrderOption
	inters       []Interceptor
	predicates   []predicate.Queue
	withEntities *EntityQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the QueueQuery builder.
func (qq *QueueQuery) Where(ps ...predicate.Queue) *QueueQuery {
	qq.predicates = append(qq.predicates, ps...)
	return qq
}

// Limit the number of records to be returned by this query.
func (qq *QueueQuery) Limit(limit int) *QueueQuery {
	qq.ctx.Limit = &limit
	return qq
}

// Offset to start from.
func (qq *QueueQuery) Offset(offset int) *QueueQuery {
	qq.ctx.Offset = &offset
	return qq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (qq *QueueQuery) Unique(unique bool) *QueueQuery {
	qq.ctx.Unique = &unique
	return qq
}

// Order specifies how the records should be ordered.
func (qq *QueueQuery) Order(o ...queue.OrderOption) *QueueQuery {
	qq.order = append(qq.order, o...)
	return qq
}

// QueryEntities chains the current query on the "entities" edge.
func (qq *QueueQuery) QueryEntities() *EntityQuery {
	query := (&EntityClient{config: qq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := qq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := qq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(queue.Table, queue.FieldID, selector),
			sqlgraph.To(entity.Table, entity.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, queue.EntitiesTable, queue.EntitiesColumn),
		)
		fromU = sqlgraph.SetNeighbors(qq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Queue entity from the query.
// Returns a *NotFoundError when no Queue was found.
func (qq *QueueQuery) First(ctx context.Context) (*Queue, error) {
	nodes, err := qq.Limit(1).All(setContextOp(ctx, qq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{queue.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (qq *QueueQuery) FirstX(ctx context.Context) *Queue {
	node, err := qq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Queue ID from the query.
// Returns a *NotFoundError when no Queue ID was found.
func (qq *QueueQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = qq.Limit(1).IDs(setContextOp(ctx, qq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{queue.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (qq *QueueQuery) FirstIDX(ctx context.Context) int {
	id, err := qq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Queue entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Queue entity is found.
// Returns a *NotFoundError when no Queue entities are found.
func (qq *QueueQuery) Only(ctx context.Context) (*Queue, error) {
	nodes, err := qq.Limit(2).All(setContextOp(ctx, qq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{queue.Label}
	default:
		return nil, &NotSingularError{queue.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (qq *QueueQuery) OnlyX(ctx context.Context) *Queue {
	node, err := qq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Queue ID in the query.
// Returns a *NotSingularError when more than one Queue ID is found.
// Returns a *NotFoundError when no entities are found.
func (qq *QueueQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = qq.Limit(2).IDs(setContextOp(ctx, qq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{queue.Label}
	default:
		err = &NotSingularError{queue.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (qq *QueueQuery) OnlyIDX(ctx context.Context) int {
	id, err := qq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Queues.
func (qq *QueueQuery) All(ctx context.Context) ([]*Queue, error) {
	ctx = setContextOp(ctx, qq.ctx, ent.OpQueryAll)
	if err := qq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Queue, *QueueQuery]()
	return withInterceptors[[]*Queue](ctx, qq, qr, qq.inters)
}

// AllX is like All, but panics if an error occurs.
func (qq *QueueQuery) AllX(ctx context.Context) []*Queue {
	nodes, err := qq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Queue IDs.
func (qq *QueueQuery) IDs(ctx context.Context) (ids []int, err error) {
	if qq.ctx.Unique == nil && qq.path != nil {
		qq.Unique(true)
	}
	ctx = setContextOp(ctx, qq.ctx, ent.OpQueryIDs)
	if err = qq.Select(queue.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (qq *QueueQuery) IDsX(ctx context.Context) []int {
	ids, err := qq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (qq *QueueQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, qq.ctx, ent.OpQueryCount)
	if err := qq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, qq, querierCount[*QueueQuery](), qq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (qq *QueueQuery) CountX(ctx context.Context) int {
	count, err := qq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (qq *QueueQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, qq.ctx, ent.OpQueryExist)
	switch _, err := qq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (qq *QueueQuery) ExistX(ctx context.Context) bool {
	exist, err := qq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the QueueQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (qq *QueueQuery) Clone() *QueueQuery {
	if qq == nil {
		return nil
	}
	return &QueueQuery{
		config:       qq.config,
		ctx:          qq.ctx.Clone(),
		order:        append([]queue.OrderOption{}, qq.order...),
		inters:       append([]Interceptor{}, qq.inters...),
		predicates:   append([]predicate.Queue{}, qq.predicates...),
		withEntities: qq.withEntities.Clone(),
		// clone intermediate query.
		sql:  qq.sql.Clone(),
		path: qq.path,
	}
}

// WithEntities tells the query-builder to eager-load the nodes that are connected to
// the "entities" edge. The optional arguments are used to configure the query builder of the edge.
func (qq *QueueQuery) WithEntities(opts ...func(*EntityQuery)) *QueueQuery {
	query := (&EntityClient{config: qq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	qq.withEntities = query
	return qq
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
//	client.Queue.Query().
//		GroupBy(queue.FieldCreatedAt).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (qq *QueueQuery) GroupBy(field string, fields ...string) *QueueGroupBy {
	qq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &QueueGroupBy{build: qq}
	grbuild.flds = &qq.ctx.Fields
	grbuild.label = queue.Label
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
//	client.Queue.Query().
//		Select(queue.FieldCreatedAt).
//		Scan(ctx, &v)
func (qq *QueueQuery) Select(fields ...string) *QueueSelect {
	qq.ctx.Fields = append(qq.ctx.Fields, fields...)
	sbuild := &QueueSelect{QueueQuery: qq}
	sbuild.label = queue.Label
	sbuild.flds, sbuild.scan = &qq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a QueueSelect configured with the given aggregations.
func (qq *QueueQuery) Aggregate(fns ...AggregateFunc) *QueueSelect {
	return qq.Select().Aggregate(fns...)
}

func (qq *QueueQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range qq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, qq); err != nil {
				return err
			}
		}
	}
	for _, f := range qq.ctx.Fields {
		if !queue.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if qq.path != nil {
		prev, err := qq.path(ctx)
		if err != nil {
			return err
		}
		qq.sql = prev
	}
	return nil
}

func (qq *QueueQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Queue, error) {
	var (
		nodes       = []*Queue{}
		_spec       = qq.querySpec()
		loadedTypes = [1]bool{
			qq.withEntities != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Queue).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Queue{config: qq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, qq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := qq.withEntities; query != nil {
		if err := qq.loadEntities(ctx, query, nodes,
			func(n *Queue) { n.Edges.Entities = []*Entity{} },
			func(n *Queue, e *Entity) { n.Edges.Entities = append(n.Edges.Entities, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (qq *QueueQuery) loadEntities(ctx context.Context, query *EntityQuery, nodes []*Queue, init func(*Queue), assign func(*Queue, *Entity)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Queue)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Entity(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(queue.EntitiesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.queue_entities
		if fk == nil {
			return fmt.Errorf(`foreign-key "queue_entities" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "queue_entities" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (qq *QueueQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := qq.querySpec()
	_spec.Node.Columns = qq.ctx.Fields
	if len(qq.ctx.Fields) > 0 {
		_spec.Unique = qq.ctx.Unique != nil && *qq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, qq.driver, _spec)
}

func (qq *QueueQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(queue.Table, queue.Columns, sqlgraph.NewFieldSpec(queue.FieldID, field.TypeInt))
	_spec.From = qq.sql
	if unique := qq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if qq.path != nil {
		_spec.Unique = true
	}
	if fields := qq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, queue.FieldID)
		for i := range fields {
			if fields[i] != queue.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := qq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := qq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := qq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := qq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (qq *QueueQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(qq.driver.Dialect())
	t1 := builder.Table(queue.Table)
	columns := qq.ctx.Fields
	if len(columns) == 0 {
		columns = queue.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if qq.sql != nil {
		selector = qq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if qq.ctx.Unique != nil && *qq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range qq.predicates {
		p(selector)
	}
	for _, p := range qq.order {
		p(selector)
	}
	if offset := qq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := qq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// QueueGroupBy is the group-by builder for Queue entities.
type QueueGroupBy struct {
	selector
	build *QueueQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (qgb *QueueGroupBy) Aggregate(fns ...AggregateFunc) *QueueGroupBy {
	qgb.fns = append(qgb.fns, fns...)
	return qgb
}

// Scan applies the selector query and scans the result into the given value.
func (qgb *QueueGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, qgb.build.ctx, ent.OpQueryGroupBy)
	if err := qgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*QueueQuery, *QueueGroupBy](ctx, qgb.build, qgb, qgb.build.inters, v)
}

func (qgb *QueueGroupBy) sqlScan(ctx context.Context, root *QueueQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(qgb.fns))
	for _, fn := range qgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*qgb.flds)+len(qgb.fns))
		for _, f := range *qgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*qgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := qgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// QueueSelect is the builder for selecting fields of Queue entities.
type QueueSelect struct {
	*QueueQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (qs *QueueSelect) Aggregate(fns ...AggregateFunc) *QueueSelect {
	qs.fns = append(qs.fns, fns...)
	return qs
}

// Scan applies the selector query and scans the result into the given value.
func (qs *QueueSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, qs.ctx, ent.OpQuerySelect)
	if err := qs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*QueueQuery, *QueueSelect](ctx, qs.QueueQuery, qs, qs.inters, v)
}

func (qs *QueueSelect) sqlScan(ctx context.Context, root *QueueQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(qs.fns))
	for _, fn := range qs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*qs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := qs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
