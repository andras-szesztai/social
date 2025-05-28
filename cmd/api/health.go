package main

import (
	"net/http"
)

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Check the health of the server
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"status": "ok", "environment": app.config.env, "version": version}
	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}

}
