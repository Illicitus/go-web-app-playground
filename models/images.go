package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Image is NOT stored in the database
type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	tmp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return tmp.String()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryId(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

var _ ImageService = &imageService{}

type imageService struct{}

func NewImageService() ImageService {
	return &imageService{}
}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	defer r.Close()

	path, err := is.mkImagePath(galleryID)
	if err != nil {
		return err
	}

	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := fmt.Sprintf("images/galleries/%v/", galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

func (is *imageService) ByGalleryId(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	imgStrings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]Image, len(imgStrings))
	for i := range imgStrings {
		imgStrings[i] = strings.Replace(imgStrings[i], path, "", 1)
		ret[i] = Image{
			Filename:  imgStrings[i],
			GalleryID: galleryID,
		}
	}
	return ret, nil
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
