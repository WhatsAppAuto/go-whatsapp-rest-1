package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/exatasmente/go-rest/api/auth"
	"github.com/exatasmente/go-rest/api/models"
	responses "github.com/exatasmente/go-rest/api/responses"
	"golang.org/x/crypto/bcrypt"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	userToken := models.UserToken{}
	err = json.Unmarshal(body, &user)
	fmt.Println(user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	userModel := models.User{}
	userGotten, err := userModel.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, err)
		return
	}
	tokenGotten, err := userToken.FindTokenByUserID(server.DB, uint32(userGotten.ID))
	if err != nil {

		token, err := server.SignIn(user.Email, user.Password)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		userToken.Prepare()
		userToken.Token = token
		userToken.UserId = userGotten.ID
		userToken.SaveToken(server.DB)
		responses.JSON(w, http.StatusOK, token)
		return
	}

	responses.JSON(w, http.StatusOK, tokenGotten.Token)
	return

}

func (server *Server) SignIn(email, password string) (string, error) {

	user := models.User{}

	err := server.DB.Debug().Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return "", err
	}
	err = models.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", err
	}
	return auth.CreateToken(user.ID)
}
