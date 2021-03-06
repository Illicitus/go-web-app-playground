package models

import (
	"github.com/jinzhu/gorm"
	"go_playground/utils"
)

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return utils.GormErr.NotFound
	}
	return err

}
