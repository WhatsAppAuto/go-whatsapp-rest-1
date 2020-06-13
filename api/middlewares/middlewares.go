package middlewares

import (
	"errors"
	"net/http"

	"github.com/exatasmente/go-whatsapp-rest/api/auth"
	"github.com/exatasmente/go-whatsapp-rest/api/models"
	responses "github.com/exatasmente/go-whatsapp-rest/api/responses"
	"github.com/jinzhu/gorm"
)

func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func SetMiddlewareAuthentication(database *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := auth.TokenValid(r)

		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}
		token := models.UserToken{}
		err = database.Debug().Model(models.UserToken{}).Where("token = ?", auth.ExtractToken(r)).Take(&token).Error
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}
		next(w, r)
	}
}
