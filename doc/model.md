# bitable model

```golang
type App struct {
	AppToken string `json:"app_token,omitempty"`
	Name     string `json:"name,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

type AppTable struct {
	TableId  string `json:"table_id,omitempty"`
	Revision int    `json:"revision,omitempty"`
	Name     string `json:"name,omitempty"`
}

type AppTableView struct {
	ViewId   string `json:"view_id,omitempty"`
	ViewName string `json:"view_name,omitempty"`
	ViewType string `json:"view_type,omitempty"`
}

type AppTableRecord struct {
	AppToken        string                 `json:"app_token,omitempty"`
	TableId         string                 `json:"table_id,omitempty"`
	RecordId        string                 `json:"record_id,omitempty"`
	Fields          map[string]interface{} `json:"fields,omitempty"`
	ForceSendFields []string               `json:"force_send_fields,omitempty"`
}

type AppTableField struct {
	FieldId   string         `json:"field_id,omitempty"`
	FieldName string         `json:"field_name,omitempty"`
	Type      int            `json:"type,omitempty"`
	Property  *FieldProperty `json:"property,omitempty"`
}

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

type FieldUrl struct {
	Text string `json:"text,omitempty"`
	Link string `json:"link,omitempty"`
}

type FieldPerson struct {
	Id     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	EnName string `json:"en_name,omitempty"`
	Email  string `json:"email,omitempty"`
}

type FieldAttachment struct {
	FileToken string `json:"file_token,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Size      int    `json:"size,omitempty"`
	Url       string `json:"url,omitempty"`
	TmpUrl    string `json:"tmp_url,omitempty"`
}

type FieldOption struct {
	Name string `json:"name,omitempty"`
	Id   string `json:"id,omitempty"`
}

type RecordUrl FieldUrl
type RecordPersons []FieldPerson
type RecordAttachments []FieldAttachment
type RecordOptions []string


```