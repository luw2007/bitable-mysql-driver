
```golang

type ViewType string

const (
	Grid    ViewType = "grid"    //  表格视图
	Kanban  ViewType = "kanban"  //  看板视图
	Gantt   ViewType = "gantt"   //  甘特视图
	Gallery ViewType = "gallery" //  画册视图
	Form    ViewType = "form"    //  表单视图
)

type FieldType int

const (
	Text              FieldType = 1    //  多行文本
	Number            FieldType = 2    //  数字
	Select            FieldType = 3    //  单选
	MultipleSelect    FieldType = 4    //  多选
	Date              FieldType = 5    //  日期
	Checkbox          FieldType = 7    //  复选框
	Person            FieldType = 11   //  人员
	Link              FieldType = 15   //  超链接
	Attachment        FieldType = 17   //  附件
	OneWayAssociation FieldType = 18   //  单向关联
	ReferenceLookup   FieldType = 19   //  引用查找
	Formula           FieldType = 20   //  公式
	TwoWayAssociation FieldType = 21   //  双向关联
	CreateTime        FieldType = 1001 //  创建时间
	UpdateTime        FieldType = 1002 //  最后更新时间
	Founder           FieldType = 1003 //  创建人
	Modifier          FieldType = 1004 //  修改人
)
```

https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/bitable-v1/app-table-field/guide