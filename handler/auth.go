package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/nedpals/supabase-go"
	"github.com/ryanzola/dreampicai/db"
	"github.com/ryanzola/dreampicai/pkg/kit/validate"
	"github.com/ryanzola/dreampicai/pkg/sb"
	"github.com/ryanzola/dreampicai/pkg/util"
	"github.com/ryanzola/dreampicai/types"
	"github.com/ryanzola/dreampicai/view/auth"
)

const (
	sessionUserKey        = "user"
	sessionAccessTokenKey = "accessToken"
)

func HandleResetPasswordIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.ResetPassword())
}

func HandleResetPasswordCreate(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)

	params := map[string]any{
		"email":      user.Email,
		"redirectTo": "http://localhost:3000/auth/reset-password",
	}
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", sb.BaseAuthURL, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", os.Getenv("SUPABASE_SECRET"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("supabase password recovery responded with a non 200 status code: %d => %s", resp.StatusCode, string(b))
	}

	return render(r, w, auth.ResetPasswordInitiated(user.Email))
}

func HandleResetPasswordUpdate(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	params := map[string]any{
		"password": r.FormValue("password"),
	}
	resp, err := sb.Client.Auth.UpdateUser(r.Context(), user.AccessToken, params)
	errors := auth.ResetPasswordErrors{
		NewPassword: "Please enter a valid password",
	}
	if err != nil {
		return render(r, w, auth.ResetPasswordForm(errors))
	}
	fmt.Printf("%+v\n", resp)

	return hxRedirect(w, r, "/")
}

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

func HandleSignupIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, auth.Signup())
}

func HandleLoginWithGoogle(w http.ResponseWriter, r *http.Request) error {
	resp, err := sb.Client.Auth.SignInWithProvider(supabase.ProviderSignInOptions{
		Provider:   "google",
		RedirectTo: "http://localhost:3000/auth/callback",
	})
	if err != nil {
		return err
	}

	http.Redirect(w, r, resp.URL, http.StatusSeeOther)
	return nil
}

func HandleLoginCreate(w http.ResponseWriter, r *http.Request) error {
	credentials := supabase.UserCredentials{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if !util.ValidateEmail(credentials.Email) {
		return render(r, w, auth.LoginForm(credentials, auth.LoginErrors{
			Email: "Please enter a valid email",
		}))
	}

	// call supabase
	resp, err := sb.Client.Auth.SignIn(r.Context(), credentials)
	if err != nil {
		slog.Error("login error", "err", err)
		return render(r, w, auth.LoginForm(credentials, auth.LoginErrors{
			InvalidCredentials: "The credentials you entered are invalid",
		}))
	}

	if err := setAuthSession(w, r, resp.AccessToken); err != nil {
		return err
	}

	return hxRedirect(w, r, "/")
}

func HandleSignupCreate(w http.ResponseWriter, r *http.Request) error {
	params := auth.SignupParams{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	errors := auth.SignupErrors{}
	if ok := validate.New(&params, validate.Fields{
		"Email":    validate.Rules(validate.Email),
		"Password": validate.Rules(validate.Password),
		"ConfirmPassword": validate.Rules(
			validate.Equal(params.Password),
			validate.Message("Password does not match"),
		),
	}).Validate(&errors); !ok {
		return render(r, w, auth.SignupForm(params, errors))
	}

	user, err := sb.Client.Auth.SignUp(r.Context(), supabase.UserCredentials{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		slog.Error("signup error", "err", err)
		return render(r, w, auth.SignupForm(params, auth.SignupErrors{
			Email: "Email is already taken",
		}))
	}

	return render(r, w, auth.SignupSuccess(user.Email))
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
