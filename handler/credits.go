package handler

import (
	"fmt"
	"net/http"
	"os"

	"dreampicai/db"
	"dreampicai/view/credits"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func HandleCreditsIndex(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, credits.Index())
}

func HandleStripeCheckoutCreate(w http.ResponseWriter, r *http.Request) error {
	stripe.Key = os.Getenv("STRIPE_API_KEY")
	baseURL := os.Getenv("HOST_URL")

	successURL := fmt.Sprintf("%s/checkout/success/{CHECKOUT_SESSION_ID}", baseURL)
	cancelURL := fmt.Sprintf("%s/checkout/cancel", baseURL)

	checkoutParams := &stripe.CheckoutSessionParams{
		// SuccessURL: stripe.String("http://localhost:7331/checkout/success/{CHECKOUT_SESSION_ID}"),
		SuccessURL: stripe.String(successURL),
		// CancelURL:  stripe.String("http://localhost:7331/checkout/cancel"),
		CancelURL: stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(chi.URLParam(r, "priceID")),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}

	s, err := session.New(checkoutParams)
	if err != nil {
		return err
	}

	return hxRedirect(w, r, s.URL)
}

func HandleStripeCheckoutSuccess(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	sessionID := chi.URLParam(r, "sessionID")
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return err
	}

	lineItemParams := stripe.CheckoutSessionListLineItemsParams{}
	lineItemParams.Session = stripe.String(sess.ID)
	iter := session.ListLineItems(&lineItemParams)
	iter.Next()
	item := iter.LineItem()
	priceID := item.Price.ID

	switch priceID {
	case os.Getenv("CREDITS_100_PRICE_ID"):
		user.Account.Credits += 100
	case os.Getenv("CREDITS_250_PRICE_ID"):
		user.Account.Credits += 250
	case os.Getenv("CREDITS_600_PRICE_ID"):
		user.Account.Credits += 600
	default:
		return fmt.Errorf("invalid price ID: %s", priceID)
	}

	if err := db.UpdateAccount(&user.Account); err != nil {
		return err
	}

	http.Redirect(w, r, "/generate", http.StatusSeeOther)
	return nil
}

func HandleStripeCheckoutCancel(w http.ResponseWriter, r *http.Request) error {
	return nil
}
