package domain

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string `gorm:"index"`
	Nickname    string
	Icon        string
	Credentials []webauthn.Credential `gorm:"-"`
}

func (User) TableName() string {
	return "user"
}

// UserRepository represent the User's repository contract
type UserRepo interface {
	CreateUser(users []*User) error
	DeleteUser(userID int64) error
	UpdateUser(userID int64, nickName *string, icon *string) error
	QueryUser(userID *int64, userName *string, nickName *string, limit, offset int) ([]*User, error)
	MGetUsers(userIDs []int64) ([]*User, error)
}

// UserUsecase represent the User's usecases
type UserUsecase interface {
	CreateUser(users []*User) error
	DeleteUser(userID int64) error
	UpdateUser(userID int64, nickName *string, icon *string) error
	QueryUser(userID *int64, userName *string, nickName *string, limit, offset int) ([]*User, error)
	MGetUsers(userIDs []int64) ([]*User, error)
	CheckUserExist(userID *int64, userName *string) (bool, error)
	FindByID(userID *int64) (*User, error)
}
