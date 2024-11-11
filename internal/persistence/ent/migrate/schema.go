// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// ActivityDataColumns holds the columns for the "activity_data" table.
	ActivityDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "timeout", Type: field.TypeInt64, Nullable: true},
		{Name: "max_attempts", Type: field.TypeInt, Default: 1},
		{Name: "scheduled_for", Type: field.TypeTime, Nullable: true},
		{Name: "retry_policy", Type: field.TypeJSON},
		{Name: "input", Type: field.TypeJSON, Nullable: true},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "entity_activity_data", Type: field.TypeInt, Unique: true},
	}
	// ActivityDataTable holds the schema information for the "activity_data" table.
	ActivityDataTable = &schema.Table{
		Name:       "activity_data",
		Columns:    ActivityDataColumns,
		PrimaryKey: []*schema.Column{ActivityDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "activity_data_entities_activity_data",
				Columns:    []*schema.Column{ActivityDataColumns[7]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// ActivityExecutionsColumns holds the columns for the "activity_executions" table.
	ActivityExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "attempt", Type: field.TypeInt, Default: 1},
		{Name: "input", Type: field.TypeJSON, Nullable: true},
		{Name: "execution_activity_execution", Type: field.TypeInt, Unique: true},
	}
	// ActivityExecutionsTable holds the schema information for the "activity_executions" table.
	ActivityExecutionsTable = &schema.Table{
		Name:       "activity_executions",
		Columns:    ActivityExecutionsColumns,
		PrimaryKey: []*schema.Column{ActivityExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "activity_executions_executions_activity_execution",
				Columns:    []*schema.Column{ActivityExecutionsColumns[3]},
				RefColumns: []*schema.Column{ExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// ActivityExecutionDataColumns holds the columns for the "activity_execution_data" table.
	ActivityExecutionDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "heartbeats", Type: field.TypeJSON, Nullable: true},
		{Name: "last_heartbeat", Type: field.TypeTime, Nullable: true},
		{Name: "progress", Type: field.TypeJSON, Nullable: true},
		{Name: "execution_details", Type: field.TypeJSON, Nullable: true},
		{Name: "activity_execution_execution_data", Type: field.TypeInt, Unique: true},
	}
	// ActivityExecutionDataTable holds the schema information for the "activity_execution_data" table.
	ActivityExecutionDataTable = &schema.Table{
		Name:       "activity_execution_data",
		Columns:    ActivityExecutionDataColumns,
		PrimaryKey: []*schema.Column{ActivityExecutionDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "activity_execution_data_activity_executions_execution_data",
				Columns:    []*schema.Column{ActivityExecutionDataColumns[5]},
				RefColumns: []*schema.Column{ActivityExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// EntitiesColumns holds the columns for the "entities" table.
	EntitiesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "handler_name", Type: field.TypeString},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"Workflow", "Activity", "Saga", "SideEffect"}},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Retried", "Cancelled", "Paused"}, Default: "Pending"},
		{Name: "step_id", Type: field.TypeString, Unique: true},
		{Name: "queue_entities", Type: field.TypeInt, Nullable: true},
		{Name: "run_entities", Type: field.TypeInt},
	}
	// EntitiesTable holds the schema information for the "entities" table.
	EntitiesTable = &schema.Table{
		Name:       "entities",
		Columns:    EntitiesColumns,
		PrimaryKey: []*schema.Column{EntitiesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "entities_queues_entities",
				Columns:    []*schema.Column{EntitiesColumns[7]},
				RefColumns: []*schema.Column{QueuesColumns[0]},
				OnDelete:   schema.SetNull,
			},
			{
				Symbol:     "entities_runs_entities",
				Columns:    []*schema.Column{EntitiesColumns[8]},
				RefColumns: []*schema.Column{RunsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// ExecutionsColumns holds the columns for the "executions" table.
	ExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "completed_at", Type: field.TypeTime, Nullable: true},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Retried", "Cancelled", "Paused"}, Default: "Pending"},
		{Name: "entity_executions", Type: field.TypeInt},
	}
	// ExecutionsTable holds the schema information for the "executions" table.
	ExecutionsTable = &schema.Table{
		Name:       "executions",
		Columns:    ExecutionsColumns,
		PrimaryKey: []*schema.Column{ExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "executions_entities_executions",
				Columns:    []*schema.Column{ExecutionsColumns[6]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// HierarchiesColumns holds the columns for the "hierarchies" table.
	HierarchiesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "parent_execution_id", Type: field.TypeInt},
		{Name: "child_execution_id", Type: field.TypeInt},
		{Name: "parent_step_id", Type: field.TypeString},
		{Name: "child_step_id", Type: field.TypeString},
		{Name: "child_type", Type: field.TypeEnum, Enums: []string{"Workflow", "Activity", "Saga", "SideEffect"}},
		{Name: "parent_type", Type: field.TypeEnum, Enums: []string{"Workflow", "Activity", "Saga", "SideEffect"}},
		{Name: "parent_entity_id", Type: field.TypeInt},
		{Name: "child_entity_id", Type: field.TypeInt},
		{Name: "run_id", Type: field.TypeInt},
	}
	// HierarchiesTable holds the schema information for the "hierarchies" table.
	HierarchiesTable = &schema.Table{
		Name:       "hierarchies",
		Columns:    HierarchiesColumns,
		PrimaryKey: []*schema.Column{HierarchiesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "hierarchies_entities_parent_entity",
				Columns:    []*schema.Column{HierarchiesColumns[7]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
			{
				Symbol:     "hierarchies_entities_child_entity",
				Columns:    []*schema.Column{HierarchiesColumns[8]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
			{
				Symbol:     "hierarchies_runs_hierarchies",
				Columns:    []*schema.Column{HierarchiesColumns[9]},
				RefColumns: []*schema.Column{RunsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// QueuesColumns holds the columns for the "queues" table.
	QueuesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Unique: true},
	}
	// QueuesTable holds the schema information for the "queues" table.
	QueuesTable = &schema.Table{
		Name:       "queues",
		Columns:    QueuesColumns,
		PrimaryKey: []*schema.Column{QueuesColumns[0]},
	}
	// RunsColumns holds the columns for the "runs" table.
	RunsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Cancelled"}, Default: "Pending"},
	}
	// RunsTable holds the schema information for the "runs" table.
	RunsTable = &schema.Table{
		Name:       "runs",
		Columns:    RunsColumns,
		PrimaryKey: []*schema.Column{RunsColumns[0]},
	}
	// SagaDataColumns holds the columns for the "saga_data" table.
	SagaDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "compensating", Type: field.TypeBool, Default: false},
		{Name: "compensation_data", Type: field.TypeJSON, Nullable: true},
		{Name: "entity_saga_data", Type: field.TypeInt, Unique: true},
	}
	// SagaDataTable holds the schema information for the "saga_data" table.
	SagaDataTable = &schema.Table{
		Name:       "saga_data",
		Columns:    SagaDataColumns,
		PrimaryKey: []*schema.Column{SagaDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "saga_data_entities_saga_data",
				Columns:    []*schema.Column{SagaDataColumns[3]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SagaExecutionsColumns holds the columns for the "saga_executions" table.
	SagaExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "step_type", Type: field.TypeEnum, Enums: []string{"transaction", "compensation"}},
		{Name: "compensation_data", Type: field.TypeJSON, Nullable: true},
		{Name: "execution_saga_execution", Type: field.TypeInt, Unique: true},
	}
	// SagaExecutionsTable holds the schema information for the "saga_executions" table.
	SagaExecutionsTable = &schema.Table{
		Name:       "saga_executions",
		Columns:    SagaExecutionsColumns,
		PrimaryKey: []*schema.Column{SagaExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "saga_executions_executions_saga_execution",
				Columns:    []*schema.Column{SagaExecutionsColumns[3]},
				RefColumns: []*schema.Column{ExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SagaExecutionDataColumns holds the columns for the "saga_execution_data" table.
	SagaExecutionDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "transaction_history", Type: field.TypeJSON, Nullable: true},
		{Name: "compensation_history", Type: field.TypeJSON, Nullable: true},
		{Name: "last_transaction", Type: field.TypeTime, Nullable: true},
		{Name: "saga_execution_execution_data", Type: field.TypeInt, Unique: true},
	}
	// SagaExecutionDataTable holds the schema information for the "saga_execution_data" table.
	SagaExecutionDataTable = &schema.Table{
		Name:       "saga_execution_data",
		Columns:    SagaExecutionDataColumns,
		PrimaryKey: []*schema.Column{SagaExecutionDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "saga_execution_data_saga_executions_execution_data",
				Columns:    []*schema.Column{SagaExecutionDataColumns[4]},
				RefColumns: []*schema.Column{SagaExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SideEffectDataColumns holds the columns for the "side_effect_data" table.
	SideEffectDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "input", Type: field.TypeJSON, Nullable: true},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "entity_side_effect_data", Type: field.TypeInt, Unique: true},
	}
	// SideEffectDataTable holds the schema information for the "side_effect_data" table.
	SideEffectDataTable = &schema.Table{
		Name:       "side_effect_data",
		Columns:    SideEffectDataColumns,
		PrimaryKey: []*schema.Column{SideEffectDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "side_effect_data_entities_side_effect_data",
				Columns:    []*schema.Column{SideEffectDataColumns[3]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SideEffectExecutionsColumns holds the columns for the "side_effect_executions" table.
	SideEffectExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "result", Type: field.TypeJSON, Nullable: true},
		{Name: "execution_side_effect_execution", Type: field.TypeInt, Unique: true},
	}
	// SideEffectExecutionsTable holds the schema information for the "side_effect_executions" table.
	SideEffectExecutionsTable = &schema.Table{
		Name:       "side_effect_executions",
		Columns:    SideEffectExecutionsColumns,
		PrimaryKey: []*schema.Column{SideEffectExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "side_effect_executions_executions_side_effect_execution",
				Columns:    []*schema.Column{SideEffectExecutionsColumns[2]},
				RefColumns: []*schema.Column{ExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SideEffectExecutionDataColumns holds the columns for the "side_effect_execution_data" table.
	SideEffectExecutionDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "effect_time", Type: field.TypeTime, Nullable: true},
		{Name: "effect_metadata", Type: field.TypeJSON, Nullable: true},
		{Name: "execution_context", Type: field.TypeJSON, Nullable: true},
		{Name: "side_effect_execution_execution_data", Type: field.TypeInt, Unique: true},
	}
	// SideEffectExecutionDataTable holds the schema information for the "side_effect_execution_data" table.
	SideEffectExecutionDataTable = &schema.Table{
		Name:       "side_effect_execution_data",
		Columns:    SideEffectExecutionDataColumns,
		PrimaryKey: []*schema.Column{SideEffectExecutionDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "side_effect_execution_data_side_effect_executions_execution_data",
				Columns:    []*schema.Column{SideEffectExecutionDataColumns[4]},
				RefColumns: []*schema.Column{SideEffectExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// VersionsColumns holds the columns for the "versions" table.
	VersionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "version", Type: field.TypeInt},
		{Name: "data", Type: field.TypeJSON},
		{Name: "entity_versions", Type: field.TypeInt},
	}
	// VersionsTable holds the schema information for the "versions" table.
	VersionsTable = &schema.Table{
		Name:       "versions",
		Columns:    VersionsColumns,
		PrimaryKey: []*schema.Column{VersionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "versions_entities_versions",
				Columns:    []*schema.Column{VersionsColumns[3]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// WorkflowDataColumns holds the columns for the "workflow_data" table.
	WorkflowDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "duration", Type: field.TypeString, Nullable: true},
		{Name: "paused", Type: field.TypeBool, Default: false},
		{Name: "resumable", Type: field.TypeBool, Default: false},
		{Name: "retry_state", Type: field.TypeJSON},
		{Name: "retry_policy", Type: field.TypeJSON},
		{Name: "input", Type: field.TypeJSON, Nullable: true},
		{Name: "entity_workflow_data", Type: field.TypeInt, Unique: true},
	}
	// WorkflowDataTable holds the schema information for the "workflow_data" table.
	WorkflowDataTable = &schema.Table{
		Name:       "workflow_data",
		Columns:    WorkflowDataColumns,
		PrimaryKey: []*schema.Column{WorkflowDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "workflow_data_entities_workflow_data",
				Columns:    []*schema.Column{WorkflowDataColumns[7]},
				RefColumns: []*schema.Column{EntitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// WorkflowExecutionsColumns holds the columns for the "workflow_executions" table.
	WorkflowExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "execution_workflow_execution", Type: field.TypeInt, Unique: true},
	}
	// WorkflowExecutionsTable holds the schema information for the "workflow_executions" table.
	WorkflowExecutionsTable = &schema.Table{
		Name:       "workflow_executions",
		Columns:    WorkflowExecutionsColumns,
		PrimaryKey: []*schema.Column{WorkflowExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "workflow_executions_executions_workflow_execution",
				Columns:    []*schema.Column{WorkflowExecutionsColumns[1]},
				RefColumns: []*schema.Column{ExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// WorkflowExecutionDataColumns holds the columns for the "workflow_execution_data" table.
	WorkflowExecutionDataColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "workflow_execution_execution_data", Type: field.TypeInt, Unique: true},
	}
	// WorkflowExecutionDataTable holds the schema information for the "workflow_execution_data" table.
	WorkflowExecutionDataTable = &schema.Table{
		Name:       "workflow_execution_data",
		Columns:    WorkflowExecutionDataColumns,
		PrimaryKey: []*schema.Column{WorkflowExecutionDataColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "workflow_execution_data_workflow_executions_execution_data",
				Columns:    []*schema.Column{WorkflowExecutionDataColumns[3]},
				RefColumns: []*schema.Column{WorkflowExecutionsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		ActivityDataTable,
		ActivityExecutionsTable,
		ActivityExecutionDataTable,
		EntitiesTable,
		ExecutionsTable,
		HierarchiesTable,
		QueuesTable,
		RunsTable,
		SagaDataTable,
		SagaExecutionsTable,
		SagaExecutionDataTable,
		SideEffectDataTable,
		SideEffectExecutionsTable,
		SideEffectExecutionDataTable,
		VersionsTable,
		WorkflowDataTable,
		WorkflowExecutionsTable,
		WorkflowExecutionDataTable,
	}
)

func init() {
	ActivityDataTable.ForeignKeys[0].RefTable = EntitiesTable
	ActivityDataTable.Annotation = &entsql.Annotation{
		Table: "activity_data",
	}
	ActivityExecutionsTable.ForeignKeys[0].RefTable = ExecutionsTable
	ActivityExecutionsTable.Annotation = &entsql.Annotation{
		Table: "activity_executions",
	}
	ActivityExecutionDataTable.ForeignKeys[0].RefTable = ActivityExecutionsTable
	ActivityExecutionDataTable.Annotation = &entsql.Annotation{
		Table: "activity_execution_data",
	}
	EntitiesTable.ForeignKeys[0].RefTable = QueuesTable
	EntitiesTable.ForeignKeys[1].RefTable = RunsTable
	EntitiesTable.Annotation = &entsql.Annotation{
		Table: "entities",
	}
	ExecutionsTable.ForeignKeys[0].RefTable = EntitiesTable
	ExecutionsTable.Annotation = &entsql.Annotation{
		Table: "executions",
	}
	HierarchiesTable.ForeignKeys[0].RefTable = EntitiesTable
	HierarchiesTable.ForeignKeys[1].RefTable = EntitiesTable
	HierarchiesTable.ForeignKeys[2].RefTable = RunsTable
	HierarchiesTable.Annotation = &entsql.Annotation{
		Table: "hierarchies",
	}
	QueuesTable.Annotation = &entsql.Annotation{
		Table: "queues",
	}
	RunsTable.Annotation = &entsql.Annotation{
		Table: "runs",
	}
	SagaDataTable.ForeignKeys[0].RefTable = EntitiesTable
	SagaDataTable.Annotation = &entsql.Annotation{
		Table: "saga_data",
	}
	SagaExecutionsTable.ForeignKeys[0].RefTable = ExecutionsTable
	SagaExecutionsTable.Annotation = &entsql.Annotation{
		Table: "saga_executions",
	}
	SagaExecutionDataTable.ForeignKeys[0].RefTable = SagaExecutionsTable
	SagaExecutionDataTable.Annotation = &entsql.Annotation{
		Table: "saga_execution_data",
	}
	SideEffectDataTable.ForeignKeys[0].RefTable = EntitiesTable
	SideEffectDataTable.Annotation = &entsql.Annotation{
		Table: "side_effect_data",
	}
	SideEffectExecutionsTable.ForeignKeys[0].RefTable = ExecutionsTable
	SideEffectExecutionsTable.Annotation = &entsql.Annotation{
		Table: "side_effect_executions",
	}
	SideEffectExecutionDataTable.ForeignKeys[0].RefTable = SideEffectExecutionsTable
	SideEffectExecutionDataTable.Annotation = &entsql.Annotation{
		Table: "side_effect_execution_data",
	}
	VersionsTable.ForeignKeys[0].RefTable = EntitiesTable
	VersionsTable.Annotation = &entsql.Annotation{
		Table: "versions",
	}
	WorkflowDataTable.ForeignKeys[0].RefTable = EntitiesTable
	WorkflowDataTable.Annotation = &entsql.Annotation{
		Table: "workflow_data",
	}
	WorkflowExecutionsTable.ForeignKeys[0].RefTable = ExecutionsTable
	WorkflowExecutionsTable.Annotation = &entsql.Annotation{
		Table: "workflow_executions",
	}
	WorkflowExecutionDataTable.ForeignKeys[0].RefTable = WorkflowExecutionsTable
	WorkflowExecutionDataTable.Annotation = &entsql.Annotation{
		Table: "workflow_execution_data",
	}
}
