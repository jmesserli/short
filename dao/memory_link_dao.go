package dao

import (
	"fmt"
	"peg.nu/short/model"
)

type MemoryLinkDao struct {
	mapping map[string]model.Link
}

func NewMemoryLinkDao() LinkDAO {
	return &MemoryLinkDao{mapping: map[string]model.Link{}}
}

func (m *MemoryLinkDao) Create(link model.Link) bool {
	existed := m.Exists(link.Short)
	m.mapping[link.Short] = link

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
	link, ok := m.mapping[short]

	if !ok {
		return model.Link{}, fmt.Errorf("unknown link")
	}

	return link, nil
}

func (m *MemoryLinkDao) GetUserLinks(user string) ([]model.Link, error) {
	var links []model.Link

	for _, link := range m.mapping {
		if link.UserId != user {
			continue
		}

		links = append(links, link)
	}

	return links, nil
}

func (m *MemoryLinkDao) Close() {}
