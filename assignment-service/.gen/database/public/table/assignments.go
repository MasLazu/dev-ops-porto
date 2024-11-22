//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Assignments = newAssignmentsTable("public", "assignments", "")

type assignmentsTable struct {
	postgres.Table

	// Columns
	ID          postgres.ColumnInteger
	UserID      postgres.ColumnString
	Title       postgres.ColumnString
	Note        postgres.ColumnString
	DueDate     postgres.ColumnTimestampz
	IsCompleted postgres.ColumnBool
	IsImportant postgres.ColumnBool
	CreatedAt   postgres.ColumnTimestamp
	UpdatedAt   postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type AssignmentsTable struct {
	assignmentsTable

	EXCLUDED assignmentsTable
}

// AS creates new AssignmentsTable with assigned alias
func (a AssignmentsTable) AS(alias string) *AssignmentsTable {
	return newAssignmentsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new AssignmentsTable with assigned schema name
func (a AssignmentsTable) FromSchema(schemaName string) *AssignmentsTable {
	return newAssignmentsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new AssignmentsTable with assigned table prefix
func (a AssignmentsTable) WithPrefix(prefix string) *AssignmentsTable {
	return newAssignmentsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new AssignmentsTable with assigned table suffix
func (a AssignmentsTable) WithSuffix(suffix string) *AssignmentsTable {
	return newAssignmentsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newAssignmentsTable(schemaName, tableName, alias string) *AssignmentsTable {
	return &AssignmentsTable{
		assignmentsTable: newAssignmentsTableImpl(schemaName, tableName, alias),
		EXCLUDED:         newAssignmentsTableImpl("", "excluded", ""),
	}
}

func newAssignmentsTableImpl(schemaName, tableName, alias string) assignmentsTable {
	var (
		IDColumn          = postgres.IntegerColumn("id")
		UserIDColumn      = postgres.StringColumn("user_id")
		TitleColumn       = postgres.StringColumn("title")
		NoteColumn        = postgres.StringColumn("note")
		DueDateColumn     = postgres.TimestampzColumn("due_date")
		IsCompletedColumn = postgres.BoolColumn("is_completed")
		IsImportantColumn = postgres.BoolColumn("is_important")
		CreatedAtColumn   = postgres.TimestampColumn("created_at")
		UpdatedAtColumn   = postgres.TimestampColumn("updated_at")
		allColumns        = postgres.ColumnList{IDColumn, UserIDColumn, TitleColumn, NoteColumn, DueDateColumn, IsCompletedColumn, IsImportantColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns    = postgres.ColumnList{UserIDColumn, TitleColumn, NoteColumn, DueDateColumn, IsCompletedColumn, IsImportantColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return assignmentsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:          IDColumn,
		UserID:      UserIDColumn,
		Title:       TitleColumn,
		Note:        NoteColumn,
		DueDate:     DueDateColumn,
		IsCompleted: IsCompletedColumn,
		IsImportant: IsImportantColumn,
		CreatedAt:   CreatedAtColumn,
		UpdatedAt:   UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}