package driver

const (
	biTableVersion        = "0.0.1-bitable"
	BitableSchema         = "bitable"
	DefaultPageSize int64 = 50
)

const (
	maxLoopTimes = 100
	loadOneTime  = "ONE_TIME"
)

type ViewType string

const (
	ViewTypeGrid    ViewType = "grid"    //  表格视图
	ViewTypeKanban  ViewType = "kanban"  //  看板视图
	ViewTypeGantt   ViewType = "gantt"   //  甘特视图
	ViewTypeGallery ViewType = "gallery" //  画册视图
	ViewTypeForm    ViewType = "form"    //  表单视图
)

type FieldType int64

const (
	FieldTypeText              FieldType = 1    //  多行文本
	FieldTypeNumber            FieldType = 2    //  数字
	FieldTypeSelect            FieldType = 3    //  单选
	FieldTypeMultipleSelect    FieldType = 4    //  多选
	FieldTypeDate              FieldType = 5    //  日期
	FieldTypeCheckbox          FieldType = 7    //  复选框
	FieldTypePerson            FieldType = 11   //  人员
	FieldTypeLink              FieldType = 15   //  超链接
	FieldTypeAttachment        FieldType = 17   //  附件
	FieldTypeOneWayAssociation FieldType = 18   //  单向关联
	FieldTypeReferenceLookup   FieldType = 19   //  引用查找
	FieldTypeFormula           FieldType = 20   //  公式
	FieldTypeTwoWayAssociation FieldType = 21   //  双向关联
	FieldTypeCreateTime        FieldType = 1001 //  创建时间
	FieldTypeUpdateTime        FieldType = 1002 //  最后更新时间
	FieldTypeFounder           FieldType = 1003 //  创建人
	FieldTypeModifier          FieldType = 1004 //  修改人
)

type RecordKey string

const (
	RecordKeyPerson      = "persons"
	RecordKeyUrl         = "url"
	RecordKeyOptions     = "options"
	RecordKeyAttachments = "attachments"
)
