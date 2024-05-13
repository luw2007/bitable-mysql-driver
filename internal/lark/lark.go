package lark

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/chyroc/lark"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
)

type BiTable struct {
	*lark.Lark
	appID string
}

var (
	larkStoreMap = sync.Map{}
)

const (
	DefaultPageSize = 50
)

func getStoreSlot(appID string) lark.Store {
	s, _ := larkStoreMap.LoadOrStore(appID, lark.NewStoreMemory())
	return s.(lark.Store)
}

func NewLarkClient(appID, appSecret, apiDomain, logLevel string, timeout time.Duration) *BiTable {
	logger := &larkLogger{logger: logrus.StandardLogger()}
	return &BiTable{
		Lark: lark.New(lark.WithAppCredential(appID, appSecret),
			lark.WithStore(getStoreSlot(appID)),
			lark.WithLogger(logger, getLarkLogLevel(logLevel)),
			lark.WithOpenBaseURL(apiDomain),
			lark.WithTimeout(timeout),
		),
		appID: appID,
	}
}

type contextKey string

const (
	userTokenContextKey contextKey = "user_access_token"
)

func WithUserToken(ctx context.Context, userToken string) context.Context {
	return context.WithValue(ctx, userTokenContextKey, userToken)
}

func (b *BiTable) GetApp(ctx context.Context, appToken string) (*AppMeta, error) {
	req := &lark.GetBitableMetaReq{
		AppToken: appToken,
	}
	resp, _, err := b.Bitable.GetBitableMeta(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	app := resp.App
	return &AppMeta{
		AppToken: app.AppToken,
		Name:     app.Name,
		Revision: app.Revision,
	}, err
}

func (b *BiTable) CreateTable(ctx context.Context, appToken, tableName string) (string, error) {
	req := &lark.CreateBitableTableReq{
		AppToken: appToken,
		Table: &lark.CreateBitableTableReqTable{
			Name: &tableName,
		},
	}
	resp, _, err := b.Bitable.CreateBitableTable(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return "", err
	}
	return resp.TableID, nil
}

func (b *BiTable) DropTable(ctx context.Context, appToken, tableID string) error {
	_, _, err := b.Bitable.DeleteBitableTable(ctx, &lark.DeleteBitableTableReq{
		AppToken: appToken,
		TableID:  tableID,
	})
	if err != nil {
		return fmt.Errorf("bitable %w", err)
	}
	return nil
}

func (b *BiTable) ListTable(ctx context.Context, appToken string, pageToken string, pageSize int64) (*PageList, error) {
	req := &lark.GetBitableTableListReq{
		PageSize:  &pageSize,
		PageToken: &pageToken,
		AppToken:  appToken,
	}
	resp, _, err := b.Bitable.GetBitableTableList(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildPageList(resp)
}

func (b *BiTable) ListALLTable(ctx context.Context, appToken string) (*PageList, error) {
	pageToken := ""
	page := PageList{}
	for {
		curPage, err := b.ListTable(ctx, appToken, pageToken, DefaultPageSize)
		if err != nil {
			return nil, err
		}
		page.Merge(curPage)
		if !curPage.HasMore {
			break
		}
	}
	return &page, nil
}

func (b *BiTable) CreateView(ctx context.Context, appToken, table, viewName string, viewType string) (*View, error) {
	req := &lark.CreateBitableViewReq{
		AppToken: appToken,
		TableID:  table,
		ViewName: viewName,
		ViewType: &viewType,
	}
	resp, _, err := b.Bitable.CreateBitableView(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	view := resp.View
	return &View{
		ViewID:   view.ViewID,
		ViewName: view.ViewName,
		ViewType: view.ViewType,
	}, nil
}

func (b *BiTable) DropView(ctx context.Context, appToken string, table string, view string) error {
	req := &lark.DeleteBitableViewReq{
		AppToken: appToken,
		TableID:  table,
		ViewID:   view,
	}
	_, _, err := b.Bitable.DeleteBitableView(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return fmt.Errorf("bitable %w", err)
	}
	return nil
}

func (b *BiTable) ListViews(ctx context.Context, appToken string, table string, pageToken string, pageSize int64) (*PageList, error) {
	req := &lark.GetBitableViewListReq{
		PageSize:  &pageSize,
		PageToken: &pageToken,
		AppToken:  appToken,
		TableID:   table,
	}
	resp, _, err := b.Bitable.GetBitableViewList(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}

	return buildPageList(resp)
}

func (b *BiTable) AddField(ctx context.Context, appToken, table, fieldName string, fieldType int64, property string) (*Field, error) {
	req := &lark.CreateBitableFieldReq{
		AppToken:  appToken,
		TableID:   table,
		FieldName: fieldName,
		Type:      fieldType,
	}
	if property != "" {
		var p lark.CreateBitableFieldReqProperty
		if err := json.Unmarshal([]byte(property), &p); err != nil {
			return nil, fmt.Errorf("bitable %w", err)
		}
		req.Property = &p
	}

	resp, _, err := b.Bitable.CreateBitableField(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildField(resp.Field)
}

func (b *BiTable) DeleteField(ctx context.Context, appToken, table, fieldID string) (bool, error) {
	req := &lark.DeleteBitableFieldReq{
		AppToken: appToken,
		TableID:  table,
		FieldID:  fieldID,
	}
	resp, _, err := b.Bitable.DeleteBitableField(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return false, err
	}
	return resp.Deleted, nil
}

func (b *BiTable) UpdateField(ctx context.Context, appToken, table, fieldID, fieldName string, fieldType int64, property string) (*Field, error) {
	req := &lark.UpdateBitableFieldReq{
		AppToken:  appToken,
		TableID:   table,
		FieldID:   fieldID,
		FieldName: fieldName,
		Type:      fieldType,
	}
	if property != "" {
		var p lark.UpdateBitableFieldReqProperty
		if err := json.Unmarshal([]byte(property), &p); err != nil {
			return nil, fmt.Errorf("bitable %w", err)
		}
		req.Property = &p
	}
	resp, _, err := b.Bitable.UpdateBitableField(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildField(resp.Field)
}

func (b *BiTable) ListFields(ctx context.Context, appToken, table, view, pageToken string, pageSize int64) (*PageList, error) {
	req := &lark.GetBitableFieldListReq{
		ViewID:    &view,
		PageToken: &pageToken,
		PageSize:  &pageSize,
		AppToken:  appToken,
		TableID:   table,
	}
	resp, _, err := b.Bitable.GetBitableFieldList(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildPageList(resp)
}

func (b *BiTable) InsertRecords(ctx context.Context, appToken, table string, data []map[string]interface{}) ([]*Record, error) {
	records := make([]*lark.BatchCreateBitableRecordReqRecord, 0, len(data))
	for _, fields := range data {
		records = append(records, &lark.BatchCreateBitableRecordReqRecord{
			Fields: fields,
		})
	}
	req := &lark.BatchCreateBitableRecordReq{
		AppToken: appToken,
		TableID:  table,
		Records:  records,
	}
	resp, _, err := b.Bitable.BatchCreateBitableRecord(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildRecords(resp.Records)
}

func (b *BiTable) DeleteRecord(ctx context.Context, appToken, table, recordID string) (bool, error) {
	req := &lark.DeleteBitableRecordReq{
		AppToken: appToken,
		TableID:  table,
		RecordID: recordID,
	}
	resp, _, err := b.Bitable.DeleteBitableRecord(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return false, err
	}
	return resp.Deleted, nil
}

func (b *BiTable) UpdateRecord(ctx context.Context, appToken, table, recordID string, fields map[string]interface{}) (*Record, error) {
	req := &lark.UpdateBitableRecordReq{
		AppToken: appToken,
		TableID:  table,
		RecordID: recordID,
		Fields:   fields,
	}
	resp, _, err := b.Bitable.UpdateBitableRecord(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	record := resp.Record
	return &Record{
		RecordID: record.RecordID,
		Fields:   record.Fields,
	}, nil
}

func (b *BiTable) UpdateRecords(ctx context.Context, appToken, table string, data map[string]map[string]interface{}) ([]*Record, error) {
	records := make([]*lark.BatchUpdateBitableRecordReqRecord, 0, len(data))
	for recordID, fields := range data {
		records = append(records, &lark.BatchUpdateBitableRecordReqRecord{
			RecordID: &recordID,
			Fields:   fields,
		})
	}
	req := &lark.BatchUpdateBitableRecordReq{
		AppToken: appToken,
		TableID:  table,
		Records:  records,
	}
	resp, _, err := b.Bitable.BatchUpdateBitableRecord(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildRecords(resp.Records)
}

func (b *BiTable) GetRecord(ctx context.Context, appToken, table, recordID string) (*Record, error) {
	req := &lark.GetBitableRecordReq{
		AppToken: appToken,
		TableID:  table,
		RecordID: recordID,
	}
	resp, _, err := b.Bitable.GetBitableRecord(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	record := resp.Record
	return &Record{
		RecordID: record.RecordID,
		Fields:   record.Fields,
	}, nil
}

func (b *BiTable) ListRecords(ctx context.Context, appToken, table, view, fieldNames, filter, sort, pageToken string, pageSize int64) (*PageList, error) {
	req := &lark.GetBitableRecordListReq{
		ViewID:     &view,
		Filter:     &filter,
		Sort:       &sort,
		FieldNames: &fieldNames,
		PageToken:  &pageToken,
		PageSize:   &pageSize,
		AppToken:   appToken,
		TableID:    table,
	}
	resp, _, err := b.Bitable.GetBitableRecordList(ctx, req, buildMethodOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("bitable %w", err)
	}
	return buildPageList(resp)
}

func (b *BiTable) ListAllFields(ctx context.Context, appToken, table string) (*PageList, error) {
	pageToken := ""
	page := PageList{}
	for {
		curPage, err := b.ListFields(ctx, appToken, table, "", pageToken, DefaultPageSize)
		if err != nil {
			return nil, err
		}
		page.Merge(curPage)
		if !curPage.HasMore {
			break
		}
	}
	return &page, nil
}

func buildField(data interface{}) (*Field, error) {
	field := Field{}
	err := copier.Copy(&field, data)
	if err != nil {
		return nil, err
	}
	return &field, err
}

func buildRecords(data interface{}) ([]*Record, error) {
	res := []*Record{}
	err := copier.Copy(&res, data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func buildPageList(resp interface{}) (*PageList, error) {
	getValue := reflect.ValueOf(resp).Elem()
	items := getValue.FieldByName("Items")
	page := &PageList{
		HasMore:   getValue.FieldByName("HasMore").Bool(),
		PageToken: getValue.FieldByName("PageToken").String(),
		Total:     getValue.FieldByName("Total").Int(),
		Items:     make([]interface{}, items.Len()),
	}
	for i := 0; i < items.Len(); i++ {
		switch item := items.Index(i).Interface().(type) {
		case *lark.GetBitableViewListRespItem:
			view := View{}
			err := copier.Copy(&view, item)
			if err != nil {
				return nil, err
			}
			page.Items[i] = &view
		case *lark.GetBitableTableListRespItem:
			table := Table{}
			err := copier.Copy(&table, item)
			if err != nil {
				return nil, err
			}
			page.Items[i] = &table
		case *lark.GetBitableFieldListRespItem:
			field := Field{}
			err := copier.Copy(&field, item)
			if err != nil {
				return nil, err
			}
			page.Items[i] = &field
		case *lark.GetBitableRecordListRespItem:
			record := Record{}
			err := copier.Copy(&record, item)
			if err != nil {
				return nil, err
			}
			page.Items[i] = &record
		}
	}
	return page, nil
}

func buildMethodOptions(ctx context.Context) []lark.MethodOptionFunc {
	options := make([]lark.MethodOptionFunc, 0)
	if userToken, ok := ctx.Value(userTokenContextKey).(string); ok && userToken != "" {
		options = append(options, lark.WithUserAccessToken(userToken))
	}
	return options
}

type Sheets struct {
	*lark.Lark
	appID string
}

func (s *Sheets) GetAPP(ctx context.Context, shetToken string) ([]*Sheet, error) {
	req := &lark.GetSheetMetaReq{
		SpreadSheetToken: shetToken,
	}
	resp, _, err := s.Drive.GetSheetMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	sheets := []*Sheet{}
	if err := copier.Copy(&sheets, &resp.Sheets); err != nil {
		return nil, err
	}
	return sheets, nil
}

func NewSheetsClient(appID, appSecret, apiDomain, logLevel string) *Sheets {
	logger := &larkLogger{logger: logrus.StandardLogger()}
	return &Sheets{
		Lark: lark.New(lark.WithAppCredential(appID, appSecret),
			lark.WithStore(getStoreSlot(appID)),
			lark.WithLogger(logger, getLarkLogLevel(logLevel)),
			lark.WithOpenBaseURL(apiDomain)),
		appID: appID,
	}
}
