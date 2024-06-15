package config

import (
	"context"
	"net/http"

	"github.com/DanCreative/veracode-admin-plus/common"
	"github.com/gorilla/schema"
)

type SettingsHandler struct {
	configService *applicationConfigService
}

// message is used on the frontend
type message struct {
	IsSuccess  bool
	ShouldShow bool
	Text       string
}

// settingsOptions is used on the frontend
type settingsOptions struct {
	RequiresNewProfile bool
}

func (s *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	common.Page("/api/rest/settings").Render(r.Context(), w)
}

func (s *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	msg := message{
		ShouldShow: false,
	}
	config, _ := s.configService.GetConfig()

	var options settingsOptions

	profiles, err := s.configService.GetProfiles()
	if err != nil {
		options.RequiresNewProfile = true
	}

	ComponentSettingsContent(msg, options, config, profiles).Render(r.Context(), w)
}

func (s *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	settings := struct {
		Config  Config  `schema:"config"`
		Profile Profile `schema:"profile"`
	}{}

	r.ParseForm()
	var decoder = schema.NewDecoder()
	ctx := r.Context()

	err := decoder.Decode(&settings, r.Form)
	if err != nil {
		renderErrorMessage(err, ctx, w)
		return
	}

	if settings.Profile.Name != "" {
		// I.E. first time login. Create the credentials file.
		err = s.configService.CreateCredentialsFile(settings.Profile)
		if err != nil {
			renderErrorMessage(err, ctx, w)
			return
		}
		settings.Config.Profile = settings.Profile.Name
	}

	err = s.configService.UpdateConfig(settings.Config)
	if err != nil {
		renderErrorMessage(err, ctx, w)
		return
	}

	err = s.configService.SetClient()
	if err != nil {
		renderErrorMessage(err, ctx, w)
		return
	}

	updatedConfig, err := s.configService.GetConfig()
	if err != nil {
		renderErrorMessage(err, ctx, w)
		return
	}

	profiles, err := s.configService.GetProfiles()
	if err != nil {
		renderErrorMessage(err, ctx, w)
		return
	}

	ComponentSettingsContent(message{
		IsSuccess:  true,
		ShouldShow: true,
		Text:       "Settings successfully updated!",
	}, settingsOptions{}, updatedConfig, profiles).Render(r.Context(), w)
}

func renderErrorMessage(err error, ctx context.Context, w http.ResponseWriter) {
	responseHeaders := w.Header()
	responseHeaders.Add("HX-Retarget", "#message")
	responseHeaders.Add("HX-Reswap", "outerHTML")

	ComponentMessage(message{
		ShouldShow: true,
		IsSuccess:  false,
		Text:       err.Error(),
	}).Render(ctx, w)
}

func NewSettingsHandler(configService *applicationConfigService) SettingsHandler {
	return SettingsHandler{
		configService: configService,
	}
}
