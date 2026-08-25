package api

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/reeceappling/goUtils/v2/logging"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/sessions/genericsessions"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

//func newOauthConfig() *oauth2.Config { // TODO: CHANGE OR DELETE
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
		store: statefulSessionStore{ // TODO: switch to stateless if we need to scale!
			sessMap:        map[SessionId]genericsessions.Session[ResolvedUserPerms]{},
			UserSessionMap: map[string]SessionId{},
		},
		ttl:     utils.Default(sessionTTL, 2*time.Hour),
		RWMutex: &sync.RWMutex{},
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
		sessionsCleanupFreq := time.Minute * 2
		sessionTTL := time.Hour * 1
		svc = NewAuthService(&sessionsCleanupFreq, &sessionTTL)
		return svc
	}
	return svc
}

type sessionStoreNotFoundError struct {
	typ string
}

func (e sessionStoreNotFoundError) Error() string {
	return fmt.Sprintf(e.typ+" not found in storage", e.typ)
}
func isStoreNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.As(err, &sessionStoreNotFoundError{})
}

var ErrUserSessionNonexistent = sessionStoreNotFoundError{typ: "user session"}
var ErrSessNonexistentSession = sessionStoreNotFoundError{typ: "session id"}

type sessionStore interface {
	SetSession(SessionId, genericsessions.Session[ResolvedUserPerms]) error
	GetSessionMap(SessionId) (genericsessions.Session[ResolvedUserPerms], error)
	SetSessionMap(SessionId, genericsessions.Session[ResolvedUserPerms]) error
	DeleteSessionMap(SessionId) error
	GetUserSessionMap(email string) (genericsessions.Session[ResolvedUserPerms], error)
	GetUserSessionMapId(email string) (s SessionId, err error)
	SetUserSessionMap(email string, id SessionId) error
	DeleteUserSessionMap(email string) error
	DeleteSession(email string) error
	SessionsToDelete() (out utils.Set[string], err error)
	AddProjectToUserSession(email string, project projectName, userPerm ProjectPerm) error
	RemoveProjectFromUserSession(email string, project projectName) error
}

type statefulSessionStore struct {
	sessMap        map[SessionId]genericsessions.Session[ResolvedUserPerms]
	UserSessionMap map[string]SessionId
}

func (storage statefulSessionStore) GetSessionMap(id SessionId) (s genericsessions.Session[ResolvedUserPerms], err error) {
	var ok bool
	s, ok = storage.sessMap[id]
	if ok {
		return s, nil
	}
	return s, ErrSessNonexistentSession
}
func (storage statefulSessionStore) GetUserSessionMap(email string) (s genericsessions.Session[ResolvedUserPerms], err error) {
	id, err := storage.GetUserSessionMapId(email)
	if err != nil {
		return s, err
	}
	return storage.GetSessionMap(id)
}
func (storage statefulSessionStore) GetUserSessionMapId(email string) (s SessionId, err error) {
	if id, ok := storage.UserSessionMap[email]; ok {
		return id, nil
	}
	return s, ErrUserSessionNonexistent
}
func (storage statefulSessionStore) SetSessionMap(id SessionId, s genericsessions.Session[ResolvedUserPerms]) error {
	storage.sessMap[id] = s
	return nil
}
func (storage statefulSessionStore) SetUserSessionMap(email string, id SessionId) error {
	storage.UserSessionMap[email] = id
	return nil
}
func (storage statefulSessionStore) DeleteSessionMap(id SessionId) error {
	delete(storage.sessMap, id)
	return nil
}
func (storage statefulSessionStore) DeleteUserSessionMap(email string) error {
	delete(storage.UserSessionMap, email)
	return nil
}
func (storage statefulSessionStore) DeleteSession(email string) error {
	id, ok := storage.UserSessionMap[email]
	if !ok {
		return errors.New("user email not found in storage")
	}
	return errors.Join(
		storage.DeleteUserSessionMap(email),
		storage.DeleteSessionMap(id),
	)
}
func (storage statefulSessionStore) SessionsToDelete() (out utils.Set[string], err error) {
	out = utils.Set[string]{}
	now := time.Now()
	for _, sess := range storage.sessMap {
		if sess.Expiry.Before(now) {
			out.Add(sess.Data.Email)
		}
	}
	return out, nil
}
func (storage statefulSessionStore) SetSession(id SessionId, updatedSess genericsessions.Session[ResolvedUserPerms]) error {
	storage.sessMap[id] = updatedSess
	storage.UserSessionMap[updatedSess.Data.Email] = id
	return nil
}
func (storage statefulSessionStore) AddProjectToUserSession(email string, project projectName, userPerm ProjectPerm) error {
	sessId, err := storage.GetUserSessionMapId(email)
	if err != nil {
		return err
	}
	sess, err := storage.GetSessionMap(sessId)
	if err != nil {
		return err
	}
	tempProjects := sess.Data.Projects
	if tempProjects == nil {
		tempProjects = map[projectName]*UserProjectPerm{}
	}
	tempProjects[project] = userPerm.UserProjectPerm()
	sess.Data.Projects = tempProjects
	return storage.SetSessionMap(sessId, sess)
}
func (storage statefulSessionStore) RemoveProjectFromUserSession(email string, project projectName) error {
	sessId, err := storage.GetUserSessionMapId(email)
	if err != nil {
		return err
	}
	existingSess, err := storage.GetSessionMap(sessId)
	if err != nil {
		if isStoreNotFoundError(err) {
			return nil
		}
		return err
	}
	// Delete project from the session and re-save the session in storage
	delete(existingSess.Data.Projects, project)
	return storage.SetSessionMap(sessId, existingSess)
}

type AuthService struct {
	store         sessionStore // TODO: this is currently stateful, if we ever need multiple servers we need it to be stateless, so offload to a centralized (or read-replica?) db or redis instance.
	ttl           time.Duration
	*sync.RWMutex // This struct MUST be used as a pointer // TODO: HATE how the mutexes are used in here..
}

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrExpired         = errors.New("session expired")
	ErrBadSession      = errors.New("bad session found in storage")
)

func (srv *AuthService) LogoutSession(sessId SessionId) error {
	//srv.Lock()// TODO: fix
	//defer srv.Unlock()// TODO: fix
	res := srv.GetSession(sessId, false) // TODO: misuses lock!
	if res.Err != nil {
		return res.Err
	}
	if res.Item.Data.Email == GuestEmail() {
		if err := srv.deleteGuestSession(sessId); err != nil {
			// TODO: handle error!
			println("failed to delete guest session: " + err.Error())
		}
	} else {
		if err := srv.deleteSession(res.Item.Data.Email); err != nil {
			// TODO: handle error!
			println("failed to delete session: " + err.Error())
		}
	}
	return nil
}

const maxSessionIdGenerationTries = 5 // TODO: max num?

func (srv *AuthService) newSessionIdForUserWithoutLock(email string) (SessionId, error) {
	_, err := srv.store.GetUserSessionMap(email)
	if err == nil {
		return "", errors.New("email already has existing session") // TODO: dont like. Maybe remove the email and their old session?
	}
	if !isStoreNotFoundError(err) {
		// failed to get session but unsure if exists!
		return "", errors.Join(err, errors.New("failed to check session with email"))
	}

	for i := 0; i < maxSessionIdGenerationTries; i++ { // TODO: num ok?
		var temp SessionId
		temp, err = generateSessionId()
		if err != nil {
			return "", err
		}
		_, err = srv.store.GetSessionMap(temp)
		if err == nil {
			err = errors.New("session already exists")
		} else {
			if isStoreNotFoundError(err) {
				// Does not exist yet, return
				return temp, nil
			}
			err = errors.Join(err, errors.New("failed to check new session id"))
		}
		continue
	}
	return "", errors.Join(errors.New("failed to generate a new unused session id"), err)
}

func (srv *AuthService) clearOldSessions() {
	srv.RLock()
	toDelete, err := srv.store.SessionsToDelete()
	srv.RUnlock()
	if err != nil {
		// TODO: what here?
		panic(err)
	}
	err = nil
	if len(toDelete) > 0 {
		for user, _ := range toDelete {
			err = errors.Join(err, srv.deleteSession(user))
		}
	}
	return
}

func (srv *AuthService) GetSession(id SessionId, refreshTTL bool) utils.Result[genericsessions.Session[ResolvedUserPerms]] {

	srv.RLock()
	sess, err := srv.store.GetSessionMap(id)
	srv.RUnlock()
	if err != nil {
		if isStoreNotFoundError(err) {
			return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](ErrSessionNotFound)
		}
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](errors.Join(err, errors.New("storage retrieval error")))
	}
	wg := &sync.WaitGroup{}
	if sess.Expiry.Before(time.Now()) {
		var e error = nil
		wg.Add(1)
		go func() { // TODO: seems unnecessary
			e = srv.deleteSession(sess.Data.Email)
			if e != nil {
				println("failed to delete expired session: " + err.Error())
				// TODO; what here?
			}
			wg.Done()
		}()

		wg.Wait()
		return utils.ErroredResult[genericsessions.Session[ResolvedUserPerms]](ErrExpired)
	}
	if refreshTTL {
		result := genericsessions.Session[ResolvedUserPerms]{}
		wg.Add(1)
		go func() {
			result = srv.setRefreshedSession(id, sess)
			wg.Done()
		}()
		wg.Wait()
		return utils.ResultFrom(result, nil)
	}
	return utils.SuccessfulResult(sess)

}

func (srv *AuthService) setRefreshedSession(id SessionId, sess genericsessions.Session[ResolvedUserPerms]) genericsessions.Session[ResolvedUserPerms] {
	result := &genericsessions.Session[ResolvedUserPerms]{}
	srv.Lock()
	defer srv.Unlock()
	updatedSess := sess.WithUpdatedExpiry(srv.ttl)
	*result = updatedSess
	err := srv.store.SetSession(id, updatedSess)
	if err != nil {
		panic("failed to set session: " + err.Error()) // TODO: what here?
	}
	return *result
}

func (srv *AuthService) createSessionFor(authinf ResolvedUserPerms) (SessionId, genericsessions.Session[ResolvedUserPerms], error) {
	var sess = genericsessions.Session[ResolvedUserPerms]{}
	if srv == nil {
		return "", sess, errors.New("nil auth service")
	}
	email := authinf.Email
	sess = genericsessions.Session[ResolvedUserPerms]{
		Data:   authinf,
		Expiry: time.Now().Add(srv.ttl),
	}
	srv.Lock()
	defer srv.Unlock()
	id, err := srv.newSessionIdForUserWithoutLock(email)
	if err != nil {
		return "", sess, errors.Join(err, errors.New("session with that ID already exists"))
	}
	err = srv.store.SetSession(id, sess)
	return id, sess, err
}

func (srv *AuthService) newFakeEmail() (string, error) {
	for i := 0; i < 5; i++ {
		fakeEmailBytes := make([]byte, 30)
		_, err := rand.Read(fakeEmailBytes)
		if err != nil {
			return "", err
		}
		email := "g-" + string(fakeEmailBytes)
		_, err = srv.store.GetUserSessionMap(email)
		if err == nil {
			continue
		}
		if !isStoreNotFoundError(err) {
			return "", errors.Join(err, errors.New("storage retrieval error"))
		}
		return email, nil
	}
	return "", errors.New("failed to generate fake email")
}

func (srv *AuthService) deleteSession(email string) error {
	srv.Lock()
	defer srv.Unlock()
	return srv.store.DeleteSession(email)
}

func (srv *AuthService) deleteGuestSession(id SessionId) error {
	srv.Lock()
	defer srv.Unlock()
	return srv.store.DeleteSessionMap(id)
}

// TODO: function to add/remove email project perms if they exist

func (serv *AuthService) TryToReAuth(ctx context.Context, sessionKey SessionId) (genericsessions.Session[ResolvedUserPerms], error) {
	if sessionKey == "" {
		env.LogIfDev(ctx, "sessionKey is empty")
		return genericsessions.Session[ResolvedUserPerms]{}, ErrBlankSessionKey
	}
	res := serv.GetSession(sessionKey, true)
	if res.Err != nil {
		env.LogIfDev(ctx, "failed to get session in TryToReAuth")
		return genericsessions.Session[ResolvedUserPerms]{}, utils.NotFound
	}
	env.LogIfDev(ctx, "reauthed user "+res.Item.Data.Email)
	bs, err := json.MarshalIndent(res.Item.Data, "", " ")
	if err == nil {
		env.LogIfDev(ctx, string(bs))
	}

	return *res.Item, nil
}

func (serv *AuthService) SessionForEmail(email string) (session SessionId, err error) {
	serv.RLock()
	id, err := serv.store.GetUserSessionMapId(email)
	serv.RUnlock()
	if err != nil {
		if isStoreNotFoundError(err) {
			return "", utils.NotFound
		} else {
			return "", errors.Join(err, errors.New("failed to get session with email"))
		}
	}
	return id, err
}

var UserWhitelist = utils.Set[string]{}

func (serv *AuthService) SigninGoogleAuthedUser(ctx context.Context, oauthUser goth.User) (sessionId SessionId, email string, err error) {
	log := logging.GetSugaredLogger(ctx) // TODO: ensure sugared logger is properly set up
	var u User
	email = oauthUser.Email
	adminEmail := os.Getenv("ADMIN_GMAIL") // TODO: del! This would allow an attacker with server access to just change an env var!
	coll := DbFrom(ctx).Collection(UserCollName)
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, email)).Decode(&u)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			err = errors.Join(errors.New("failed to get user. May exist?"), err)
			log.Errorw("failed to get user. May exist?", "error", err)
			env.LogIfDev(ctx, err.Error()) // TODO: del?
			return "", oauthUser.Email, err
		}
		if !UserWhitelist.Contains(email) && email != adminEmail {
			err = errors.New("user does not exist and is not on the account creation whitelist!")
			log.Errorw("user does not exist and is not on the account creation whitelist!", "error", err)
			env.LogIfDev(ctx, err.Error()) // TODO: del?
			return "", email, err
		}
		u = User{
			Email: email,
			Perms: UserPerms{
				Admin:    AcctTypeNormal(),
				Projects: []projectName{},
			},
		}

		if adminEmail != "" && email == adminEmail {
			log.Infow("Creating Admin user for email: " + adminEmail)
			u.Perms = UserPerms{
				Admin:    AcctTypeAdmin(),
				Projects: []projectName{},
			}
		} else {
			log.Infow("Creating Non-Admin user for email: " + u.Email)
		}

		_, err = coll.InsertOne(ctx, u)
		if err != nil {
			log.Infow("Failed to add user for email: " + u.Email)
			return "", email, err
		}
		// remove user from account creation whitelist
		if email != adminEmail {
			UserWhitelist.Remove(email) // TODO: ensure ok
		}
	}
	//if u.Perms.Admin == nil { // TODO: del or reenable for testing
	//	env.LogIfDev(ctx, "Admin on perms was nil when it should not have been!")
	//} else {
	//	adm := "was not admin"
	//	if *u.Perms.Admin {
	//		adm = "was admin"
	//	}
	//	env.LogIfDev(ctx, "Admin on perms "+adm)
	//}
	sessionId, _, err = serv.registerSessionAndResolvePerms(ctx, u)
	return
}

func (serv *AuthService) SigninPassAuthedUser(ctx context.Context, u User) (sessionId SessionId, err error) {
	sessionId, _, err = serv.registerSessionAndResolvePerms(ctx, u)
	return
}

func (serv *AuthService) SigninGuestUser() (sessionId SessionId, err error) {
	if serv == nil {
		return "", errors.New("nil auth service")
	}
	serv.Lock()
	defer serv.Unlock()
	email, err := serv.newFakeEmail()
	if err != nil {
		return "", err
	}

	sess := genericsessions.Session[ResolvedUserPerms]{
		Data: ResolvedUserPerms{
			Email:       email,
			AccountType: AcctTypeGuest(),
			Projects:    nil,
		},
		Expiry: time.Now().Add(serv.ttl),
	}

	id, err := serv.newSessionIdForUserWithoutLock(email)
	if err != nil {
		return "", errors.Join(err, errors.New("session with that ID already exists"))
	}
	err = serv.store.SetSession(id, sess)
	return id, err
}
func (serv *AuthService) SigninTestUser(email string) (sessionId SessionId, err error) {
	if serv == nil {
		return "", errors.New("nil auth service")
	}
	serv.Lock()
	defer serv.Unlock()
	var projsPerms map[projectName]*UserProjectPerm
	switch email {
	case testUserEmailPAA, testUserEmailPAB, testUserEmailPAC: // Can write (admin)
		upp := UserProjectPerm(true)
		projsPerms = map[projectName]*UserProjectPerm{
			TestProjectNamePublic: &upp,
		}
	case testUserEmailPWA, testUserEmailPWB, testUserEmailPWC: // Can write
		upp := UserProjectPerm(false)
		projsPerms = map[projectName]*UserProjectPerm{
			TestProjectNamePublic: &upp,
		}
	case testUserEmailPRA, testUserEmailPRB, testUserEmailPRC: // Can read
		projsPerms = map[projectName]*UserProjectPerm{
			TestProjectNamePublic: nil,
		}
	case testUserEmailPNA, testUserEmailPNB, testUserEmailPNC: // No specific projects for user
		projsPerms = map[projectName]*UserProjectPerm{}
	default:
		return "", errors.New("invalid test user email")
	}
	sess := genericsessions.Session[ResolvedUserPerms]{
		Data: ResolvedUserPerms{
			Email:       email,
			AccountType: AcctTypeNormal(),
			Projects:    projsPerms,
		},
		Expiry: time.Now().Add(serv.ttl),
	}

	id, err := serv.newSessionIdForUserWithoutLock(email)
	if err != nil {
		return "", errors.Join(err, errors.New("session with that ID already exists"))
	}
	err = serv.store.SetSession(id, sess)
	return id, err
}

func generateSessionId() (SessionId, error) {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	return SessionId(b), err
}

func (serv *AuthService) registerSessionAndResolvePerms(ctx context.Context, usr User) (sessionId SessionId, sess genericsessions.Session[ResolvedUserPerms], err error) {
	// Check if email already has a session
	_, err = serv.SessionForEmail(usr.Email)
	if err == nil {
		// User already exists! Remove old one
		err = serv.deleteSession(usr.Email)
		if err != nil {
			return "", sess, err
		}
	}
	var resolvedPerms = ResolvedUserPerms{
		Email:       usr.Email,
		AccountType: usr.Perms.Admin,
		// Projects are set later for normal users
	}
	if usr.Perms.Admin.IsGuest() {
		return serv.createSessionFor(ResolvedUserPerms{
			Email:       usr.Email,
			AccountType: nil,
			Projects:    nil, // Guests have no projects
		})
	} else {
		resolvedPerms, err = usr.ResolvePerms(ctx)
		if err != nil {
			return "", sess, err
		}
		return serv.createSessionFor(resolvedPerms)
	}
}

//// TODO: ON ENTRY PERMS CHANGE, that affect each email, modify email session perms?

var (
	ErrBlankSessionKey = errors.New("blank session key")
)

const (
	AuthPermsContextHeaderKey = "Auth-Info"
)

// TODO: UNUSED RIGHT NOW, USE FOR NATIVE APP?
func MiddlewareAuthOnContext(next http.Handler, notLoggedInHandler http.Handler, errorHandler func(error) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		svc := GetAuthService(ctx)
		var sess genericsessions.Session[ResolvedUserPerms]
		sessionId, err := SessionIdFromRequest(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		} else {
			env.LogIfDev(ctx, string("Trying to reauth session ID "+sessionId))
			sess, err = svc.TryToReAuth(r.Context(), sessionId)
			if err != nil {
				env.LogIfDev(ctx, "failed to reAuth: "+err.Error())
				if errors.Is(err, ErrBlankSessionKey) || errors.Is(err, utils.NotFound) {
					if notLoggedInHandler != nil {
						notLoggedInHandler.ServeHTTP(w, r)
					} else {
						next.ServeHTTP(w, r)
					}
					return
				}

				errorHandler(err).ServeHTTP(w, r)
				return
			}
		}
		ctxWithAuthInfo := SetAuthInfo(r.Context(), sess.Data)

		// TODO: ensure session cookies persist!
		next.ServeHTTP(w, r.WithContext(ctxWithAuthInfo))
	})
}

func AuthSplitterMiddleware() func(http.Handler, http.Handler, func(error) http.Handler) http.Handler {
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
			var sess genericsessions.Session[ResolvedUserPerms]
			sessionId, err := SessionIdFromRequest(r)
			if err != nil {
				handleNoSessionCookie.ServeHTTP(w, r)
				return
				////env.LogIfDev(ctx, "no session id found on request. Signing in as guest")
				////sessionId, err = svc.SigninGuestUser()
				////if err != nil {
				////	env.LogIfDev(ctx, "Failed to sign in as guest: "+err.Error())
				////	handleAuthErr(err).ServeHTTP(w, r)
				////	return
				////}
				////var ok bool
				////svc.RLock()
				////sess, ok = svc.sessMap[sessionId]
				////svc.RUnlock()
				////if !ok {
				////	env.LogIfDev(ctx, "guest not found in session map")
				////	handleAuthErr(err).ServeHTTP(w, r)
				////	return
				////}
				//err = gothic.StoreInSession(SessionIdKey, string(sessionId), r, w)
				//if err != nil {
				//	e := errors.Join(errors.New("sessId storage fail"), err)
				//	env.LogIfDev(ctx, e.Error())
				//	handleAuthErr(err).ServeHTTP(w, r)
				//	return
				//}
			} else {
				env.LogIfDev(ctx, string("Trying to reauth session ID "+sessionId))
				sess, err = svc.TryToReAuth(r.Context(), sessionId)
				if err != nil {
					env.LogIfDev(ctx, "failed to reAuth: "+err.Error())
					if errors.Is(err, ErrBlankSessionKey) || errors.Is(err, utils.NotFound) {
						handleNoSessionCookie.ServeHTTP(w, r)
						return
					}

					handleAuthErr(err).ServeHTTP(w, r)
					return
				}
			}
			ctxWithAuthInfo := SetAuthInfo(r.Context(), sess.Data)

			// TODO: ensure session cookies persist!
			onAuthed.ServeHTTP(w, r.WithContext(ctxWithAuthInfo))
		})
	}
}
func authAdminSplitterMiddleware() func(http.Handler, http.Handler, func(error) http.Handler) http.Handler {
	return func(onAuthed, handleNoSessionCookie http.Handler, handleAuthErr func(error) http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			svc := GetAuthService(ctx)
			sessionId, err := SessionIdFromRequest(r)
			if err != nil {
				env.LogIfDev(ctx, "no session id found on request: "+err.Error())
				handleAuthErr(err).ServeHTTP(w, r)
				return
			}
			env.LogIfDev(ctx, string("Trying to reauth session ID "+sessionId))
			sess, err := svc.TryToReAuth(r.Context(), sessionId)
			if err != nil {
				env.LogIfDev(ctx, "failed to reAuth: "+err.Error())
				handleAuthErr(err).ServeHTTP(w, r)
				return
			}
			if !sess.Data.IsAdmin() {
				env.LogIfDev(ctx, "non-admin tried to access admin area")
				http.Error(w, "Access Denied. Admin area", http.StatusForbidden)
				return
			}
			// TODO: ensure session cookies persist!
			onAuthed.ServeHTTP(w, r.WithContext(SetAuthInfo(r.Context(), sess.Data)))
		})
	}
}
func GetResolvedUserPerms(ctx context.Context) (ResolvedUserPerms, error) {
	usr, ok := ctx.Value(AuthPermsContextHeaderKey).(ResolvedUserPerms)
	if !ok {
		env.LogIfDev(ctx, "no auth info on context")
		return ResolvedUserPerms{}, errors.New("no auth info on context")
	}

	return usr, nil
}

func GetAuthInfo(ctx context.Context) (ResolvedUserPerms, error) {
	return GetResolvedUserPerms(ctx)
}
func GetUserEmail(ctx context.Context) string {
	user, _ := GetAuthInfo(ctx)
	return user.Email
}
func GetUserEmailPtr(ctx context.Context) *string {
	user, _ := GetAuthInfo(ctx)
	return &user.Email
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
		finalQuery := r.URL.Query()
		finalQuery.Set("destination", r.URL.String())
		http.Redirect(w, r, redirectUrl+"?"+finalQuery.Encode(), http.StatusTemporaryRedirect)
	})
	return func(nextHandler http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(AuthSplitterMiddleware()(nextHandler, redirectHandler, customDenyHandler))
	}
}

func (serv *AuthService) BasicSplitterMiddleware() func(http.Handler, http.Handler) http.Handler {
	return func(authSuccess, authFailure http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(AuthSplitterMiddleware()(authSuccess, authFailure, func(error) http.Handler {
			return authFailure
		}))
	}
}

var denyHandler = customDenyHandler(errors.New("Denied access"))

func customDenyHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied access or nonexistent: "+err.Error(), http.StatusForbidden)
		return
	})
}

func (serv *AuthService) AuthOrDenyMiddleware(nextHandler http.Handler) http.Handler {
	return serv.necessaryFirstMiddleware(AuthSplitterMiddleware()(nextHandler, denyHandler, func(error) http.Handler { return denyHandler }))
}
func (serv *AuthService) AuthAdminOrDenyMiddleware(nextHandler http.Handler) http.Handler {
	return serv.necessaryFirstMiddleware(authAdminSplitterMiddleware()(nextHandler, denyHandler, func(error) http.Handler { return denyHandler }))
}

func (serv *AuthService) OnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, authServiceContextKey, serv)
}

const SessionIdKey = "SessionId"

var ErrNoValidSessionOnRequest = errors.New("no valid session on request")

func SessionIdFromRequest(r *http.Request) (SessionId, error) {
	sessId, err := gothic.GetFromSession(SessionIdKey, r)
	if err != nil {
		return "", errors.Join(ErrNoValidSessionOnRequest, err)
	}
	return SessionId(sessId), nil
}
