package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// router handles the plugin's HTTP endpoints (interactive message actions and dialogs).
	router *mux.Router
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.router = p.initRouter()

	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTrigger,
		DisplayName:      "Spoiler",
		Description:      "Post a message that stays hidden until someone chooses to reveal it.",
		AutoComplete:     true,
		AutoCompleteDesc: "Post a message that stays hidden until someone chooses to reveal it",
		AutoCompleteHint: "[message]",
		AutocompleteData: model.NewAutocompleteData(commandTrigger, "[message]", "Post a message that stays hidden until someone chooses to reveal it"),
	}); err != nil {
		return errors.Wrap(err, "failed to register /spoiler command")
	}

	return nil
}

// ExecuteCommand handles the /spoiler slash command.
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return p.executeSpoilerCommand(args), nil
}

// ServeHTTP handles HTTP requests sent to /plugins/<plugin-id>/...
func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}
