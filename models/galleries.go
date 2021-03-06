package models

import (
	"github.com/jinzhu/gorm"
	"go_playground/utils"
)

type Gallery struct {
	gorm.Model
	UserID uint    `gorm:"not_null;index"`
	Title  string  `gorm:"not_null"`
	Images []Image `gorm:"-"`
}

func (g *Gallery) ImagesSplitN(n int) [][]Image {
	ret := make([][]Image, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}
	for i, img := range g.Images {
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}
	return ret
}

type GalleryDB interface {
	ByUserID(userID uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	ByID(id uint) (*Gallery, error)
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type GalleryService interface {
	GalleryDB
}

var _ GalleryService = &galleryService{}

type galleryService struct {
	GalleryDB
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{&galleryGorm{db}},
	}
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
}

type galleryValFunc func(*Gallery) error

func runGalleryValFuncs(gallery *Gallery, fns ...galleryValFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery

	db := gg.db.Where("id = ?", id)
	return &gallery, first(db, &gallery)
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	gg.db.Where("user_id = ?", userID).Find(&galleries)

	return galleries, nil
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	valFuncs := []galleryValFunc{
		gv.titleRequired,
		gv.userIDRequired,
	}
	if err := runGalleryValFuncs(gallery, valFuncs...); err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	valFuncs := []galleryValFunc{
		gv.titleRequired,
		gv.userIDRequired,
	}
	if err := runGalleryValFuncs(gallery, valFuncs...); err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

func (gv *galleryValidator) userIDRequired(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return utils.ValErr.UserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(gallery *Gallery) error {
	if gallery.Title == "" {
		return utils.ValErr.TitleRequired
	}
	return nil
}

func (gv *galleryValidator) idGreaterThanZero(gallery *Gallery) error {
	if gallery.ID <= 0 {
		return utils.ValErr.InvalidID
	}
	return nil
}

func (gv *galleryValidator) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	if err := runGalleryValFuncs(&gallery, gv.idGreaterThanZero); err != nil {
		return err
	}
	return gv.GalleryDB.Delete(id)
}
