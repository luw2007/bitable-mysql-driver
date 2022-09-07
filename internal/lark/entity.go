package lark

type FieldURL struct {
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

type AppMeta struct {
	AppToken string `json:"app_token,omitempty"` // 多维表格的 app_token
	Name     string `json:"name,omitempty"`      // 多维表格的名字
	Revision int64  `json:"revision,omitempty"`  // 多维表格的版本号
}

type Table struct {
	TableID  string `json:"table_id,omitempty"` // 数据表 id
	Revision int64  `json:"revision,omitempty"` // 数据表的版本号
	Name     string `json:"name,omitempty"`     // 数据表名字
}

type View struct {
	ViewID   string `json:"view_id,omitempty"`   // 视图Id
	ViewName string `json:"view_name,omitempty"` // 视图名字
	ViewType string `json:"view_type,omitempty"` // 视图类型
}

type Record struct {
	RecordID string                 `json:"record_id,omitempty"` // 记录 id
	Fields   map[string]interface{} `json:"fields,omitempty"`    // 记录字段
}

type Field struct {
	FieldID   string         `json:"field_id,omitempty"`
	FieldName string         `json:"field_name,omitempty"`
	Type      int64          `json:"type,omitempty"`
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

type PageList struct {
	HasMore   bool          `json:"has_more,omitempty"`   // 是否还有更多项
	PageToken string        `json:"page_token,omitempty"` // 分页标记，当 has_more 为 true 时，会同时返回新的 page_token，否则不返回 page_token
	Total     int64         `json:"total,omitempty"`      // 总数
	Items     []interface{} `json:"items,omitempty"`      // 字段信息
}

func (l *PageList) Merge(page *PageList) {
	if page.Total > 0 {
		l.PageToken = page.PageToken
		l.HasMore = l.HasMore || page.HasMore
		l.Total += page.Total
		l.Items = append(l.Items, page.Items...)
	}
}

type Sheet struct {
	SheetID        string               `json:"sheetId,omitempty"`        // sheet 的 id
	Title          string               `json:"title,omitempty"`          // sheet 的标题
	Index          int64                `json:"index,omitempty"`          // sheet 的位置
	RowCount       int64                `json:"rowCount,omitempty"`       // sheet 的最大行数
	ColumnCount    int64                `json:"columnCount,omitempty"`    // sheet 的最大列数
	FrozenRowCount int64                `json:"frozenRowCount,omitempty"` // 该 sheet 的冻结行数，小于等于 sheet 的最大行数，0表示未设置冻结
	FrozenColCount int64                `json:"frozenColCount,omitempty"` // 该 sheet 的冻结列数，小于等于 sheet 的最大列数，0表示未设置冻结
	Merges         []*SheetMerge        `json:"merges,omitempty"`         // 该 sheet 中合并单元格的范围
	ProtectedRange *SheetProtectedRange `json:"protectedRange,omitempty"` // 该 sheet 中保护范围
	BlockInfo      *SheetBlockInfo      `json:"blockInfo,omitempty"`      // 若含有该字段，则此工作表不为表格
}

// SheetMerge ...
type SheetMerge struct {
	StartRowIndex    int64 `json:"startRowIndex,omitempty"`    // 合并单元格范围的开始行下标，index 从 0 开始
	StartColumnIndex int64 `json:"startColumnIndex,omitempty"` // 合并单元格范围的开始列下标，index 从 0 开始
	RowCount         int64 `json:"rowCount,omitempty"`         // 合并单元格范围的行数量
	ColumnCount      int64 `json:"columnCount,omitempty"`      // 合并单元格范围的列数量
}

// SheetProtectedRange ...
type SheetProtectedRange struct {
	Dimension *SheetProtectedRangeDimension `json:"dimension,omitempty"` // 保护行列的信息，如果为保护工作表，则该字段为空
	ProtectID string                        `json:"protectId,omitempty"` // 保护范围ID
	LockInfo  string                        `json:"lockInfo,omitempty"`  // 保护说明
	SheetID   string                        `json:"sheetId,omitempty"`   // 保护工作表 ID
}

// SheetProtectedRangeDimension ...
type SheetProtectedRangeDimension struct {
	StartIndex     int64  `json:"startIndex,omitempty"`     // 保护行列的起始位置，位置从1开始
	EndIndex       int64  `json:"endIndex,omitempty"`       // 保护行列的结束位置，位置从1开始
	MajorDimension string `json:"majorDimension,omitempty"` // 若为ROWS，则为保护行；为COLUMNS，则为保护列
	SheetID        string `json:"sheetId,omitempty"`        // 保护范围所在工作表 ID
}

// SheetBlockInfo ...
type SheetBlockInfo struct {
	BlockToken string `json:"blockToken,omitempty"` // block的token
	BlockType  string `json:"blockType,omitempty"`  // block的类型
}
