package handler

import (
	"log/slog"
	"net/http"
	"os"

	"dreampicai/db"
	"dreampicai/pkg/kit/validate"
	"dreampicai/pkg/sb"
	"dreampicai/pkg/util"
	"dreampicai/types"
	"dreampicai/view/auth"

	"github.com/gorilla/sessions"
	"github.com/nedpals/supabase-go"
)

const (
	sessionUserKey        = "user"
	sessionAccessTokenKey = "accessToken"
)

func HandleAccountSetupIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.AccountSetup())
}

func HandleAccountSetupCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.AccountSetupParams{
		Username: r.FormValue("username"),
	}

	var errors auth.AccountSetupErrors
	ok := validate.New(&params, validate.Fields{
		"Username": validate.Rules(validate.Min(2), validate.Max(50)),
	}).Validate(&errors)
	if !ok {
		return render(r, w, auth.AccountSetupForm(params, errors))
	}

	user := getAuthenticatedUser(r)
	account := types.Account{
		UserID:   user.ID,
		Username: params.Username,
	}
	if err := db.CreateAccount(&account); err != nil {
		return err
	}

	return hxRedirect(w, r, "/")
}

func HandleLoginIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.Login())
}

func HandleLoginWithGoogle(w http.ResponseWriter, r *http.Request) error {
	resp, err := sb.Client.Auth.SignInWithProvider(supabase.ProviderSignInOptions{
		Provider: "google",
		// RedirectTo: "http://localhost:7331/auth/callback",
		RedirectTo: "https://dreampicai-zkd2aag6vq-ue.a.run.app/auth/callback",
	})
	if err != nil {
		return err
	}

	http.Redirect(w, r, resp.URL, http.StatusSeeOther)
	return nil
}

func HandleLoginCreate(w http.ResponseWriter, r *http.Request) error {
	credentials := supabase.UserCredentials{
		Email: r.FormValue("email"),
	}

	if !util.ValidateEmail(credentials.Email) {
		return render(r, w, auth.LoginForm(credentials, auth.LoginErrors{
			Email: "Please enter a valid email",
		}))
	}

	// call supabase
	err := sb.Client.Auth.SendMagicLink(r.Context(), credentials.Email)
	if err != nil {
		slog.Error("login error", "err", err)
		return render(r, w, auth.LoginForm(credentials, auth.LoginErrors{
			InvalidCredentials: err.Error(),
		}))
	}

	// if err := setAuthSession(w, r, resp.AccessToken); err != nil {
	// 	return err
	// }

	return render(r, w, auth.MagicLinkSuccess(credentials.Email))
}

func HandleAuthCallback(w http.ResponseWriter, r *http.Request) error {
	accessToken := r.URL.Query().Get("access_token")
	if len(accessToken) == 0 {
		return render(r, w, auth.CallbackScript())
	}

	if err := setAuthSession(w, r, accessToken); err != nil {
		return err
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func HandleLogoutCreate(w http.ResponseWriter, r *http.Request) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = ""
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)

	return nil
}

func setAuthSession(w http.ResponseWriter, r *http.Request, accessToken string) error {
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	session, _ := store.Get(r, sessionUserKey)
	session.Values[sessionAccessTokenKey] = accessToken
	return session.Save(r, w)
}
