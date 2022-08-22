package dao

import (
	"peg.nu/short/model"
)

type UnsplashDAO interface {
	Get() (error, *model.Image)
	Update(img model.Image) (error, bool)
}
