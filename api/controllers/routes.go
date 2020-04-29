package controllers

import "github.com/exatasmente/go-rest/api/middlewares"

func (s *Server) initializeRoutes() {
	// Home Route
	s.Router.HandleFunc("/", middlewares.SetMiddlewareJSON(s.Home)).Methods("GET")
	s.Router.HandleFunc("/file/{id}", middlewares.SetMiddlewareAuthentication(s.DB, s.FileDownload)).Methods("POST")
	// Login Route
	s.Router.HandleFunc("/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST")

	//Users routes
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST")
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareAuthentication(s.DB, s.GetUsers)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(s.GetUser)).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.DB, s.UpdateUser))).Methods("PUT")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareAuthentication(s.DB, s.DeleteUser)).Methods("DELETE")

	// Whatsapp Routes
	s.Router.HandleFunc("/send/text", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendText)).Methods("POST")
	s.Router.HandleFunc("/send/image", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendImage)).Methods("POST")
	s.Router.HandleFunc("/send/video", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendVideo)).Methods("POST")
	s.Router.HandleFunc("/send/audio", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendAudio)).Methods("POST")
	s.Router.HandleFunc("/send/location", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendLocation)).Methods("POST")
	s.Router.HandleFunc("/send/document", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppSendDocument)).Methods("POST")
	s.Router.HandleFunc("/wp/login", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppLogin)).Methods("POST")
	s.Router.HandleFunc("/wp/logout", middlewares.SetMiddlewareAuthentication(s.DB, s.WhatsAppLogout)).Methods("POST")
}
