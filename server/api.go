package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
)

const (
	revealPath = "/api/v1/reveal"

	// legacyRevealPath is the button URL used by spoilers posted with version 1.x of the
	// plugin. Those posts still carry it, so keep serving it.
	legacyRevealPath = "/show"
)

func (p *Plugin) initRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(p.requireUser)

	router.HandleFunc(revealPath, p.handleReveal).Methods(http.MethodPost)
	router.HandleFunc(legacyRevealPath, p.handleReveal).Methods(http.MethodPost)

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
// the hidden content only to that person, as an ephemeral post right below the spoiler.
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

	p.writeActionResponse(w, &model.PostActionIntegrationResponse{
		EphemeralText:    revealMessage(text),
		SkipSlackParsing: true,
	})
}

// revealMessage formats the spoiler as a quote under a small heading, so it reads as part of
// the spoiler post above it. Quoting every line keeps lists, code blocks and paragraphs
// inside the quote.
func revealMessage(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return "**🙈 Spoiler**\n" + strings.Join(lines, "\n")
}

func (p *Plugin) writeActionResponse(w http.ResponseWriter, response *model.PostActionIntegrationResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.API.LogError("Failed to write action response", "error", err.Error())
	}
}
