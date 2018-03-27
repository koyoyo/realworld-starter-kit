package models

import (
	"time"
)

type Follower struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	FollowerID  uint `gorm:"unique_index:follow"`
	Follower    User
	FollowingID uint `gorm:"unique_index:follow"`
	Following   User
}

type Profile struct {
	ID        uint    `json:"-"`
	Username  string  `json:"username"`
	Bio       string  `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

type ProfileResponse struct {
	Profile Profile `json:"profile"`
}

func (db *DB) GetUserProfile(username string) *ProfileResponse {
	user := User{}
	db.Where(&User{Username: username}).First(&user)
	return &ProfileResponse{
		Profile: Profile{
			ID:       user.ID,
			Username: user.Username,
			Bio:      user.Bio,
			Image:    user.Image,
		},
	}
}

func (db *DB) IsFollowing(followerID, followingID uint) bool {
	var count uint
	db.Model(&Follower{}).Where(&Follower{FollowerID: followerID, FollowingID: followingID}).Count(&count)
	return count > 0
}

func (db *DB) Follow(followerID, followingID uint) {
	follower := Follower{}
	db.FirstOrCreate(&follower, Follower{FollowerID: followerID, FollowingID: followingID})
}

func (db *DB) Unfollow(followerID, followingID uint) {
	db.Where(&Follower{FollowerID: followerID, FollowingID: followingID}).Delete(Follower{})
}
