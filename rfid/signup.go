package rfid

import (
	"encoding/json"
	"io"
	"net/http"
)

var signupSessionsMap = map[string]any{} // TODO: get rid of?

type signupRequest struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
	Email          string `json:"email"`
}

type adminSignupRequest struct {
	AdminUsername       string `json:"adminUsername"`
	HashedAdminPassword string `json:"hashedAdminPassword"`
	signupRequest
}

func SignupHandler(proxyHandler http.Handler, dbUser, dbPass string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			proxyHandler.ServeHTTP(w, r)
		case http.MethodPost:
			bs, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
				return
			}
			if r.Header.Get("isAdmin") == "true" {
				req := adminSignupRequest{}
				if err = json.Unmarshal(bs, &req); err != nil {
					http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
					return
				}
				if req.AdminUsername != dbUser {
					http.Error(w, "admin credential mismatch: "+err.Error(), http.StatusForbidden)
					return
				}
				expectedHashedDbPass, err := HashPassword("", dbPass)
				if err = json.Unmarshal(bs, &req); err != nil {
					http.Error(w, "failed to get expected db pass: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if req.HashedAdminPassword != expectedHashedDbPass {
					http.Error(w, "admin credential mismatch: "+err.Error(), http.StatusForbidden)
					return
				}
				// try to sign up with new creds
				err = CreateUser(r.Context(), req.Email, req.Username, &req.HashedPassword, nil)
				if err != nil {
					http.Error(w, "failed to create user with admin creds: "+err.Error(), http.StatusForbidden)
					return
				}
			}
			// TODO: if signup is google, create account
			// TODO: if signup user/pass
			////// TODO: create signupCode
			////// TODO: Create user/hashedPass/email entry in signupSessionsMap (with a timeout)
			////// TODO: Send email to user so that they can click the link with the signupCode
			// TODO: take user and hashedPass (maybe email?)
			// TODO:
		default:
			http.Error(w, "Unsupported http request method: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

var ConfirmSignupHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	endpt := r.PathValue("token")
	// TODO: THIS
})
