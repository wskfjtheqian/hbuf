package build

import (
	"errors"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
)

type Error struct {
	Pos token.Pos
	Msg string
}

func NewError(pos token.Pos, message string) *Error {
	return &Error{Msg: message, Pos: pos}
}

func (b Error) Error() string {
	return b.Msg
}

func ErrorToFileError(err error, fSet *token.FileSet) error {
	var val *Error
	if errors.As(err, &val) {
		return &scanner.Error{
			Pos: fSet.Position(val.Pos),
			Msg: val.Msg,
		}
	}
	return err
}
