package dao

import (
	"peg.nu/short/model"
)

type MemoryUnsplashDao struct {
	image *model.Image
}

func NewMemoryUnsplashDao() UnsplashDAO {
	return &MemoryUnsplashDao{image: nil}
}

func (m *MemoryUnsplashDao) Get() (error, *model.Image) {
	if m.image == nil {
		return nil, &model.Image{}
	}

	return nil, &*m.image
}

func (m *MemoryUnsplashDao) Update(img model.Image) (error, bool) {
	m.image = &img

	return nil, true
}
