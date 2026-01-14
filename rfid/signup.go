package rfid

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

//var googleOauthConfig = &oauth2.Config{
//	RedirectURL:  "http://localhost:8000/auth/google/callback", // TODO: FIX
//	ClientID:     "{PATTERN}.apps.googleusercontent.com",       // TODO: FIX
//	ClientSecret: "{SECRET}",                                   // TODO: FIX
//	Scopes: []string{
//		"https://www.googleapis.com/auth/userinfo.email",
//		"https://www.googleapis.com/auth/userinfo.profile",
//	},
//	Endpoint: google.Endpoint,
//}

//func SignupHandler() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		bs, err := io.ReadAll(r.Body)
//		if err != nil {
//			http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
//			return
//		}
//
//		googleToken := string(bs)
//		userBs, err := getUserDataFromGoogle(googleToken)
//		if err != nil {
//			http.Error(w, "failed to get email data: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		w.Write(userBs) // TODO: FIX!
//		userEmail := ""
//
//		// TODO: if signup is google, create account
//		// TODO: if signup email/pass
//		////// TODO: create signupCode
//		////// TODO: Create email/hashedPass/email entry in signupSessionsMap (with a timeout)
//		////// TODO: Send email to email so that they can click the link with the signupCode
//		// TODO: take email and hashedPass (maybe email?)
//		// TODO:
//	}
//}

//func getUserDataFromGoogle(code string) ([]byte, error) {
//	// Use code to get token and get email info from Google.
//	token, err := googleOauthConfig.Exchange(context.Background(), code)
//	if err != nil {
//		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
//	}
//
//	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken) // TODO: FIX
//	if err != nil {
//		return nil, fmt.Errorf("failed getting email info: %s", err.Error())
//	}
//	defer response.Body.Close()
//	contents, err := ioutil.ReadAll(response.Body)
//	if err != nil {
//		return nil, fmt.Errorf("failed read response: %s", err.Error())
//	}
//
//	saveUser(contents)
//	saveToken(contents, token)
//	return contents, nil
//}
//
//var ConfirmSignupHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	endpt := r.PathValue("token")
//	// TODO: THIS
//})
