package inertia

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type FlashErrors map[string]string

const errFlashCookie = "inertia_errflash"

// flashError sets a temp cookie for the next request's errors
func flashError(w http.ResponseWriter, r *http.Request, newErrors FlashErrors) {
	c, err := r.Cookie(errFlashCookie)
	fe := make(FlashErrors)
	if err != nil {
		c = &http.Cookie{
			Name:     errFlashCookie,
			Value:    "",
			Expires:  time.Now().Add(time.Minute),
			Secure:   true,
			HttpOnly: true,
		}
	} else {
		// we try if it fails, it fails
		_ = json.Unmarshal([]byte(c.Value), &fe)
	}

	for k, v := range newErrors {
		fe[k] = v
	}

	newV, err := json.Marshal(fe)

	if err != nil {
		slog.Error("could not marshal flashed errors", slog.String("err", err.Error()))
	}

	c.Value = base64.StdEncoding.EncodeToString(newV)

	http.SetCookie(w, c)
}

// getFlashErrs (and delete)
func getFlashErrs(w http.ResponseWriter, r *http.Request) (fe FlashErrors) {
	c, err := r.Cookie(errFlashCookie)
	if err != nil {
		return
	}

	val, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return
	}

	err = json.Unmarshal(val, &fe)

	if err != nil {
		return
	}

	// reset cookie
	c.Expires = time.Unix(0, 0)

	http.SetCookie(w, c)

	return
}
