package dao

import (
	"fmt"
	"peg.nu/short/model"
)

type MemoryLinkDao struct {
	mapping map[string]string
}

func NewMemoryLinkDao() LinkDAO {
	return &MemoryLinkDao{mapping: map[string]string{}}
}

func (m *MemoryLinkDao) Create(link model.Link) bool {
	existed := m.Exists(link.Short)
	m.mapping[link.Short] = link.Long

	return existed
}

func (m *MemoryLinkDao) Exists(short string) bool {
	_, ok := m.mapping[short]
	return ok
}

func (m *MemoryLinkDao) Delete(short string) bool {
	delete(m.mapping, short)

	return true
}

func (m *MemoryLinkDao) Get(short string) (model.Link, error) {
	long, ok := m.mapping[short]

	if !ok {
		return model.Link{}, fmt.Errorf("unknown link")
	}

	return model.Link{
		Short: short,
		Long:  long,
	}, nil
}

func (m *MemoryLinkDao) Close() {}
