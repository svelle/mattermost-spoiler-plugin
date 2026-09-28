package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	commandTrigger = "spoiler"

	// spoilerPostType is rendered by the webapp plugin component. Clients without the
	// webapp plugin (e.g. the mobile apps) fall back to the post message and attachment.
	spoilerPostType = model.PostCustomTypePrefix + "spoiler"

	// propSpoilerText holds the hidden content of a spoiler post.
	propSpoilerText = "spoiler_text"

	// spoilerMessage is the visible text of a spoiler post. It is what clients without the
	// webapp plugin, push notifications, search results and thread previews show, so it
	// must never contain the hidden content.
	spoilerMessage = "🙈 Spoiler — hidden until revealed"

	revealButtonLabel = "Reveal spoiler"
	spoilerColor      = "#8B5CF6"

	// maxSpoilerRunes matches the maximum length of a regular post message.
	maxSpoilerRunes = model.PostMessageMaxRunesV2
)

// parseSpoilerText extracts the text following the /spoiler trigger, preserving any
// formatting (including newlines) the user typed.
func parseSpoilerText(command string) string {
	command = strings.TrimLeftFunc(command, unicode.IsSpace)
	if i := strings.IndexFunc(command, unicode.IsSpace); i >= 0 {
		return strings.TrimSpace(command[i:])
	}
	return ""
}

func (p *Plugin) executeSpoilerCommand(args *model.CommandArgs) *model.CommandResponse {
	text := parseSpoilerText(args.Command)
	if text == "" {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Type the message you want to hide after the command, for example: `/spoiler The butler did it`",
		}
	}

	if utf8.RuneCountInString(text) > maxSpoilerRunes {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Spoilers can be at most %d characters long.", maxSpoilerRunes),
		}
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeInChannel,
		Type:         spoilerPostType,
		Text:         spoilerMessage,
		Props: model.StringInterface{
			propSpoilerText: text,
		},
		Attachments: []*model.MessageAttachment{spoilerAttachment()},
	}
}

// spoilerAttachment builds the card shown on clients without the webapp plugin. Its
// button shows the hidden content right below the post, visible only to the person who
// tapped it.
func spoilerAttachment() *model.MessageAttachment {
	return &model.MessageAttachment{
		Fallback: spoilerMessage,
		Color:    spoilerColor,
		Actions: []*model.PostAction{{
			Type:  model.PostActionTypeButton,
			Name:  revealButtonLabel,
			Style: "primary",
			Integration: &model.PostActionIntegration{
				// Relative plugin URLs are routed to the plugin inside the server, so the
				// button keeps working regardless of the Site URL, subpaths or proxies.
				URL: fmt.Sprintf("/plugins/%s%s", manifest.Id, revealPath),
			},
		}},
	}
}

// spoilerText returns the hidden content of a spoiler post.
func spoilerText(post *model.Post) (string, bool) {
	if post == nil || post.Type != spoilerPostType {
		return "", false
	}
	text, ok := post.GetProp(propSpoilerText).(string)
	return text, ok && text != ""
}
