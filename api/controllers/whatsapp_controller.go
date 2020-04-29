package controllers

import (
	// "encoding/json"
	// "errors"
	// "fmt"
	// "io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	//"github.com/gorilla/mux"
	"github.com/exatasmente/go-rest/api/auth"
	// "github.com/exatasmente/go-rest/api/models"
	// "github.com/exatasmente/go-rest/api/utils/formaterror"
	config "github.com/exatasmente/go-rest/api/helpers"
	libs "github.com/exatasmente/go-rest/api/libs"
	responses "github.com/exatasmente/go-rest/api/responses"
)

type reqWhatsAppLogin struct {
	Output   string
	Timeout  int
	WhatsApp struct {
		Client struct {
			Version struct {
				Major int
				Minor int
				Build int
			}
		}
	}
}

type resWhatsAppLogin struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		QRCode  string `json:"qrcode"`
		Timeout int    `json:"timeout"`
	} `json:"data"`
}

type reqWhatsAppSendMessage struct {
	MSISDN        string
	Message       string
	QuotedID      string
	QuotedMessage string
	Delay         int
}

type reqWhatsAppSendLocation struct {
	MSISDN        string
	Latitude      float64
	Longitude     float64
	QuotedID      string
	QuotedMessage string
	Delay         int
}

type resWhatsAppSendMessage struct {
	MessageID string `json:"msgid"`
}

func (server *Server) WhatsAppLogin(w http.ResponseWriter, r *http.Request) {
	jid := auth.ExtractToken(r)
	r.ParseForm()
	var err error
	var reqBody reqWhatsAppLogin
	reqBody.Output = r.FormValue("output")
	reqTimeout := r.FormValue("timeout")

	reqVersionClientMajor := r.FormValue("client_version_major")
	reqVersionClientMinor := r.FormValue("client_version_minor")
	reqVersionClientBuild := r.FormValue("client_version_build")

	if len(reqBody.Output) == 0 {
		reqBody.Output = "json"
	}

	if len(reqTimeout) == 0 {
		reqBody.Timeout = 5
	} else {
		reqBody.Timeout, err = strconv.Atoi(reqTimeout)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	if len(reqVersionClientMajor) == 0 {
		reqBody.WhatsApp.Client.Version.Major = config.Config.GetInt("WHATSAPP_CLIENT_VERSION_MAJOR")
	} else {
		reqBody.WhatsApp.Client.Version.Major, err = strconv.Atoi(reqVersionClientMajor)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	if len(reqVersionClientMinor) == 0 {
		reqBody.WhatsApp.Client.Version.Minor = config.Config.GetInt("WHATSAPP_CLIENT_VERSION_MINOR")
	} else {
		reqBody.WhatsApp.Client.Version.Minor, err = strconv.Atoi(reqVersionClientMinor)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	if len(reqVersionClientBuild) == 0 {
		reqBody.WhatsApp.Client.Version.Build = config.Config.GetInt("WHATSAPP_CLIENT_VERSION_BUILD")
	} else {
		reqBody.WhatsApp.Client.Version.Build, err = strconv.Atoi(reqVersionClientBuild)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}
	qrstr := make(chan string)
	errmsg := make(chan error)

	go func() {
		libs.WASessionConnect(jid, reqBody.WhatsApp.Client.Version.Major, reqBody.WhatsApp.Client.Version.Minor, reqBody.WhatsApp.Client.Version.Build, reqBody.Timeout, qrstr, errmsg, server.DB)
	}()

	select {
	case qrcode := <-qrstr:
		qrcode = "data:image/png;base64," + qrcode

		switch strings.ToLower(reqBody.Output) {
		case "json":
			var response resWhatsAppLogin

			response.Status = true
			response.Code = 200
			response.Message = "Success"
			response.Data.QRCode = qrcode
			response.Data.Timeout = reqBody.Timeout

			responses.ResponseWrite(w, response.Code, response)
		case "html":
			var response string

			response = `
        <html>
          <head>
            <title>WhatsApp Login</title>
          </head>
          <body>
            <img src="` + qrcode + `" />
            <p>
              <b>QR Code Scan</b>
              <br/>
              Timeout in ` + strconv.Itoa(reqBody.Timeout) + ` Second(s)
            </p>
          </body>
        </html>
      `

			w.Write([]byte(response))
		default:
			responses.ResponseBadRequest(w, "")
		}
	case err := <-errmsg:
		if len(err.Error()) != 0 {
			responses.ResponseInternalError(w, err.Error())
			return
		}

		responses.ResponseSuccess(w, "")
	}
}

func (server *Server) WhatsAppLogout(w http.ResponseWriter, r *http.Request) {
	var err error
	jid := auth.ExtractToken(r)
	err = libs.WASessionLogout(jid)
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	responses.ResponseSuccess(w, "")
}

func (server *Server) WhatsAppSendText(w http.ResponseWriter, r *http.Request) {
	var err error
	jid := auth.ExtractToken(r)

	r.ParseForm()

	var reqBody reqWhatsAppSendMessage
	reqBody.MSISDN = r.FormValue("msisdn")
	reqBody.Message = r.FormValue("message")
	reqBody.QuotedID = r.FormValue("quotedid")
	reqBody.QuotedMessage = r.FormValue("quotedmsg")
	reqDelay := r.FormValue("delay")

	if len(reqDelay) == 0 {
		reqBody.Delay = 0
	} else {
		reqBody.Delay, err = strconv.Atoi(reqDelay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	if len(reqBody.MSISDN) == 0 || len(reqBody.Message) == 0 {
		responses.ResponseBadRequest(w, "")
		return
	}

	id, err := libs.WAMessageText(jid, reqBody.MSISDN, reqBody.Message, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	var resBody resWhatsAppSendMessage
	resBody.MessageID = id

	responses.ResponseSuccessWithData(w, "", resBody)
}

func WhatsAppSendContent(w http.ResponseWriter, r *http.Request, c string) {
	var err error
	jid := auth.ExtractToken(r)

	err = r.ParseMultipartForm(config.Config.GetInt64("SERVER_UPLOAD_LIMIT"))
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	var reqBody reqWhatsAppSendMessage
	reqBody.MSISDN = r.FormValue("msisdn")
	reqBody.QuotedID = r.FormValue("quotedid")
	reqBody.QuotedMessage = r.FormValue("quotedmsg")
	reqDelay := r.FormValue("delay")

	if len(reqDelay) == 0 {
		reqBody.Delay = 0
	} else {
		reqBody.Delay, err = strconv.Atoi(reqDelay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	var mpFileStream multipart.File
	var mpFileHeader *multipart.FileHeader

	switch c {
	case "document":
		mpFileStream, mpFileHeader, err = r.FormFile("document")
		reqBody.Message = mpFileHeader.Filename

	case "audio":
		mpFileStream, mpFileHeader, err = r.FormFile("audio")

	case "image":
		mpFileStream, mpFileHeader, err = r.FormFile("image")
		reqBody.Message = r.FormValue("message")

	case "video":
		mpFileStream, mpFileHeader, err = r.FormFile("video")
		reqBody.Message = r.FormValue("message")
	}

	if err != nil {
		responses.ResponseBadRequest(w, err.Error())
		return
	}
	defer mpFileStream.Close()

	mpFileType := mpFileHeader.Header.Get("Content-Type")

	if len(reqBody.MSISDN) == 0 {
		responses.ResponseBadRequest(w, "")
		return
	}

	var id string

	switch c {
	case "document":
		id, err = libs.WAMessageDocument(jid, reqBody.MSISDN, mpFileStream, mpFileType, reqBody.Message, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}

	case "audio":
		id, err = libs.WAMessageAudio(jid, reqBody.MSISDN, mpFileStream, mpFileType, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}

	case "image":
		id, err = libs.WAMessageImage(jid, reqBody.MSISDN, mpFileStream, mpFileType, reqBody.Message, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}

	case "video":
		id, err = libs.WAMessageVideo(jid, reqBody.MSISDN, mpFileStream, mpFileType, reqBody.Message, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	var resBody resWhatsAppSendMessage
	resBody.MessageID = id

	responses.ResponseSuccessWithData(w, "", resBody)
}

func (server *Server) WhatsAppSendDocument(w http.ResponseWriter, r *http.Request) {
	WhatsAppSendContent(w, r, "document")
}

func (server *Server) WhatsAppSendAudio(w http.ResponseWriter, r *http.Request) {
	WhatsAppSendContent(w, r, "audio")
}

func (server *Server) WhatsAppSendImage(w http.ResponseWriter, r *http.Request) {
	WhatsAppSendContent(w, r, "image")
}

func (server *Server) WhatsAppSendVideo(w http.ResponseWriter, r *http.Request) {
	WhatsAppSendContent(w, r, "video")
}

func (server *Server) WhatsAppSendLocation(w http.ResponseWriter, r *http.Request) {
	var err error
	jid := auth.ExtractToken(r)

	r.ParseForm()

	var reqBody reqWhatsAppSendLocation
	reqBody.MSISDN = r.FormValue("msisdn")
	reqBody.QuotedID = r.FormValue("quotedid")
	reqBody.QuotedMessage = r.FormValue("quotedmsg")
	reqDelay := r.FormValue("delay")

	reqBody.Latitude, err = strconv.ParseFloat(r.FormValue("latitude"), 64)
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	reqBody.Longitude, err = strconv.ParseFloat(r.FormValue("longitude"), 64)
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	if len(reqDelay) == 0 {
		reqBody.Delay = 0
	} else {
		reqBody.Delay, err = strconv.Atoi(reqDelay)
		if err != nil {
			responses.ResponseInternalError(w, err.Error())
			return
		}
	}

	if len(reqBody.MSISDN) == 0 {
		responses.ResponseBadRequest(w, "")
		return
	}

	id, err := libs.WAMessageLocation(jid, reqBody.MSISDN, reqBody.Latitude, reqBody.Longitude, reqBody.QuotedID, reqBody.QuotedMessage, reqBody.Delay)
	if err != nil {
		responses.ResponseInternalError(w, err.Error())
		return
	}

	var resBody resWhatsAppSendMessage
	resBody.MessageID = id

	responses.ResponseSuccessWithData(w, "", resBody)
}

// func (server *Server) CreateWpSession(w http.ResponseWriter, r *http.Request) {

// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 	}
// 	session := models.WpSession{}
// 	err = json.Unmarshal(body, &session)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	session.Prepare()
// 	err = session.Validate()
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	sessionCreated, err := session.SaveWpSession(server.DB)

// 	if err != nil {

// 		formattedError := formaterror.FormatError(err.Error())

// 		responses.ERROR(w, http.StatusInternalServerError, formattedError)
// 		return
// 	}
// 	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.RequestURI, userCreated.ID))
// 	responses.JSON(w, http.StatusCreated, sessionCreated)
// }

// func (server *Server) GetWpSession(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	sid, err := strconv.ParseUint(vars["id"], 10, 32)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	session := models.WpSession{}
// 	sessionGotten, err := session.FindWpSessionByID(server.DB, uint64(sid))
// 	if err != nil {
// 		responses.ERROR(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	responses.JSON(w, http.StatusOK, sessionGotten)
// }

// func (server *Server) UpdateSession(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	sid, err := strconv.ParseUint(vars["id"], 10, 32)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	session := models.WpSession{}
// 	sessionGotten :=session.FindWpSessionByID(server.DB, uint64(sid))
// 	err = json.Unmarshal(body, &session)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	tokenID, err := auth.ExtractTokenID(r)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
// 		return
// 	}
// 	if tokenID != uint32(uid) {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
// 		return
// 	}
// 	session.Prepare()
// 	err = session.Validate()
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	updatedSession, err := session.UpdateAWpSession(server.DB)
// 	if err != nil {
// 		formattedError := formaterror.FormatError(err.Error())
// 		responses.ERROR(w, http.StatusInternalServerError, formattedError)
// 		return
// 	}
// 	responses.JSON(w, http.StatusOK, updatedSession)
// }

// func (server *Server) DeleteS(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)

// 	user := models.User{}

// 	uid, err := strconv.ParseUint(vars["id"], 10, 32)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	tokenID, err := auth.ExtractTokenID(r)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
// 		return
// 	}
// 	if tokenID != 0 && tokenID != uint32(uid) {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
// 		return
// 	}
// 	_, err = user.DeleteAUser(server.DB, uint32(uid))
// 	if err != nil {
// 		responses.ERROR(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
// 	responses.JSON(w, http.StatusNoContent, "")
// }
