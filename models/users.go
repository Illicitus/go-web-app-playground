package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go_playground/utils"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
)

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

type UserDB interface {
	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Methods for querying for single user
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
}

// UserService is a set of methods used to manipulate and work with the user model
type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(db *gorm.DB, pepper, hmacKey string) UserService {
	ug := &userGorm{db}
	hmac := utils.NewHMAC(hmacKey)
	uv := newUserValidator(ug, hmac, pepper)

	return &userService{
		UserDB: uv,
		pepper: pepper,
	}
}

var _ UserService = &userService{}

type userService struct {
	UserDB
	pepper string
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, utils.GormErr.InvalidPassword
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User

	db := ug.db.Where("id = ?", id)
	return &user, first(db, &user)
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User

	db := ug.db.Where("email = ?", email)
	return &user, first(db, &user)
}

func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User

	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac       utils.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

func newUserValidator(udb UserDB, hmac utils.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: emailRegex,
		pepper:     pepper,
	}
}

func (uv *userValidator) defaultRemember(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := utils.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token

	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := utils.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return utils.ValErr.RememberTooShort
	}

	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return utils.ValErr.RememberRequired
	}

	return nil
}

func (uv *userValidator) idGreaterThanZero(user *User) error {
	if user.ID <= 0 {
		return utils.ValErr.InvalidID
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)

	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return utils.ValErr.EmailRequired
	}

	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return utils.ValErr.InvalidEmail
	}

	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)

	// Email doesn't exists, no error to user
	if err == utils.GormErr.NotFound {
		return nil
	}

	if err != nil {
		return err
	}

	// Email exists, return validation error then
	if user.ID != existing.ID {
		return utils.ValErr.EmailTaken
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return utils.ValErr.PasswordTooShort
	}

	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return utils.ValErr.PasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return utils.ValErr.PasswordRequired
	}

	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)

	return nil
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwBytes := []byte(user.Password + uv.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pwBytes), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

func (uv *userValidator) Create(user *User) error {
	valFuncs := []userValFunc{
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.defaultRemember,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
	}
	if err := runUserValFuncs(user, valFuncs...); err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	valFuncs := []userValFunc{
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
	}
	if err := runUserValFuncs(user, valFuncs...); err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	if err := runUserValFuncs(&user, uv.idGreaterThanZero); err != nil {
		return err
	}

	return uv.UserDB.Delete(id)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{Remember: token}
	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{Email: email}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}
