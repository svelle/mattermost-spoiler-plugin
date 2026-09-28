package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParseSpoilerText(t *testing.T) {
	for name, tc := range map[string]struct {
		command  string
		expected string
	}{
		"no text":             {"/spoiler", ""},
		"only whitespace":     {"/spoiler   ", ""},
		"single word":         {"/spoiler secret", "secret"},
		"keeps inner spacing": {"/spoiler  the  butler did it ", "the  butler did it"},
		"multi-line":          {"/spoiler\nline one\n\n**line two**", "line one\n\n**line two**"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseSpoilerText(tc.command))
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	p := &Plugin{}

	t.Run("posts a spoiler", func(t *testing.T) {
		resp, appErr := p.ExecuteCommand(nil, &model.CommandArgs{Command: "/spoiler The butler did it"})
		require.Nil(t, appErr)

		assert.Equal(t, model.CommandResponseTypeInChannel, resp.ResponseType)
		assert.Equal(t, spoilerPostType, resp.Type)
		assert.Equal(t, "The butler did it", resp.Props[propSpoilerText])
		assert.NotContains(t, resp.Text, "butler", "the visible message must not leak the spoiler")

		require.Len(t, resp.Attachments, 1)
		require.Len(t, resp.Attachments[0].Actions, 1)
		action := resp.Attachments[0].Actions[0]
		assert.Equal(t, "/plugins/"+manifest.Id+revealPath, action.Integration.URL)
		assert.NoError(t, resp.Attachments[0].IsValid())
	})

	t.Run("empty spoiler shows usage", func(t *testing.T) {
		resp, appErr := p.ExecuteCommand(nil, &model.CommandArgs{Command: "/spoiler "})
		require.Nil(t, appErr)
		assert.Equal(t, model.CommandResponseTypeEphemeral, resp.ResponseType)
		assert.Empty(t, resp.Type)
	})
}

func newRevealRequest(t *testing.T, path, userID string, body *model.PostActionIntegrationRequest) *http.Request {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	if userID != "" {
		r.Header.Set("Mattermost-User-Id", userID)
	}
	return r
}

func decodeActionResponse(t *testing.T, w *httptest.ResponseRecorder) *model.PostActionIntegrationResponse {
	t.Helper()
	require.Equal(t, http.StatusOK, w.Code)
	var resp model.PostActionIntegrationResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return &resp
}

func TestReveal(t *testing.T) {
	const (
		userID    = "user1"
		channelID = "channel1"
		postID    = "post1"
	)
	spoilerPost := &model.Post{Id: postID, ChannelId: channelID, Type: spoilerPostType}
	spoilerPost.AddProp(propSpoilerText, "The butler did it")

	setup := func(t *testing.T) (*Plugin, *plugintest.API) {
		api := &plugintest.API{}
		t.Cleanup(func() { api.AssertExpectations(t) })
		p := &Plugin{}
		p.SetAPI(api)
		p.router = p.initRouter()
		return p, api
	}

	t.Run("rejects unauthenticated requests", func(t *testing.T) {
		p, _ := setup(t)
		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, revealPath, "", &model.PostActionIntegrationRequest{PostId: postID}))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("opens a dialog with the spoiler", func(t *testing.T) {
		p, api := setup(t)
		api.On("GetPost", postID).Return(spoilerPost, nil)
		api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(true)
		api.On("OpenInteractiveDialog", mock.MatchedBy(func(req model.OpenDialogRequest) bool {
			return req.TriggerId == "trigger1" &&
				req.Dialog.IntroductionText == "The butler did it" &&
				req.URL == "/plugins/"+manifest.Id+dialogClosePath &&
				req.IsValid() == nil
		})).Return(nil)

		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, revealPath, userID, &model.PostActionIntegrationRequest{PostId: postID, TriggerId: "trigger1"}))
		assert.Empty(t, decodeActionResponse(t, w).EphemeralText)
	})

	t.Run("falls back to an ephemeral post when the dialog fails", func(t *testing.T) {
		p, api := setup(t)
		api.On("GetPost", postID).Return(spoilerPost, nil)
		api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(true)
		api.On("OpenInteractiveDialog", mock.Anything).Return(model.NewAppError("test", "test", nil, "", http.StatusBadRequest))
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, revealPath, userID, &model.PostActionIntegrationRequest{PostId: postID, TriggerId: "trigger1"}))
		assert.Contains(t, decodeActionResponse(t, w).EphemeralText, "The butler did it")
	})

	t.Run("serves buttons of posts created by version 1.x", func(t *testing.T) {
		p, api := setup(t)
		api.On("GetPost", postID).Return(spoilerPost, nil)
		api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(true)

		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, legacyRevealPath, userID, &model.PostActionIntegrationRequest{PostId: postID}))
		assert.Contains(t, decodeActionResponse(t, w).EphemeralText, "The butler did it")
	})

	t.Run("does not reveal spoilers from channels the user cannot read", func(t *testing.T) {
		p, api := setup(t)
		api.On("GetPost", postID).Return(spoilerPost, nil)
		api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(false)

		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, revealPath, userID, &model.PostActionIntegrationRequest{PostId: postID, TriggerId: "trigger1"}))
		assert.NotContains(t, decodeActionResponse(t, w).EphemeralText, "butler")
	})

	t.Run("ignores posts that are not spoilers", func(t *testing.T) {
		p, api := setup(t)
		otherPost := &model.Post{Id: "post2", ChannelId: channelID}
		otherPost.AddProp(propSpoilerText, "not a spoiler post")
		api.On("GetPost", "post2").Return(otherPost, nil)
		api.On("HasPermissionToChannel", userID, channelID, model.PermissionReadChannelContent).Return(true)

		w := httptest.NewRecorder()
		p.ServeHTTP(nil, w, newRevealRequest(t, revealPath, userID, &model.PostActionIntegrationRequest{PostId: "post2"}))
		assert.NotContains(t, decodeActionResponse(t, w).EphemeralText, "not a spoiler post")
	})
}
