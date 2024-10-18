// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// ActivitiesColumns holds the columns for the "activities" table.
	ActivitiesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "identity", Type: field.TypeString},
		{Name: "step_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled"}, Default: "Pending"},
		{Name: "handler_name", Type: field.TypeString},
		{Name: "input", Type: field.TypeJSON},
		{Name: "retry_policy", Type: field.TypeJSON, Nullable: true},
		{Name: "timeout", Type: field.TypeTime, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
	}
	// ActivitiesTable holds the schema information for the "activities" table.
	ActivitiesTable = &schema.Table{
		Name:       "activities",
		Columns:    ActivitiesColumns,
		PrimaryKey: []*schema.Column{ActivitiesColumns[0]},
	}
	// ActivityExecutionsColumns holds the columns for the "activity_executions" table.
	ActivityExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "run_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Retried"}, Default: "Pending"},
		{Name: "attempt", Type: field.TypeInt, Default: 1},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "activity_executions", Type: field.TypeString},
	}
	// ActivityExecutionsTable holds the schema information for the "activity_executions" table.
	ActivityExecutionsTable = &schema.Table{
		Name:       "activity_executions",
		Columns:    ActivityExecutionsColumns,
		PrimaryKey: []*schema.Column{ActivityExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "activity_executions_activities_executions",
				Columns:    []*schema.Column{ActivityExecutionsColumns[8]},
				RefColumns: []*schema.Column{ActivitiesColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// ExecutionRelationshipsColumns holds the columns for the "execution_relationships" table.
	ExecutionRelationshipsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "run_id", Type: field.TypeString},
		{Name: "parent_entity_id", Type: field.TypeString},
		{Name: "child_entity_id", Type: field.TypeString},
		{Name: "parent_id", Type: field.TypeString},
		{Name: "child_id", Type: field.TypeString},
		{Name: "parent_type", Type: field.TypeEnum, Enums: []string{"workflow", "activity", "saga", "side_effect", "yield", "signal"}},
		{Name: "child_type", Type: field.TypeEnum, Enums: []string{"workflow", "activity", "saga", "side_effect", "yield", "signal"}},
		{Name: "parent_step_id", Type: field.TypeString},
		{Name: "child_step_id", Type: field.TypeString},
	}
	// ExecutionRelationshipsTable holds the schema information for the "execution_relationships" table.
	ExecutionRelationshipsTable = &schema.Table{
		Name:       "execution_relationships",
		Columns:    ExecutionRelationshipsColumns,
		PrimaryKey: []*schema.Column{ExecutionRelationshipsColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "executionrelationship_parent_id_child_id",
				Unique:  true,
				Columns: []*schema.Column{ExecutionRelationshipsColumns[4], ExecutionRelationshipsColumns[5]},
			},
		},
	}
	// FeatureFlagVersionsColumns holds the columns for the "feature_flag_versions" table.
	FeatureFlagVersionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "workflow_type", Type: field.TypeString},
		{Name: "workflow_id", Type: field.TypeString},
		{Name: "change_id", Type: field.TypeString},
		{Name: "version", Type: field.TypeInt},
	}
	// FeatureFlagVersionsTable holds the schema information for the "feature_flag_versions" table.
	FeatureFlagVersionsTable = &schema.Table{
		Name:       "feature_flag_versions",
		Columns:    FeatureFlagVersionsColumns,
		PrimaryKey: []*schema.Column{FeatureFlagVersionsColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "featureflagversion_workflow_type_change_id",
				Unique:  true,
				Columns: []*schema.Column{FeatureFlagVersionsColumns[1], FeatureFlagVersionsColumns[3]},
			},
		},
	}
	// RunsColumns holds the columns for the "runs" table.
	RunsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "run_id", Type: field.TypeString, Unique: true},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"workflow", "activity"}},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "run_workflow", Type: field.TypeString, Nullable: true},
		{Name: "run_activity", Type: field.TypeString, Nullable: true},
	}
	// RunsTable holds the schema information for the "runs" table.
	RunsTable = &schema.Table{
		Name:       "runs",
		Columns:    RunsColumns,
		PrimaryKey: []*schema.Column{RunsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "runs_workflows_workflow",
				Columns:    []*schema.Column{RunsColumns[4]},
				RefColumns: []*schema.Column{WorkflowsColumns[0]},
				OnDelete:   schema.SetNull,
			},
			{
				Symbol:     "runs_activities_activity",
				Columns:    []*schema.Column{RunsColumns[5]},
				RefColumns: []*schema.Column{ActivitiesColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// SagasColumns holds the columns for the "sagas" table.
	SagasColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "run_id", Type: field.TypeString},
		{Name: "step_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Compensating", "Compensated"}, Default: "Pending"},
		{Name: "saga_definition", Type: field.TypeJSON},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
	}
	// SagasTable holds the schema information for the "sagas" table.
	SagasTable = &schema.Table{
		Name:       "sagas",
		Columns:    SagasColumns,
		PrimaryKey: []*schema.Column{SagasColumns[0]},
	}
	// SagaExecutionsColumns holds the columns for the "saga_executions" table.
	SagaExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "handler_name", Type: field.TypeString},
		{Name: "step_type", Type: field.TypeEnum, Enums: []string{"Transaction", "Compensation"}},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed"}, Default: "Pending"},
		{Name: "sequence", Type: field.TypeInt},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "completed_at", Type: field.TypeTime, Nullable: true},
		{Name: "saga_steps", Type: field.TypeString},
	}
	// SagaExecutionsTable holds the schema information for the "saga_executions" table.
	SagaExecutionsTable = &schema.Table{
		Name:       "saga_executions",
		Columns:    SagaExecutionsColumns,
		PrimaryKey: []*schema.Column{SagaExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "saga_executions_sagas_steps",
				Columns:    []*schema.Column{SagaExecutionsColumns[8]},
				RefColumns: []*schema.Column{SagasColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SideEffectsColumns holds the columns for the "side_effects" table.
	SideEffectsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "identity", Type: field.TypeString},
		{Name: "step_id", Type: field.TypeString},
		{Name: "handler_name", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed"}, Default: "Pending"},
		{Name: "retry_policy", Type: field.TypeJSON, Nullable: true},
		{Name: "timeout", Type: field.TypeTime, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
	}
	// SideEffectsTable holds the schema information for the "side_effects" table.
	SideEffectsTable = &schema.Table{
		Name:       "side_effects",
		Columns:    SideEffectsColumns,
		PrimaryKey: []*schema.Column{SideEffectsColumns[0]},
	}
	// SideEffectExecutionsColumns holds the columns for the "side_effect_executions" table.
	SideEffectExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed"}, Default: "Pending"},
		{Name: "attempt", Type: field.TypeInt, Default: 1},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "side_effect_executions", Type: field.TypeString},
	}
	// SideEffectExecutionsTable holds the schema information for the "side_effect_executions" table.
	SideEffectExecutionsTable = &schema.Table{
		Name:       "side_effect_executions",
		Columns:    SideEffectExecutionsColumns,
		PrimaryKey: []*schema.Column{SideEffectExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "side_effect_executions_side_effects_executions",
				Columns:    []*schema.Column{SideEffectExecutionsColumns[7]},
				RefColumns: []*schema.Column{SideEffectsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// SignalsColumns holds the columns for the "signals" table.
	SignalsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "step_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled"}, Default: "Pending"},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "consumed", Type: field.TypeBool, Default: false},
	}
	// SignalsTable holds the schema information for the "signals" table.
	SignalsTable = &schema.Table{
		Name:       "signals",
		Columns:    SignalsColumns,
		PrimaryKey: []*schema.Column{SignalsColumns[0]},
	}
	// SignalExecutionsColumns holds the columns for the "signal_executions" table.
	SignalExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "run_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled"}, Default: "Pending"},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "signal_executions", Type: field.TypeString},
	}
	// SignalExecutionsTable holds the schema information for the "signal_executions" table.
	SignalExecutionsTable = &schema.Table{
		Name:       "signal_executions",
		Columns:    SignalExecutionsColumns,
		PrimaryKey: []*schema.Column{SignalExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "signal_executions_signals_executions",
				Columns:    []*schema.Column{SignalExecutionsColumns[7]},
				RefColumns: []*schema.Column{SignalsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// WorkflowsColumns holds the columns for the "workflows" table.
	WorkflowsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "step_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Retried", "Cancelled"}, Default: "Pending"},
		{Name: "identity", Type: field.TypeString},
		{Name: "handler_name", Type: field.TypeString},
		{Name: "input", Type: field.TypeJSON},
		{Name: "retry_policy", Type: field.TypeJSON, Nullable: true},
		{Name: "is_paused", Type: field.TypeBool, Default: false},
		{Name: "is_ready", Type: field.TypeBool, Default: false},
		{Name: "timeout", Type: field.TypeTime, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
	}
	// WorkflowsTable holds the schema information for the "workflows" table.
	WorkflowsTable = &schema.Table{
		Name:       "workflows",
		Columns:    WorkflowsColumns,
		PrimaryKey: []*schema.Column{WorkflowsColumns[0]},
	}
	// WorkflowExecutionsColumns holds the columns for the "workflow_executions" table.
	WorkflowExecutionsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true},
		{Name: "run_id", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"Pending", "Running", "Completed", "Failed", "Paused", "Retried", "Cancelled"}, Default: "Pending"},
		{Name: "output", Type: field.TypeJSON, Nullable: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "started_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "workflow_executions", Type: field.TypeString},
	}
	// WorkflowExecutionsTable holds the schema information for the "workflow_executions" table.
	WorkflowExecutionsTable = &schema.Table{
		Name:       "workflow_executions",
		Columns:    WorkflowExecutionsColumns,
		PrimaryKey: []*schema.Column{WorkflowExecutionsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "workflow_executions_workflows_executions",
				Columns:    []*schema.Column{WorkflowExecutionsColumns[7]},
				RefColumns: []*schema.Column{WorkflowsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		ActivitiesTable,
		ActivityExecutionsTable,
		ExecutionRelationshipsTable,
		FeatureFlagVersionsTable,
		RunsTable,
		SagasTable,
		SagaExecutionsTable,
		SideEffectsTable,
		SideEffectExecutionsTable,
		SignalsTable,
		SignalExecutionsTable,
		WorkflowsTable,
		WorkflowExecutionsTable,
	}
)

func init() {
	ActivityExecutionsTable.ForeignKeys[0].RefTable = ActivitiesTable
	RunsTable.ForeignKeys[0].RefTable = WorkflowsTable
	RunsTable.ForeignKeys[1].RefTable = ActivitiesTable
	SagaExecutionsTable.ForeignKeys[0].RefTable = SagasTable
	SideEffectExecutionsTable.ForeignKeys[0].RefTable = SideEffectsTable
	SignalExecutionsTable.ForeignKeys[0].RefTable = SignalsTable
	WorkflowExecutionsTable.ForeignKeys[0].RefTable = WorkflowsTable
}
