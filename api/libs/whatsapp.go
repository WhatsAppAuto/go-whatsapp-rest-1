package libs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"

	"github.com/jinzhu/gorm"

	"strings"
	"sync"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
	waproto "github.com/Rhymen/go-whatsapp/binary/proto"
	models "github.com/exatasmente/go-whatsapp-rest/api/models"
	qrcode "github.com/skip2/go-qrcode"
)

var wac = make(map[string]*whatsapp.Conn)
var wacMutex = make(map[string]*sync.Mutex)

func WAParseJID(jid string) string {
	components := strings.Split(jid, "@")

	if len(components) > 1 {
		jid = components[0]
	}

	suffix := "@s.whatsapp.net"

	if len(strings.SplitN(jid, "-", 2)) == 2 {
		suffix = "@g.us"
	}

	return jid + suffix
}

func WAGetSendMutexSleep() time.Duration {
	rand.Seed(time.Now().UnixNano())

	waitMin := 1000
	waitMax := 3000

	return time.Duration(rand.Intn(waitMax-rand.Intn(waitMin)) + waitMin)
}

func WASendWithMutex(jid string, content interface{}) (string, error) {
	mutex, ok := wacMutex[jid]

	if !ok {
		mutex := &sync.Mutex{}
		wacMutex[jid] = mutex
	}

	mutex.Lock()
	time.Sleep(WAGetSendMutexSleep() * time.Millisecond)

	id, err := wac[jid].Send(content)
	mutex.Unlock()

	return id, err
}

func WASyncVersion(conn *whatsapp.Conn, versionClientMajor int, versionClientMinor int, versionClientBuild int) (string, error) {
	// Bug Happend When Using This Function
	// Then Set Manualy WhatsApp Client Version
	// versionServer, err := whatsapp.CheckCurrentServerVersion()
	// if err != nil {
	// 	return "", err
	// }

	conn.SetClientVersion(versionClientMajor, versionClientMinor, versionClientBuild)
	versionClient := conn.GetClientVersion()

	return fmt.Sprintf("whatsapp version %v.%v.%v", versionClient[0], versionClient[1], versionClient[2]), nil
}

func WATestPing(conn *whatsapp.Conn) error {
	ok, err := conn.AdminTest()
	if !ok {
		if err != nil {
			return err
		}

		return errors.New("something when wrong while trying to ping, please check phone connectivity")
	}

	return nil
}

func WAGenerateQR(timeout int, chanqr chan string, qrstr chan<- string) {
	select {
	case tmp := <-chanqr:
		png, _ := qrcode.Encode(tmp, qrcode.Medium, 256)
		qrstr <- base64.StdEncoding.EncodeToString(png)
	}
}

func WASessionInit(jid string, versionClientMajor int, versionClientMinor int, versionClientBuild int, timeout int) error {
	if wac[jid] == nil {
		conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
		conn.AddHandler(myHandler{})
		if err != nil {
			return err
		}
		conn.SetClientName("Wp GO REST", "GO Rest")

		info, err := WASyncVersion(conn, versionClientMajor, versionClientMinor, versionClientBuild)
		if err != nil {
			return err
		}
		fmt.Println("whatsapp", info)

		wacMutex[jid] = &sync.Mutex{}
		wac[jid] = conn
	}

	return nil
}

func WASessionLoad(jid string, DB *gorm.DB) (whatsapp.Session, error) {
	sessionModel := models.WpSession{}
	token := models.UserToken{}

	err := DB.Debug().Model(models.UserToken{}).Where("token = ?", jid).Take(&token).Error
	if err != nil {
		return whatsapp.Session{}, err
	}
	err = DB.Debug().Model(models.WpSession{}).Where("user_id = ?", token.UserId).Take(&sessionModel).Error
	if err != nil {
		return whatsapp.Session{}, err
	}
	session := whatsapp.Session{
		ClientId:    sessionModel.ClientId,
		ClientToken: sessionModel.ClientToken,
		ServerToken: sessionModel.ServerToken,
		EncKey:      sessionModel.EncKey,
		MacKey:      sessionModel.MacKey,
		Wid:         sessionModel.Wid,
	}
	return session, nil
}

func WASessionSave(jid string, session whatsapp.Session, DB *gorm.DB) error {
	sessionModel := models.WpSession{}
	sessionModel.Prepare()
	err := DB.Debug().Model(models.WpSession{}).Where("token = ?", jid).Take(&sessionModel).Error
	if err != nil {
		return err
	}
	sessionModel.ServerToken = session.ServerToken
	sessionModel.ClientToken = session.ClientToken
	sessionModel.ClientId = session.ClientId
	sessionModel.EncKey = session.EncKey
	sessionModel.MacKey = session.MacKey
	sessionModel.Wid = session.Wid
	_, err = sessionModel.UpdateAWpSession(DB)
	if err != nil {
		return err
	}

	return nil
}

func WASessionExist(jid string, DB *gorm.DB) bool {
	token := models.UserToken{}
	err := DB.Debug().Model(models.UserToken{}).Where("token = ?", jid).Take(&token).Error
	if err != nil {
		return false
	}

	return true
}

func WASessionConnect(jid string, versionClientMajor int, versionClientMinor int, versionClientBuild int, timeout int, qrstr chan<- string, errmsg chan<- error, DB *gorm.DB) {
	chanqr := make(chan string)

	session, err := WASessionLoad(jid, DB)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()

		err = WASessionLogin(jid, versionClientMajor, versionClientMinor, versionClientBuild, timeout, chanqr, DB)
		if err != nil {
			errmsg <- err
			return
		}
		return
	}

	err = WASessionRestore(jid, versionClientMajor, versionClientMinor, versionClientBuild, timeout, session, DB)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()

		err = WASessionLogin(jid, versionClientMajor, versionClientMinor, versionClientBuild, timeout, chanqr, DB)
		if err != nil {
			errmsg <- err
			return
		}
	}

	err = WATestPing(wac[jid])
	if err != nil {
		errmsg <- err
		return
	}

	errmsg <- errors.New("")
	return
}

func WASessionLogin(jid string, versionClientMajor int, versionClientMinor int, versionClientBuild int, timeout int, qrstr chan<- string, DB *gorm.DB) error {
	if wac[jid] != nil {
		delete(wac, jid)
	}

	err := WASessionInit(jid, versionClientMajor, versionClientMinor, versionClientBuild, timeout)
	if err != nil {
		return err
	}

	session, err := wac[jid].Login(qrstr)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return errors.New("connection is invalid")
		default:
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(jid, session, DB)
	if err != nil {
		return err
	}

	return nil
}

func WASessionRestore(jid string, versionClientMajor int, versionClientMinor int, versionClientBuild int, timeout int, sess whatsapp.Session, DB *gorm.DB) error {
	if wac[jid] != nil {
		delete(wac, jid)
	}

	err := WASessionInit(jid, versionClientMajor, versionClientMinor, versionClientBuild, timeout)
	if err != nil {
		return err
	}

	session, err := wac[jid].RestoreWithSession(sess)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return errors.New("connection is invalid")
		default:
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(jid, session, DB)
	if err != nil {
		return err
	}

	return nil
}

func WASessionLogout(jid string) error {
	if wac[jid] != nil {
		err := wac[jid].Logout()
		if err != nil {
			return err
		}
		delete(wac, jid)
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WASessionValidate(jid string) error {
	if wac[jid] == nil {
		return errors.New("connection is invalid")
	}

	return nil
}

func WAMessageText(jid string, jidDest string, msgText string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		Text: msgText,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}

func WAMessageDocument(jid string, jidDest string, msgDocumentStream multipart.File, msgDocumentType string, msgDocumentInfo string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.DocumentMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		Content:  msgDocumentStream,
		Type:     msgDocumentType,
		FileName: msgDocumentInfo,
		Title:    msgDocumentInfo,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}

func WAMessageAudio(jid string, jidDest string, msgAudioStream multipart.File, msgAudioType string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.AudioMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		Content: msgAudioStream,
		Type:    msgAudioType,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}

func WAMessageImage(jid string, jidDest string, msgImageStream multipart.File, msgImageType string, msgImageInfo string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		Content: msgImageStream,
		Type:    msgImageType,
		Caption: msgImageInfo,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}

func WAMessageVideo(jid string, jidDest string, msgVideoStream multipart.File, msgVideoType string, msgVideoInfo string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.VideoMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		Content: msgVideoStream,
		Type:    msgVideoType,
		Caption: msgVideoInfo,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}

func WAMessageLocation(jid string, jidDest string, msgLatitude float64, msgLongitude float64, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string

	err := WASessionValidate(jid)
	if err != nil {
		return "", errors.New(err.Error())
	}

	rJid := WAParseJID(jidDest)

	content := whatsapp.LocationMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: rJid,
		},
		DegreesLatitude:  msgLatitude,
		DegreesLongitude: msgLongitude,
	}

	if len(msgQuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &msgQuoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: msgQuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     rJid,
		}

		content.ContextInfo = ctxQuotedInfo
	}

	id, err = WASendWithMutex(jid, content)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "sending message timed out":
			return id, nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			delete(wac, jid)
			return "", errors.New("connection is invalid")
		default:
			return "", err
		}
	}

	return id, nil
}
