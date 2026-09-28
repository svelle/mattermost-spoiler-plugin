package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
)

const (
	revealPath      = "/api/v1/reveal"
	dialogClosePath = "/api/v1/dialog/close"

	// legacyRevealPath is the button URL used by spoilers posted with version 1.x of the
	// plugin. Those posts still carry it, so keep serving it.
	legacyRevealPath = "/show"

	revealDialogTitle = "🙈 Spoiler"
)

func (p *Plugin) initRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(p.requireUser)

	router.HandleFunc(revealPath, p.handleReveal).Methods(http.MethodPost)
	router.HandleFunc(legacyRevealPath, p.handleReveal).Methods(http.MethodPost)
	router.HandleFunc(dialogClosePath, p.handleDialogClose).Methods(http.MethodPost)

	return router
}

// requireUser rejects requests that the Mattermost server has not authenticated.
func (p *Plugin) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-Id") == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleReveal is called when someone presses the reveal button on a spoiler post. It shows
// the hidden content only to that person: in a dialog when possible, otherwise as an
// ephemeral post.
func (p *Plugin) handleReveal(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")

	var request model.PostActionIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		p.writeActionResponse(w, &model.PostActionIntegrationResponse{EphemeralText: "Could not read the request."})
		return
	}

	post, appErr := p.API.GetPost(request.PostId)
	if appErr != nil {
		p.writeActionResponse(w, &model.PostActionIntegrationResponse{EphemeralText: "This spoiler could not be found."})
		return
	}

	// The endpoint can be called directly with any post ID, so make sure the user is
	// allowed to read the channel the spoiler was posted in.
	if !p.API.HasPermissionToChannel(userID, post.ChannelId, model.PermissionReadChannelContent) {
		p.writeActionResponse(w, &model.PostActionIntegrationResponse{EphemeralText: "This spoiler could not be found."})
		return
	}

	text, ok := spoilerText(post)
	if !ok {
		p.writeActionResponse(w, &model.PostActionIntegrationResponse{EphemeralText: "This spoiler is empty."})
		return
	}

	if request.TriggerId != "" {
		appErr = p.API.OpenInteractiveDialog(model.OpenDialogRequest{
			TriggerId: request.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s%s", manifest.Id, dialogClosePath),
			Dialog: model.Dialog{
				CallbackId:       post.Id,
				Title:            revealDialogTitle,
				IntroductionText: text,
				SubmitLabel:      "Done",
			},
		})
		if appErr == nil {
			p.writeActionResponse(w, &model.PostActionIntegrationResponse{})
			return
		}
		p.API.LogWarn("Failed to open spoiler dialog, falling back to an ephemeral post", "post_id", post.Id, "error", appErr.Error())
	}

	p.writeActionResponse(w, &model.PostActionIntegrationResponse{
		EphemeralText:    "**🙈 Spoiler** _(only visible to you)_\n\n" + text,
		SkipSlackParsing: true,
	})
}

// handleDialogClose acknowledges the "Done" button of the reveal dialog.
func (p *Plugin) handleDialogClose(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&model.SubmitDialogResponse{}); err != nil {
		p.API.LogError("Failed to write dialog response", "error", err.Error())
	}
}

func (p *Plugin) writeActionResponse(w http.ResponseWriter, response *model.PostActionIntegrationResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to write action response", "error", err.Error())
	}
}
