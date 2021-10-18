package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"service-rss/internal/config"
)

const (
	userInfoUrlRaw  = "https://www.googleapis.com/oauth2/v2/userinfo"
	scope           = "https://www.googleapis.com/auth/userinfo.email"
	tokenCookieName = "rsstoken"
)

type userInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

type googleAuthHandler struct {
	oauthConf *oauth2.Config
}

func NewGoogleAuthHandler(cfg *config.Config) Handler {
	return &googleAuthHandler{
		oauthConf: &oauth2.Config{
			ClientID:     cfg.GoogleAuthClientID,
			ClientSecret: cfg.GoogleAuthClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.GoogleAuthRedirectURL,
			Scopes:       []string{scope},
		},
	}
}

func (h *googleAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginUrl, err := url.Parse(h.oauthConf.Endpoint.AuthURL)
	if err != nil {
		log.WithError(err).Error("failed to parse endpoint")
	}

	parameters := url.Values{}
	parameters.Add("client_id", h.oauthConf.ClientID)
	parameters.Add("scope", strings.Join(h.oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", h.oauthConf.RedirectURL)
	parameters.Add("response_type", "code")

	loginUrl.RawQuery = parameters.Encode()
	redirectUrl := loginUrl.String()

	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

func (h *googleAuthHandler) GetEmail(w http.ResponseWriter, r *http.Request) (string, error) {
	tokenCookie, err := r.Cookie(tokenCookieName)

	var accessToken string
	if err == nil && tokenCookie != nil && len(tokenCookie.Value) > 0 {
		accessToken = tokenCookie.Value
	}

	if len(accessToken) == 0 {
		code := r.FormValue("code")
		if len(code) == 0 {
			return "", nil
		}

		token, err := h.oauthConf.Exchange(r.Context(), code)
		if err != nil {
			log.WithError(err).Error("failed to get token")
			return "", err
		}

		accessToken = token.AccessToken

		cookie := http.Cookie{
			Name:    tokenCookieName,
			Value:   accessToken,
			Expires: token.Expiry,
		}
		http.SetCookie(w, &cookie)
	}

	userInfoUrl, err := url.Parse(userInfoUrlRaw)
	if err != nil {
		log.WithError(err).Error("failed to parse user info url")
	}

	parameters := url.Values{}
	parameters.Add("access_token", accessToken)

	userInfoUrl.RawQuery = parameters.Encode()
	userInfoUrlString := userInfoUrl.String()

	resp, err := http.Get(userInfoUrlString)
	if err != nil {
		log.WithError(err).Error("failed to get user info")
		return "", err
	}
	defer resp.Body.Close()

	ui := &userInfo{}
	err = json.NewDecoder(resp.Body).Decode(ui)
	if err != nil {
		log.WithError(err).Error("failed to decode user info")
		return "", err
	}

	return ui.Email, nil
}
