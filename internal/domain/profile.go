package domain

import (
	"database/sql"
	"github.com/radyatamaa/dating-apps-api/pkg/database/paginator"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"gorm.io/gorm"
	"time"
)

// Entity
type Profile struct {
	ID       int    `gorm:"column:id;primarykey;autoIncrement:true"`
	User     User   `gorm:"foreignkey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	UserID   int    `gorm:"column:user_id"`
	Name     string `gorm:"type:varchar(255);column:name"`
	Photo    string `gorm:"type:text;column:photo"`
	Age      int    `gorm:"column:age"`
	Bio      string `gorm:"type:text;column:bio"`
	Longitude float64 `gorm:"column:longitude"`
	Latitude  float64 `gorm:"column:latitude"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName name of table
func (r Profile) TableName() string {
	return "profile"
}

type ProfileQueryWithUser struct {
	ID       int    `gorm:"column:id;primarykey;autoIncrement:true"`
	UserID   int    `gorm:"column:user_id"`
	Name     string `gorm:"type:varchar(255);column:name"`
	Photo    string `gorm:"type:text;column:photo"`
	Age      int    `gorm:"column:age"`
	Bio      string `gorm:"type:text;column:bio"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Distance float64  `gorm:"column:distance"`
	User
}

// TableName name of table
func (r ProfileQueryWithUser) TableName() string {
	return "profile"
}
//////////////////////////

// Requests
type UpdateLiveLocationProfilesRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
//////////////////////////

// Responses
type GetProfilesResponse struct {
	Id          int     `json:"id"`
	Name     string `json:"name"`
	Photo string `json:"photo"`
	Age      int `json:"age"`
	Bio      string `json:"bio"`
	Verified bool `json:"verified"`
	Distance string `json:"distance"`
}

type GetProfilesResponsePaginationResponse struct {
	Data      []GetProfilesResponse           `json:"data"`
	Paginator paginator.MetaPaginatorResponse `json:"paginator"`
}
//////////////////////////

// Mapping
func IsPremium(premiumExp sql.NullTime) bool {
	return premiumExp.Valid && premiumExp.Time.After(time.Now())
}

func FromProfileToGetProfilesResponse(data ProfileQueryWithUser) GetProfilesResponse {
	scala := "m"
	distanceCalculate := helper.MilsToMeters(data.Distance)
	if distanceCalculate >= 1000 {
		scala = "km"
		distanceCalculate = helper.MetersToKilometers(distanceCalculate)
	}
	distance := helper.FloatToString(distanceCalculate) + scala

	return GetProfilesResponse{
		Id:    data.ID,
		Name:  data.Name,
		Photo: data.Photo,
		Age:   data.Age,
		Bio:   data.Bio,
		Verified: IsPremium(data.PremiumExpiresAt),
		Distance: distance,
	}
}

func ToGetProfilesResponsePaginationResponsee(data []GetProfilesResponse, page, limit, offset, totalAllRecords int) *GetProfilesResponsePaginationResponse {
	return &GetProfilesResponsePaginationResponse{
		Data:      data,
		Paginator: paginator.MetaPaginatorResponse{}.MappingPaginator(page, limit, offset, totalAllRecords, len(data)),
	}
}
//////////////////////////

// Seeder
func SeederDataUserProfile(db *gorm.DB) {
	for i := 0; i < 10; i++ {
		dataUser := User{
			PasswordHash:     "password",
			Email:            helper.GenerateRandomEmail(),
		}
		db.Create(&dataUser)
		latitude,longitude := helper.GenerateRandomLatLong()
		db.Create(&Profile{
			UserID:    dataUser.ID,
			Name:      helper.GenerateRandomString(15),
			Photo:     "https://fastly.picsum.photos/id/660/536/354.jpg?hmac=rleJ6NCajocyX8aMHVw-b2M6nmTjnUV56Y2YKnxmkG4",
			Age:       21,
			Bio:       "dummy",
			Latitude: latitude,
			Longitude: longitude,
		})
	}
}