package goplus

import "hbuf/plugin/core/wire"

// FileContext 单个 .hbuf 文件的模板根上下文。
// 四类输出各自持有独立的 import 集合,避免互相污染。
type FileContext struct {
	Ctx *Ctx

	databaseImports Imports
	verifyImports   Imports
	exportImports   Imports
	mqImports       Imports

	Datas    []*DataContext
	Verifies []*VerifyContext
	Exports  []*ExportContext
	Mqs      []*MqContext
}

// NewFileContext 构建文件的模板上下文:遍历全部 data 类型,
// 为每类输出生成对应的结构化数据与 import 登记。
func NewFileContext(ctx *Ctx, file *wire.File) (*FileContext, error) {
	fc := &FileContext{
		Ctx:             ctx,
		databaseImports: Imports{},
		verifyImports:   Imports{},
		exportImports:   Imports{},
		mqImports:       Imports{},
	}
	for _, typ := range file.Types {
		if typ == nil || typ.Data == nil {
			continue
		}
		dc, err := ctx.buildDatabase(fc.databaseImports, typ.Data)
		if err != nil {
			return nil, err
		}
		fc.Datas = append(fc.Datas, dc)

		vc, err := ctx.buildVerify(fc.verifyImports, typ.Data)
		if err != nil {
			return nil, err
		}
		fc.Verifies = append(fc.Verifies, vc)

		if ec := ctx.buildExport(fc.exportImports, typ.Data); ec != nil {
			fc.Exports = append(fc.Exports, ec)
		}
		if mc := ctx.buildMq(fc.mqImports, typ.Data); mc != nil {
			fc.Mqs = append(fc.Mqs, mc)
		}
	}
	return fc, nil
}
