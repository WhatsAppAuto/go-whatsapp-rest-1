package libs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Rhymen/go-whatsapp"
)

type wpResponse struct {
	Id          string
	RemoteJid   string
	SenderJid   string
	FromMe      bool
	Timestamp   uint64
	PushName    string
	MessageType string
}
type wpTextResponse struct {
	Info    wpResponse
	Text    string
	Context whatsapp.ContextInfo
}
type wpDataResponse struct {
	Info    wpResponse
	Data    string
	Caption string
	Context whatsapp.ContextInfo
}

type wpLocationResponse struct {
	Info      wpResponse
	Latitude  float64
	Longitude float64
	Name      string
	Address   string
	Url       string
	Context   whatsapp.ContextInfo
}
type myHandler struct{}

func (myHandler) HandleError(err error) {
	fmt.Fprintf(os.Stderr, "%v", err)
}
func (myHandler) HandleLiveLocationMessage(message whatsapp.LiveLocationMessage) {


	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// _, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(data))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }


}

func (myHandler) HandleLocationMessage(message whatsapp.LocationMessage) {

	response := wpLocationResponse{
		Info: wpResponse{
			Id:          message.Info.Id,
			RemoteJid:   message.Info.RemoteJid,
			SenderJid:   message.Info.SenderJid,
			FromMe:      message.Info.FromMe,
			Timestamp:   message.Info.Timestamp,
			PushName:    message.Info.PushName,
			MessageType: "location",
		},
		Latitude:  message.DegreesLatitude,
		Longitude: message.DegreesLongitude,
		Name:      message.Name,
		Address:   message.Address,
		Url:       message.Url,
		Context:   message.ContextInfo,
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// _, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(data))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	

}

func (myHandler) HandleStickerMessage(message whatsapp.StickerMessage) {
	if message.Info.RemoteJid == "558599628852-1585619935@g.us" {
		data, err := json.Marshal(message)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
		// _, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(data))
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	}

}

func (myHandler) HandleTextMessage(message whatsapp.TextMessage) {
	response := wpTextResponse{
		Info: wpResponse{
			Id:          message.Info.Id,
			RemoteJid:   message.Info.RemoteJid,
			SenderJid:   message.Info.SenderJid,
			FromMe:      message.Info.FromMe,
			Timestamp:   message.Info.Timestamp,
			PushName:    message.Info.PushName,
			MessageType: "text",
		},
		Text:    message.Text,
		Context: message.ContextInfo,
	}
	data, err := json.Marshal(response)
	_, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return

	}

}

func (myHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	response := wpDataResponse{
		Info: wpResponse{
			Id:          message.Info.Id,
			RemoteJid:   message.Info.RemoteJid,
			SenderJid:   message.Info.SenderJid,
			FromMe:      message.Info.FromMe,
			Timestamp:   message.Info.Timestamp,
			PushName:    message.Info.PushName,
			MessageType: "data/image",
		},
		Caption: message.Caption,
		Data:    "",
		Context: message.ContextInfo,
	}
	data, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err1 := os.Create("./share/store/" + string(message.Info.Id))
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	response.Data = string(message.Info.Id)
	payload, err2 := json.Marshal(response)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	fmt.Println(string(data))

	_, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(err)
		return
	}

}

func (myHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	//fmt.Println(message.Info)
}

func (myHandler) HandleVideoMessage(message whatsapp.VideoMessage) {

	response := wpDataResponse{
		Info: wpResponse{
			Id:          message.Info.Id,
			RemoteJid:   message.Info.RemoteJid,
			SenderJid:   message.Info.SenderJid,
			FromMe:      message.Info.FromMe,
			Timestamp:   message.Info.Timestamp,
			PushName:    message.Info.PushName,
			MessageType: "data/image",
		},
		Caption: message.Caption,
		Data:    "",
		Context: message.ContextInfo,
	}
	data, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err1 := os.Create("./share/store/" + string(message.Info.Id))
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	response.Data = string(message.Info.Id)
	payload, err2 := json.Marshal(response)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	fmt.Println(string(data))

	_, err = http.Post("http://127.0.0.1:8000/botman", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(err)
		return
	}

}

func (myHandler) HandleAudioMessage(message whatsapp.AudioMessage) {
	if message.Info.FromMe == true {
		return
	}
	data, err := message.Download()
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err1 := os.Create("./share/store/" + string(message.Info.Id))
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("wrote %d bytes\n", n)
}

func (myHandler) HandleJsonMessage(message string) {
	fmt.Println(message)
}

func (myHandler) HandleContactMessage(message whatsapp.ContactMessage) {
	fmt.Println(message)
}
