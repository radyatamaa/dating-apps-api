package domain

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// Entity
type User struct {
	ID              int       `gorm:"column:id;primarykey;autoIncrement:true"`
	PasswordHash    string    `gorm:"type:varchar(255);column:password_hash"`
	Email           string    `gorm:"type:varchar(255);column:email"`
	PremiumExpiresAt sql.NullTime `gorm:"column:premium_expires_at"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

// TableName name of table
func (r User) TableName() string {
	return "users"
}
func (r *User) BeforeCreate(tx *gorm.DB) (err error) {
	if r.PasswordHash != ""{
		if bytes, err := bcrypt.GenerateFromPassword([]byte(r.PasswordHash), 10); err != nil {
			return err
		} else {
			r.PasswordHash = string(bytes)
		}
	}
	return nil
}

type UserQueryWithProfile struct {
	ID              int       `gorm:"column:id;primarykey;autoIncrement:true"`
	PasswordHash    string    `gorm:"type:varchar(255);column:password_hash"`
	Email           string    `gorm:"type:varchar(255);column:email"`
	PremiumExpiresAt sql.NullTime `gorm:"column:premium_expires_at"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
	ProfileId 		int `gorm:"column:profile_id"`
}

// TableName name of table
func (r UserQueryWithProfile) TableName() string {
	return "users"
}

//////////////////////////

// Requests
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password"  validate:"required"`
}

type RegisterRequest struct {
	Name string `form:"name" validate:"required,max=50"`
	Age int `form:"age" validate:"required"`
	Bio string `form:"bio" validate:"required,max=100"`
	Photo string `form:"-"`
	Email    string `form:"email" validate:"required,email_address,unique_store=email:users,max=100"`
	Password string `form:"password"  validate:"required,max=20"`
}
//////////////////////////

// Responses
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiredAt string    `json:"expired_at"`
	User      UserLogin `json:"user"`
}

type UserLogin struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
}
//////////////////////////

// Mapping
func FromUserToUserLogin(data *UserQueryWithProfile) UserLogin {
	return UserLogin{
		Id:    data.ID,
		Email: data.Email,
	}
}

func(r RegisterRequest) ToUser() User {
	return User{
		Email:            r.Email,
		PasswordHash: 	  r.Password,
	}
}

func(r RegisterRequest) ToProfile(userId int) Profile {
	return Profile{
		UserID:    userId,
		Name:      r.Name,
		Photo:     r.Photo,
		Age:       r.Age,
		Bio:       r.Bio,
	}
}