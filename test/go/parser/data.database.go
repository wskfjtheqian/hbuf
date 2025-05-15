package parser

import (
	"context"
	"database/sql"
	"github.com/wskfjtheqian/hbuf_golang/pkg/db"
)

func (val *GetInfoReq) DbScan() (string, []any) {
	return `user_id, name, age`,
		[]any{&val.UserId, db.NewJson(&val.Name), &val.Age}
}

func (val *GetInfoReq) DbName() string {
	return `get_info_req`
}

func (g InfoReq) DbGet(ctx context.Context) (*GetInfoReq, error) {
	tableName := db.GET(ctx).Table("get_info_req")
	s := db.NewSql()
	s.T("SELECT user_id, name, age FROM ").T(tableName).T(" WHERE del_time = 0")
	s.T(" LIMIT 1")
	var val *GetInfoReq
	_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {
		val = &GetInfoReq{}
		return false, rows.Scan(&val.UserId, db.NewJson(&val.Name), &val.Age)
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (g InfoSet) DbSet(ctx context.Context) (int64, int64, error) {
	tableName := db.GET(ctx).Table("get_info_req")
	s := db.NewSql()
	s.T("UPDATE ").T(tableName).T(" SET ").Del(",")
	s.T(",").T("name = if(user_id > ").V(*g.UserId).T(",\"asdsa\", ").V(&g.Name).T(")")
	s.T(",").T("age = ").V(&g.Age)
	s.T("WHERE 1 = 1 ")
	if nil != g.UserId {
		s.T("AND user_id = ").V(db.NewJson(&g.UserId))
	}

	return s.Exec(ctx)
}
