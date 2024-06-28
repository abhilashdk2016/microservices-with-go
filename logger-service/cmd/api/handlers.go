package main

import (
	"log"
	"net/http"

	"github.com/abhilashdk2016/my-logger/data"
	"github.com/abhilashdk2016/toolkit/v2"
)

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	var tools toolkit.Tools
	err := tools.ReadJSON(w, r, &requestPayload)
	if err != nil {
		log.Println("Error while reading JSON")
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err = app.Models.LogEntry.Insert(event)
	if err != nil {
		log.Println("Error while inserting log entry")
		tools.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	resp := toolkit.JSONResponse{
		Error:   false,
		Message: "logged",
	}

	tools.WriteJSON(w, http.StatusAccepted, &resp)
}
