package models

import (
	"errors"

	"time"

	"github.com/jinzhu/gorm"
)

type WpSession struct {
	ID          uint64 `gorm:"primary_key;auto_increment" json:"id"`
	User        User   `json:"user"`
	UserID      uint32 `gorm:"not null" json:"user_id"`
	ClientId    string
	ClientToken string
	ServerToken string
	EncKey      []byte
	MacKey      []byte
	Wid         string
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ws *WpSession) Prepare() {
	ws.ID = 0
	ws.User = User{}
	ws.ClientId = ""
	ws.ClientToken = ""
	ws.ServerToken = ""
	ws.EncKey = nil
	ws.MacKey = nil
	ws.Wid = ""
	ws.CreatedAt = time.Now()
	ws.UpdatedAt = time.Now()
}

func (ws *WpSession) Validate() error {
	return nil
}

func (ws *WpSession) SaveWpSession(db *gorm.DB) (*WpSession, error) {
	var err error
	err = db.Debug().Model(&WpSession{}).Create(&ws).Error
	if err != nil {
		return &WpSession{}, err
	}
	if ws.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", ws.UserID).Take(&ws.User).Error
		if err != nil {
			return &WpSession{}, err
		}
	}
	return ws, nil
}

func (ws *WpSession) FindAllWpSessions(db *gorm.DB) (*[]WpSession, error) {
	var err error
	sessions := []WpSession{}
	err = db.Debug().Model(&WpSession{}).Limit(100).Find(&sessions).Error
	if err != nil {
		return &[]WpSession{}, err
	}
	if len(sessions) > 0 {
		for i, _ := range sessions {
			err := db.Debug().Model(&User{}).Where("id = ?", sessions[i].UserID).Take(&sessions[i].User).Error
			if err != nil {
				return &[]WpSession{}, err
			}
		}
	}
	return &sessions, nil
}

func (ws *WpSession) FindWpSessionByID(db *gorm.DB, sid uint64) (*WpSession, error) {
	var err error
	err = db.Debug().Model(&WpSession{}).Where("id = ?", sid).Take(&ws).Error
	if err != nil {
		return &WpSession{}, err
	}
	if ws.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", ws.UserID).Take(&ws.User).Error
		if err != nil {
			return &WpSession{}, err
		}
	}
	return ws, nil
}

func (ws *WpSession) FindWpSessionByUser(db *gorm.DB, user *User) (*WpSession, error) {
	var err error
	err = db.Debug().Model(&WpSession{}).Where("UserID = ?", user.ID).Take(&ws).Error
	if err != nil {
		return &WpSession{}, err
	}

	return ws, nil
}

func (ws *WpSession) UpdateAWpSession(db *gorm.DB) (*WpSession, error) {

	var err error
	err = db.Debug().Model(&WpSession{}).Where("id = ?", ws.ID).Updates(
		WpSession{
			ClientId:    ws.ClientId,
			ClientToken: ws.ClientToken,
			ServerToken: ws.ServerToken,
			EncKey:      ws.EncKey,
			MacKey:      ws.MacKey,
			Wid:         ws.Wid,
			UpdatedAt:   time.Now(),
		}).Error
	if err != nil {
		return &WpSession{}, err
	}
	if ws.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", ws.UserID).Take(&ws.User).Error
		if err != nil {
			return &WpSession{}, err
		}
	}
	return ws, nil
}

func (ws *WpSession) DeleteAWpSession(db *gorm.DB, sid uint64, uid uint32) (int64, error) {

	db = db.Debug().Model(&WpSession{}).Where("id = ? and user_id = ?", sid, uid).Take(&WpSession{}).Delete(&WpSession{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Post not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
