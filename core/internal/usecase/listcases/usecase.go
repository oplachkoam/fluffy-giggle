package listcases

import (
	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
)

type Usecase struct {
	cases port.CaseCatalog
}

func New(cases port.CaseCatalog) *Usecase {
	return &Usecase{cases: cases}
}

func (u *Usecase) Execute() []*entity.Case {
	return u.cases.List()
}
