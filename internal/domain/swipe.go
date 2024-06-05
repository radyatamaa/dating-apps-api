package domain

import (
	"time"
)

// Entity
type Swipe struct {
	ID        int       `gorm:"column:id;primarykey;autoIncrement:true"`
	User     User   `gorm:"foreignkey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	UserID   int    `gorm:"column:user_id;uniqueIndex:idx_user_profile"`
	Profile     Profile   `gorm:"foreignkey:ProfileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	ProfileID int       `gorm:"column:profile_id;uniqueIndex:idx_user_profile"`
	SwipeType string    `gorm:"type:varchar(255);column:swipe_type"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

// TableName name of table
func (r Swipe) TableName() string {
	return "swipes"
}
//////////////////////////

// Requests
type SwipeProfileRequest struct {
	ProfileID int `json:"profile_id" validate:"required,check_fk=ProfileID:profile:id"`
	SwipeType string `json:"swipe_type" validate:"required,enum=LIKE-PASS"`
}
//////////////////////////


// Mapping
func (s SwipeProfileRequest) ToSwipe(userId int) Swipe  {
	return Swipe{
		UserID:    userId,
		ProfileID: s.ProfileID,
		SwipeType: s.SwipeType,
	}
}
