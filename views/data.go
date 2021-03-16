package views

import (
	"go_playground/models"
	"go_playground/utils"
	"log"
)

const (
	AlertLvlDanger  = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
)

type Alert struct {
	Level,
	Message string
}

type Data struct {
	Alert *Alert
	Yield interface{}
	User  *models.User
}

func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(utils.PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlDanger,
			Message: pErr.Public(),
		}
	} else {
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertLvlDanger,
			Message: AlertMsgGeneric,
		}
	}
}
