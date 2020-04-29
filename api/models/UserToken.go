package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

type UserToken struct {
	ID        uint32 `gorm:"primary_key;auto_increment" json:"id"`
	Token     string `gorm:unique;not null`
	UserId    uint32
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (t *UserToken) Prepare() {
	t.ID = 0
	t.Token = ""
	t.UserId = 0
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
}

func (u *UserToken) Validate(action string) error {
	if u.Token == "" {
		return errors.New("Required Token")
	}
	return nil
}

func (u *UserToken) SaveToken(db *gorm.DB) (*UserToken, error) {

	var err error
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &UserToken{}, err
	}
	return u, nil
}

func (u *UserToken) FindAllTokens(db *gorm.DB) (*[]UserToken, error) {
	var err error
	tokens := []UserToken{}
	err = db.Debug().Model(&UserToken{}).Limit(100).Find(&tokens).Error
	if err != nil {
		return &[]UserToken{}, err
	}
	return &tokens, err
}
func (u *UserToken) FindTokenByUserID(db *gorm.DB, userId uint32) (*UserToken, error) {
	var err error
	err = db.Debug().Model(UserToken{}).Where("user_id = ?", userId).Take(&u).Error
	if err != nil {
		return &UserToken{}, err
	}
	return u, err
}
func (u *UserToken) GetToken(db *gorm.DB, token string) (*UserToken, error) {
	var err error
	err = db.Debug().Model(UserToken{}).Where("token = ?", token).Take(&u).Error
	if err != nil {
		return &UserToken{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &UserToken{}, errors.New("Token Not Found")
	}
	return u, err
}

func (u *UserToken) UpdateAToken(db *gorm.DB, tokenId string) (*UserToken, error) {
	db = db.Debug().Model(&UserToken{}).Where("token = ?", tokenId).Take(&UserToken{}).UpdateColumns(
		map[string]interface{}{
			"token":     u.Token,
			"update_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &UserToken{}, db.Error
	}

	err := db.Debug().Model(&UserToken{}).Where("token = ?", tokenId).Take(&u).Error
	if err != nil {
		return &UserToken{}, err
	}
	return u, nil
}

func (u *UserToken) DeleteAToken(db *gorm.DB, tokenId string) (int64, error) {

	db = db.Debug().Model(&UserToken{}).Where("token = ?", tokenId).Take(&UserToken{}).Delete(&UserToken{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
