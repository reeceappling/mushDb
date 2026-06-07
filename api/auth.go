package api

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/reeceappling/goUtils/v2/logging"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/sessions/genericsessions"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

//func newOauthConfig() *oauth2.Config { // TODO: CHANGE
//	return &oauth2.Config{
//		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
//		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
//		RedirectURL:  "http://api/authCallback", // TODO: ensure ok!
//		Scopes: []string{
//			"https://www.googleapis.com/auth/userinfo.profile",
//			"https://www.googleapis.com/auth/userinfo.email"},
//		Endpoint: google.Endpoint,
//	}
//}

var sessionAesCipher cipher.AEAD // 32 bytes for AES-256

func init() {
	// Generate AES key for this run
	key := make([]byte, 32) // 32 bytes for AES-256
	_, err := rand.Read(key)
	if err != nil {
		panic("failed to generate random session encryption key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic("failed to generate random session encryption key")
	}
	sessionAesCipher, err = cipher.NewGCM(block)
	if err != nil {
		panic("failed to generate random session encryption key")
	}
}

const authServiceContextKey = "authenticationService"

type SessionId string

// TODO: DONE ON CONTEXT WILL STOP CLEANUP?

func NewAuthService(sessionsCleanupFreq, sessionTTL *time.Duration) *AuthService {
	out := &AuthService{
		sessMap:        map[SessionId]genericsessions.Session[ResolvedUserPerms]{},
		UserSessionMap: map[string]SessionId{},
		ttl:            utils.Default(sessionTTL, 2*time.Hour),
		RWMutex:        &sync.RWMutex{},
	}
	// Start cleanup thread
	t := time.NewTicker(utils.Default(sessionsCleanupFreq, 5*time.Minute))
	go func() {
		select {
		case <-t.C:
			// TODO: CLEANUP SESSIONS ON TIMEOUTS?????
			out.clearOldSessions() // TODO: ok?
		}
	}()
	return out
}

func GetAuthService(ctx context.Context) *AuthService {
	svc, ok := ctx.Value(authServiceContextKey).(*AuthService)
	if !ok || svc == nil {
		svc = NewAuthService(utils.Pointer(time.Minute*2), utils.Pointer(time.Hour*1)) // TODO: FIX!!!!
		return svc
	}
	return svc
}

type AuthService struct {
	sessMap        map[SessionId]genericsessions.Session[ResolvedUserPerms]
	UserSessionMap map[string]SessionId
	ttl            time.Duration
	*sync.RWMutex  // This struct MUST be used as a pointer // TODO: HATE how the mutexes are used in here
}

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrExpired         = errors.New("session expired")
	ErrBadSession      = errors.New("bad session found in storage")
)

func (srv *AuthService) LogoutSession(sessId SessionId) error {
	srv.Lock()
	defer srv.Unlock()
	res := srv.GetSession(sessId, false)
	if res.Err != nil {
		return res.Err
	}
	if res.Item.Data.Email == GuestEmail() {
		srv.deleteGuestSession(sessId)
	} else {
		srv.deleteSession(sessId, res.Item.Data.Email)
	}
	return nil
}

// TODO: ensure working
func (srv *AuthService) clearOldSessions() {
	srv.RLock()

	toDelete := map[SessionId]string{}
	now := time.Now()
	for sessid, sess := range srv.sessMap {
		if sess.Expiry.Before(now) {
			toDelete[sessid] = sess.Data.Email
		}
	}
	srv.RUnlock()
	if len(toDelete) > 0 {
		for id, user := range toDelete {
			srv.deleteSession(id, user)
		}
	}
	return
}

func (srv *AuthService) GetSession(id SessionId, refreshTTL bool) utils.Result[genericsessions.Session[ResolvedUserPerms]] { // TODO: PUSH THIS UPDATE TO THE REPO

	srv.RLock()
	sess, ok := srv.sessMap[id]
	if !ok {
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](ErrSessionNotFound)
	}
	wg := &sync.WaitGroup{}
	if sess.Expiry.Before(time.Now()) {
		wg.Add(1)
		go func() {
			srv.deleteSession(id, sess.Data.Email)
			wg.Done()
		}()
		srv.RUnlock()
		wg.Wait()
		// TODO; is this ok?
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](ErrExpired)
	}
	if refreshTTL {
		result := genericsessions.Session[ResolvedUserPerms]{}
		wg.Add(1)
		go func() {
			// TODO: FIXME!
			result = srv.setRefreshedSession(id, sess)
			wg.Done()
		}()

		srv.RUnlock()
		wg.Wait()
		return utils.ResultFrom(result, nil)
	}
	return utils.SuccessfulResult(sess)

}

// TODO: ensure ok
func (srv *AuthService) setRefreshedSession(id SessionId, sess genericsessions.Session[ResolvedUserPerms]) genericsessions.Session[ResolvedUserPerms] {
	result := &genericsessions.Session[ResolvedUserPerms]{}
	srv.Lock()
	defer srv.Unlock()
	updatedSess := sess.WithUpdatedExpiry(srv.ttl)
	*result = updatedSess
	srv.sessMap[id] = updatedSess
	srv.UserSessionMap[updatedSess.Data.Email] = id // TODO: ok?
	return *result
}

// TODO; ensure ok
func (srv *AuthService) addSessionIfNotExists(id SessionId, authinf ResolvedUserPerms) utils.Result[genericsessions.Session[ResolvedUserPerms]] {
	result := &genericsessions.Session[ResolvedUserPerms]{}
	srv.Lock()
	defer srv.Unlock()
	if _, authExists := srv.sessMap[id]; authExists {
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](errors.New("session with that ID already exists")) // TODO: dont like
	}
	if _, userHasSessionAlready := srv.UserSessionMap[authinf.Email]; userHasSessionAlready {
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](errors.New("email already has existing session")) // TODO: dont like. Maybe remove the email and their old session?
	}
	sess := genericsessions.Session[ResolvedUserPerms]{Data: authinf}
	updatedSess := sess.WithUpdatedExpiry(srv.ttl)
	*result = updatedSess
	srv.sessMap[id] = updatedSess
	srv.UserSessionMap[updatedSess.Data.Email] = id
	return utils.SuccessfulResult(*result)
}

// TODO; ensure ok
func (srv *AuthService) addGuestSessionIfNotExists(id SessionId) error {
	result := &genericsessions.Session[ResolvedUserPerms]{}
	srv.Lock()
	defer srv.Unlock()
	if _, exists := srv.sessMap[id]; exists {
		return errors.New("session with that ID already exists")
	}
	updatedSess := genericsessions.Session[ResolvedUserPerms]{Data: ResolvedUserPerms{Email: GuestEmail(), admin: nil}}.WithUpdatedExpiry(srv.ttl)
	*result = updatedSess
	srv.sessMap[id] = updatedSess
	return nil
}

func (srv *AuthService) deleteSession(id SessionId, email string) {
	srv.Lock()
	defer srv.Unlock()
	delete(srv.UserSessionMap, email)
	delete(srv.sessMap, id)
	return
}

func (srv *AuthService) deleteGuestSession(id SessionId) {
	srv.Lock()
	defer srv.Unlock()
	delete(srv.sessMap, id)
	return
}

// TODO: function to add/remove email project perms if they exist

func (serv *AuthService) TryToReAuth(sessionKey SessionId) (genericsessions.Session[ResolvedUserPerms], error) {
	if sessionKey == "" {
		println("sessionKey is empty") // TODO; del
		return genericsessions.Session[ResolvedUserPerms]{}, ErrBlankSessionKey
	}
	res := serv.GetSession(sessionKey, true) // TODO: needs update
	if res.Err != nil {
		println("failed to get session in TryToReAuth") // TODO; del
		return genericsessions.Session[ResolvedUserPerms]{}, utils.NotFound
	}
	println("reauthed user " + res.Item.Data.Email)
	return *res.Item, nil
}

// TODO: use this

func (serv *AuthService) SessionForEmail(email string) (session SessionId, err error) {
	serv.RLock()
	defer serv.RUnlock()

	sess, exists := serv.UserSessionMap[email]
	if !exists {
		return "", utils.NotFound
	}
	return sess, nil
}

// TODO: USE THIS

func (serv *AuthService) SigninGoogleAuthedUser(ctx context.Context, oauthUser goth.User) (sessionId SessionId, email string, err error) {
	var u User // TODO: get this from db
	email = oauthUser.Email
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(UserCollName)
	userResult := coll.FindOne(ctx, bsonFindFilter("_id", email))
	raw, _ := userResult.Raw() // TODO; del
	println(raw.String())      // TODO; del
	err = userResult.Decode(&u)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return "", oauthUser.Email, errors.Join(errors.New("failed to get user. May exist"), err)
		}
		// If not found, create user!
		u = User{
			Email: email,
			Perms: UserPerms{
				Admin:    nil,
				Projects: []projectName{},
			},
		}
		adminEmail := os.Getenv("ADMIN_GMAIL")
		if adminEmail != "" && email == adminEmail {
			println("Admin user signed up!") // TODO: DEL
			logging.GetLogger(ctx).Info("Admin user signed up with email " + adminEmail)
			u = User{
				Email: email,
				Perms: UserPerms{
					Admin:    utils.Pointer(true),
					Projects: []projectName{}, // TODO: ADD PROJECTS???
				},
			}
		}
		bss, _ := json.Marshal(u)              // TODO: del
		println("inserting user", string(bss)) // TODO: del
		_, err = coll.InsertOne(ctx, u)
		if err != nil {
			return "", email, err
		}
		if adminEmail != "" && email == adminEmail {
			if err = coll.FindOne(ctx, bsonFindFilter("_id", email)).Decode(&u); err != nil {
				println("failed to check Admin user")
				return "", email, err
			}
			if u.Perms.Admin == nil || !(*u.Perms.Admin) {
				return "", email, errors.New("result does not show Admin")
			}
		}
		println("user created, continuing")
	}
	if u.Perms.Admin == nil {
		println("Admin on perms was nil when it should not have been!")
	} else {
		println("Admin on perms was correct!")
	}
	sessionId, _, err = serv.registerSessionAndResolvePerms(ctx, u)
	return
}

// TODO: USE THIS!
func (serv *AuthService) SigninGuestUser(ctx context.Context) (sessionId SessionId, err error) {
	return serv.registerGuestSession(ctx)
}

func generateSessionId() (SessionId, error) {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	return SessionId(b), err
}

// TODO: ADD ABILITY FOR PROJECTS CHANGING TO CHANGE USERS
// TODO: MODIFY SO THAT WE CAN GRAB OLD PERMS IF THEY EXIST

// TODO: USE!
func (serv *AuthService) registerSessionAndResolvePerms(ctx context.Context, usr User) (sessionId SessionId, auths genericsessions.Session[ResolvedUserPerms], err error) {
	// TODO: RESOLVE ALL USER INFO FOR PROJECTS
	// Check if email already has a session
	sessId, err := serv.SessionForEmail(usr.Email)
	if err == nil {
		// User already exists! Remove old one
		serv.deleteSession(sessId, usr.Email)
	}
	sessId, err = generateSessionId() // TODO: make sure does not already exist!
	if err != nil {
		return
	}
	// Resolve auth info
	resolvedPerms, err := usr.ResolvePerms(ctx) // TODO; FIX?
	if err != nil {
		return "", auths, err
	}
	authsResult := serv.addSessionIfNotExists(sessId, resolvedPerms)
	if authsResult.Err != nil {
		return sessId, auths, authsResult.Err
	}
	serv.UserSessionMap[usr.Email] = sessId
	return sessId, *authsResult.Item, nil
}

// TODO: USE!
func (serv *AuthService) registerGuestSession(ctx context.Context) (sessionId SessionId, err error) {
	sessId, err := generateSessionId() // TODO: make sure does not already exist!
	if err != nil {
		return
	}
	err = serv.addGuestSessionIfNotExists(sessId)
	if err != nil {
		return sessId, err
	}
	return sessId, nil
}

//// TODO: ON PROJECT PERMS CHANGE, OR ENTRY PERMS CHANGE, that affect each email, modify email session perms
//// TODO: USE!
//func (serv *AuthService) changeSessionProjectPerms(projName projectName, newProjectPerms map[Base58Str]perms.Perm) { // TODO: USE THIS
//
//	for b58User, newPerms := range newProjectPerms {
//		serv.RLock()
//		sessId, exists := serv.UserSessionMap[b58User]
//		if !exists {
//			serv.RUnlock()
//			continue
//		}
//		session, exists := serv.sessMap[sessId]
//		serv.RUnlock()
//		if !exists {
//			continue
//		}
//		if session.Data.Opts == nil {
//			continue // TODO: ok?
//		}
//		// TODO: do we need to create Projects in opts if it does not exist?
//		serv.Lock()
//		if newPerms == perms.None {
//			delete(session.Data.Opts.Projects, projName)
//		} else {
//			session.Data.Opts.Projects[projName] = canWriteBoolForPerm(newPerms)
//		}
//		serv.Unlock()
//
//	}
//	//// TODO: delete all below?
//	//
//	//// TODO: SESSION ITERATOR
//	//authInfo := ResolvedUserPerms{} // TODO: if project is in session and perms do not match, change
//	//// TODO: make sure we arent mistakenly modifying the map somewhere else at the same time (rwMutex on authInfo?)
//	//if authInfo.Opts != nil {
//	//	for proj, newPerm := range newProjectPerms {
//	//		canWrite, exists := (*authInfo.Opts).Projects[proj]
//	//		if !exists {
//	//			continue
//	//		}
//	//		if newPerm == perms.None {
//	//			delete((*authInfo.Opts).Projects, proj)
//	//			continue
//	//		}
//	//		if canWrite != canWriteBoolForPerm(newPerm) {
//	//			(*authInfo.Opts).Projects[proj] = !canWrite
//	//		}
//	//	}
//	//}
//}

var (
	ErrBlankSessionKey = errors.New("blank session key")
)

const (
	AuthPermsContextHeaderKey = "Auth-Info"
	AuthSessionCookieKey      = "SessionId" // TODO: FIX THIS! I REMOVED IT SOMEWHERE!
)

// authSplitterMiddleware , if the email does not supply a session, will need to handle GetSessionCookie returning http.ErrNoCookie
//
// handleAuthErr is typically just a http.Handler that will redirect when err==http.ErrNoCookie and truly error out if err is !=nil otherwise

func authSplitterMiddleware() func(http.Handler, http.Handler, func(error) http.Handler) http.Handler {
	return func(onAuthed, handleNoSessionCookie http.Handler, handleAuthErr func(error) http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			//println("COOKIES!")
			//cs := r.Cookies()
			//if len(cs) == 0 {
			//	println("NO COOKIES FOUND")
			//}
			//for _, c := range cs {
			//	println(c.Name, c.Path, c.Domain, c.Value) // TODO: del
			//}
			svc := GetAuthService(ctx)
			sessionId, err := SessionIdFromRequest(r)
			if err != nil {
				println("no session id found on request", err.Error()) // TODO: del
				handleAuthErr(err).ServeHTTP(w, r)
				return
			}
			println("Trying to reauth session ID " + sessionId)
			sess, err := svc.TryToReAuth(sessionId)
			if err != nil {
				println("failed to reAuth", err.Error())
				handleAuthErr(err).ServeHTTP(w, r)
				return
			}
			// TODO: ensure session cookies persist!
			println("Serving next success handler")
			onAuthed.ServeHTTP(w, r.WithContext(SetAuthInfo(r.Context(), sess.Data)))
		})
	}
}
func GetResolvedUserPerms(ctx context.Context) (ResolvedUserPerms, error) {
	usr, ok := ctx.Value(AuthPermsContextHeaderKey).(ResolvedUserPerms)
	if !ok {
		println("no auth info on context")
		return ResolvedUserPerms{}, errors.New("no auth info on context")
	}

	return usr, nil
}

// TODO; retire this
func GetAuthInfo(ctx context.Context) (ResolvedUserPerms, error) {
	return GetResolvedUserPerms(ctx)
}

func SetAuthInfo(ctxIn context.Context, perms ResolvedUserPerms) context.Context {
	return context.WithValue(ctxIn, AuthPermsContextHeaderKey, perms)
}

func (serv *AuthService) necessaryFirstMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = serv.OnContext(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (serv *AuthService) AuthOrRedirectMiddleware(redirectUrl string) func(http.Handler) http.Handler {
	var redirectHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: PUT PAGE FROM IN THE DATA SO WE CAN PULL IT LATER?
		http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
	})

	//var redirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	// TODO: PUT PAGE FROM IN THE DATA SO WE CAN PULL IT LATER?
	//	http.Redirect(w, r, redirectUrl, http.StatusSeeOther) // TODO: status ok?
	//})
	return func(nextHandler http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(authSplitterMiddleware()(nextHandler, redirectHandler, customDenyHandler))
	}
}

func (serv *AuthService) BasicSplitterMiddleware() func(http.Handler, http.Handler) http.Handler {
	return func(authSuccess, authFailure http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(authSplitterMiddleware()(authSuccess, authFailure, func(error) http.Handler {
			return authFailure
		}))
	}
}

var denyHandler = customDenyHandler(errors.New("Forbidden"))

func customDenyHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied here"+err.Error(), http.StatusForbidden)
		return
	})
}

func (serv *AuthService) AuthOrDenyMiddleware(nextHandler http.Handler) http.Handler {
	return serv.necessaryFirstMiddleware(authSplitterMiddleware()(nextHandler, denyHandler, func(error) http.Handler { return denyHandler }))
}

func (serv *AuthService) OnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, authServiceContextKey, serv)
}

// TODO: use
func setSessionCookie(w http.ResponseWriter, r *http.Request, sessionId string, session genericsessions.Session[ResolvedUserPerms]) {
	http.SetCookie(w, &http.Cookie{
		Name:    AuthSessionCookieKey,
		Value:   sessionId,
		Quoted:  false,
		Path:    r.URL.Path, // TODO: ok?
		Domain:  r.URL.Host,
		Expires: session.Expiry,
		//RawExpires:  "",    // TODO: ????????????????
		MaxAge:   0, // TODO: ????????????????
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteNoneMode, // TODO: ok?
		//Partitioned: false,                // TODO: ????????????????
		//Raw:         "",                   // TODO: ????????????????
		//Unparsed:    nil,                  // TODO: ????????????????
	})
}

// TODO: RENAME AND USE
func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
	out, err := r.Cookie(AuthSessionCookieKey)
	if err != nil {
		return nil, errors.Join(http.ErrNoCookie, err)
	}
	return out, err
}

const SessionIdKey = "SessionId"

func SessionIdFromRequest(r *http.Request) (SessionId, error) {
	sessId, err := gothic.GetFromSession(SessionIdKey, r)
	if err != nil {
		return "", err
	}
	return SessionId(sessId), nil
}
