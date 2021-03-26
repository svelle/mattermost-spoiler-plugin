package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) OnActivate() (err error) {
	err = p.API.RegisterCommand(&model.Command{
		Trigger:          "spoiler",
		DisplayName:      "Spoiler Text",
		Description:      "hides text from others until they click to reveal it",
		AutoComplete:     true,
		AutoCompleteDesc: "type what you want to hide from others",
	})
	return
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("Mattermost-User-Id")
	w.WriteHeader(200)
	if userId == "" {
		return
	}

	var action_data model.PostActionIntegrationRequest
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)
	defer r.Body.Close()
	if err := decoder.Decode(&action_data); err != nil {
		encoder.Encode(&model.PostActionIntegrationResponse{
			EphemeralText: err.Error(),
		})
		return
	}

	url := fmt.Sprintf("/plugins/%s/show", manifest.Id)
	if r.RequestURI != url {
		encoder.Encode(&model.PostActionIntegrationResponse{
			EphemeralText: r.RequestURI,
		})
		return
	}
	if post, err := p.API.GetPost(action_data.PostId); err != nil {
		encoder.Encode(&model.PostActionIntegrationResponse{
			EphemeralText: err.Error(),
		})
		return
	} else {
		if text, ok := post.Props["spoiler_text"]; ok {
			encoder.Encode(&model.PostActionIntegrationResponse{
				EphemeralText: text.(string),
			})
		}
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	output := fmt.Sprintf("This is a hidden message. Click Show to reveal")
	spoilered_text := strings.TrimPrefix(args.Command, "/spoiler ")
	resp := &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         output,
		Type:         "custom_spoiler",
		Props: model.StringInterface{
			"spoiler_text": spoilered_text,
			"attachments": []*model.SlackAttachment{{
				Actions: []*model.PostAction{{
					Integration: &model.PostActionIntegration{
						Context: model.StringInterface{},
						URL:     fmt.Sprintf("%s/plugins/%s/show", args.SiteURL, manifest.Id),
					},
					Type: model.POST_ACTION_TYPE_BUTTON,
					Name: "Show",
				},
				}}},
		},
	}
	return resp, nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
