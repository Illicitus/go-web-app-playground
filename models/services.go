package models

import "github.com/jinzhu/gorm"

type Services struct {
	UserService    UserService
	GalleryService GalleryService
	ImageService   ImageService
	db             *gorm.DB
}

func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return &Services{
		db:             db,
		UserService:    NewUserService(db),
		GalleryService: NewGalleryService(db),
		ImageService:   NewImageService(),
	}, nil
}

func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}).Error
}

// Close user service db connection
func (s *Services) Close() error {
	return s.db.Close()
}
