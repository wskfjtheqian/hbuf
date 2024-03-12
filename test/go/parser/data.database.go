package parser

import (
	"context"
	"database/sql"
	"github.com/wskfjtheqian/hbuf_golang/pkg/db"
)

func (val *GetInfoReq) DbScan() (string, []any) {
	return `user_id, name, age`,
		[]any{&val.UserId, &val.Name, &val.Age}
}

func (g InfoReq) DbGet(ctx context.Context) (*GetInfoReq, error) {
	s := db.NewSql()
	s.T("SELECT user_id, name, age FROM get_info_req WHERE del_time IS NULL")
	s.T(" LIMIT 1")
	var val *GetInfoReq
	_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {
		val = &GetInfoReq{}
		return false, rows.Scan(&val.UserId, &val.Name, &val.Age)
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (g InfoSet) DbSet(ctx context.Context) (int64, int64, error) {
	s := db.NewSql()
	s.T("UPDATE get_info_req SET ").Del(",")
	s.T(",").T("name = if(user_id > ").T(*g.UserId).T(",\"asdsa\", ").V(&g.Name).T(")")
	s.T(",").T("age = ").V(&g.Age)
	s.T("WHERE 1 = 1 ")
	if nil != g.UserId {
		s.T("AND user_id = ").V(db.NewJson(&g.UserId))
	}

	return s.Exec(ctx)
}
