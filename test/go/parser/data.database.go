package parser

import (
	"context"
	"database/sql"
	"github.com/wskfjtheqian/hbuf_golang/pkg/db"
)

func (val *Test) DbScan() (string, []any) {
	return `id`,
		[]any{&val.Id}
}

func (val *Test) DbName() string {
	return `test`
}

func (g Test) DbGet(ctx context.Context) (*Test, error) {
	tableName := db.GET(ctx).Table("test")
	s := db.NewSql()
	s.T("SELECT id FROM ").T(tableName).T(" WHERE delete_time IS  NULL")
	s.T("AND id = ").V(&g.Id)
	s.T(" LIMIT 1")
	var val *Test
	_, err := s.Query(ctx, func(rows *sql.Rows) (bool, error) {
		val = &Test{}
		return false, rows.Scan(&val.Id)
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}
