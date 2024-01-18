package parser

import (
	"context"
	"database/sql"
	"github.com/wskfjtheqian/hbuf_golang/pkg/db"
)

func (val *GetInfoReq) DbScan() (string, []any) {
	return `user_id`,
		[]any{&val.UserId}
}

func (g InfoReq) DbGet(ctx context.Context) (*GetInfoReq, error) {
	s := db.NewSql()
	s.T("SELECT user_id FROM get_info_req WHERE del_time IS NULL")
	s.T(" LIMIT 1")
	var val GetInfoReq
	count, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {
		return false, rows.Scan(&val.UserId)
	})
	if err != nil {
		return nil, err
	}
	if 0 == count {
		return nil, nil
	}
	return &val, nil
}
