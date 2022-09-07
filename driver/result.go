package driver

import (
	"github.com/luw2007/bitable-mysql-driver/internal/lark"
)

// Person a field value for person.
type Person lark.FieldPerson

// Attachment a field value for attachment.
type Attachment lark.FieldAttachment

// URL a field value for link.
type URL lark.FieldURL

type RecordUrl URL
type RecordPersons []Person
type RecordAttachments []Attachment
type RecordOptions []string

// AppMeta a App meta.
type AppMeta lark.AppMeta

// Table a table meta.
type Table lark.Table

// Record a record meta.
type Record lark.Record

// View a view meta.
type View lark.View

// Field a field info
type Field struct {
	FieldID   string         `json:"field_id,omitempty"`
	FieldName string         `json:"field_name,omitempty"`
	Type      int64          `json:"type,omitempty"`
	Property  *FieldProperty `json:"property,omitempty"`
}

// FieldProperty a field property. Different types have different values.
type FieldProperty struct {
	Options    []*FieldOption `json:"options,omitempty"`
	Formatter  string         `json:"formatter,omitempty"`
	DateFormat string         `json:"date_format,omitempty"`
	TimeFormat string         `json:"time_format,omitempty"`
	AutoFill   bool           `json:"auto_fill,omitempty"`
	Multiple   bool           `json:"multiple,omitempty"`
	TableId    string         `json:"table_id,omitempty"`
	ViewId     string         `json:"view_id,omitempty"`
	Fields     []string       `json:"fields,omitempty"`
}

// FieldOption when field.Type can be select.
type FieldOption struct {
	Name string `json:"name,omitempty"`
	Id   string `json:"id,omitempty"`
}

type bitableResult struct {
	lastInsertID int64
	rowsAffected int64
	err          error
}

func (b bitableResult) LastInsertId() (int64, error) {
	return b.lastInsertID, b.err
}

func (b bitableResult) RowsAffected() (int64, error) {
	return b.rowsAffected, b.err
}
