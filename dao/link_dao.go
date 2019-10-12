package dao

import "peg.nu/short/model"

type LinkDAO interface {
	Create(link model.Link) bool
	Exists(short string) bool
	Delete(short string) bool
	Get(short string) (model.Link, error)

	Close()
}
