package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/abhilashdk2016/toolkit/v2"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var tools toolkit.Tools
	err := tools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		log.Println("Error while reading JSON")
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println(requestPayload.Email)
	log.Println(requestPayload.Password)
	// validate the user
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		log.Println("Error while fetching user by email")
		log.Println(err.Error())
		logErr := app.logAuth("unable to find user event from auth service", fmt.Sprintf("%s -> Error while fetching user by email", err.Error()))
		if logErr != nil {
			tools.ErrorJSON(w, logErr, http.StatusBadRequest)
		}
		tools.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	log.Println("valid", valid)
	if err != nil || !valid {
		log.Println("Error while validating user password")
		log.Println(err.Error())
		logErr := app.logAuth("password not matched event from auth service", fmt.Sprintf("%s -> Error while validating user password", err.Error()))
		if err != nil {
			tools.ErrorJSON(w, logErr, http.StatusBadRequest)
		}
		tools.ErrorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// log authentication
	err = app.logAuth("successfull login event from auth service", fmt.Sprintf("%s -> successfully logged in", user.Email))
	if err != nil {
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	payload := toolkit.JSONResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	tools.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logAuth(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
