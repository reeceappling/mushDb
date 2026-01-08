package rfid

import (
	"context"
	"errors"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/sessions/genericsessions"
	"net/http"
	"sync"
	"time"
)

// var sessionAesCipher cipher.AEAD // 32 bytes for AES-256
//
//	func init() {
//		// Generate AES key for this run
//		key := make([]byte, 32) // 32 bytes for AES-256
//		_, err := rand.Read(key)
//		if err != nil {
//			panic("failed to generate random session encryption key")
//		}
//		block, err := aes.NewCipher(key)
//		if err != nil {
//			panic("failed to generate random session encryption key")
//		}
//		sessionAesCipher, err = cipher.NewGCM(block)
//		if err != nil {
//			panic("failed to generate random session encryption key")
//		}
//	}
//
// const authServiceContextKey = "authenticationService"
type SessionId string

// // TODO: DONE ON CONTEXT WILL STOP CLEANUP?
//
//	func NewAuthService(sessionsCleanupFreq, sessionTTL *time.Duration) *AuthService {
//		out := &AuthService{
//			sessMap:        map[SessionId]genericsessions.Session[AuthInfo]{},
//			UserSessionMap: map[Base58Str]SessionId{},
//			ttl:            utils.Default(sessionTTL, 2*time.Hour),
//			RWMutex:        &sync.RWMutex{},
//		}
//		// Start cleanup thread
//		t := time.NewTicker(utils.Default(sessionsCleanupFreq, 5*time.Minute))
//		go func() {
//			select {
//			case <-t.C:
//				// TODO: CLEANUP SESSIONS ON TIMEOUTS?????
//				out.clearOldSessions() // TODO: ok?
//			}
//		}()
//		return out
//	}
//
//	func GetAuthService(ctx context.Context) (*AuthService, error) {
//		svc, ok := ctx.Value(authServiceContextKey).(*AuthService)
//		if !ok {
//			return nil, errors.New("auth service not found")
//		}
//		if svc == nil {
//			return nil, errors.New("auth service is nil")
//		}
//		return svc, nil
//	}
type AuthService struct {
	sessMap        map[SessionId]genericsessions.Session[AuthInfo]
	UserSessionMap map[Base58Str]SessionId
	ttl            time.Duration
	*sync.RWMutex  // This struct MUST be used as a pointer // TODO: HATE how the mutexes are used in here
}

// var (
//
//	ErrSessionNotFound = errors.New("session not found")
//	ErrExpired         = errors.New("session expired")
//	ErrBadSession      = errors.New("bad session found in storage")
//
// )
//
//	func (srv *AuthService) clearOldSessions() {
//		srv.RLock()
//
//		toDelete := map[SessionId]Base58Str{}
//		now := time.Now()
//		for sessid, sess := range srv.sessMap {
//			if sess.Expiry.Before(now) {
//				toDelete[sessid] = sess.Data.Id.asBase58()
//			}
//		}
//		srv.RUnlock()
//		if len(toDelete) > 0 {
//			for id, user := range toDelete {
//				srv.deleteSession(id, user)
//			}
//		}
//		return
//	}
//
// func (srv *AuthService) getSession(id SessionId, refreshTTL bool) utils.Result[genericsessions.Session[AuthInfo]] { // TODO: PUSH THIS UPDATE TO THE REPO
//
//	srv.RLock()
//	sess, ok := srv.sessMap[id]
//	if !ok {
//		return utils.ErroredResult[genericsessions.Session[AuthInfo]](ErrSessionNotFound)
//	}
//	wg := &sync.WaitGroup{}
//	if sess.Expiry.Before(time.Now()) {
//		wg.Add(1)
//		go func() {
//			srv.deleteSession(id, sess.Data.Id.asBase58())
//			wg.Done()
//		}()
//		srv.RUnlock()
//		wg.Wait()
//		return utils.ErroredResult[genericsessions.Session[AuthInfo]](ErrExpired)
//	}
//	if refreshTTL {
//		out := &utils.Result[genericsessions.Session[AuthInfo]]{}
//		wg.Add(1)
//		go func() {
//			*out.Item = srv.setRefreshedSession(id, sess)
//			wg.Done()
//		}()
//		srv.RUnlock()
//		wg.Wait()
//		return *out // TODO: ok?
//	}
//	return utils.SuccessfulResult(sess)
//
// }
//
//	func (srv *AuthService) setRefreshedSession(id SessionId, sess genericsessions.Session[AuthInfo]) genericsessions.Session[AuthInfo] {
//		result := &genericsessions.Session[AuthInfo]{}
//		srv.Lock()
//		defer srv.Unlock()
//		updatedSess := sess.WithUpdatedExpiry(srv.ttl)
//		*result = updatedSess
//		srv.sessMap[id] = updatedSess
//		srv.UserSessionMap[updatedSess.Data.Id.asBase58()] = id // TODO: ok?
//		return *result
//	}
//
//	func (srv *AuthService) addSessionIfNotExists(id SessionId, authinf AuthInfo) utils.Result[genericsessions.Session[AuthInfo]] {
//		result := &genericsessions.Session[AuthInfo]{}
//		srv.Lock()
//		defer srv.Unlock()
//		if _, authExists := srv.sessMap[id]; authExists {
//			return utils.ErroredResult[genericsessions.Session[AuthInfo]](errors.New("session with that ID already exists")) // TODO: dont like
//		}
//		if _, userHasSessionAlready := srv.UserSessionMap[authinf.Id.asBase58()]; userHasSessionAlready {
//			return utils.ErroredResult[genericsessions.Session[AuthInfo]](errors.New("user already has existing session")) // TODO: dont like. Maybe remove the user and their old session?
//		}
//		sess := genericsessions.Session[AuthInfo]{Data: authinf}
//		updatedSess := sess.WithUpdatedExpiry(srv.ttl)
//		*result = updatedSess
//		srv.sessMap[id] = updatedSess
//		srv.UserSessionMap[updatedSess.Data.Id.asBase58()] = id // TODO: ok?
//		return utils.SuccessfulResult(*result)
//	}
//
//	func (srv *AuthService) deleteSession(id SessionId, userId Base58Str) {
//		srv.Lock()
//		defer srv.Unlock()
//		delete(srv.UserSessionMap, userId)
//		delete(srv.sessMap, id)
//		return
//	}
//
// // TODO: function to add/remove user project perms if they exist
//
//	func (serv *AuthService) TryToReAuth(sessionKey SessionId) (genericsessions.Session[AuthInfo], error) {
//		if sessionKey == "" {
//			return genericsessions.Session[AuthInfo]{}, ErrBlankSessionKey
//		}
//		res := serv.getSession(sessionKey, true) // TODO: needs update
//		if res.Err != nil {
//			return genericsessions.Session[AuthInfo]{}, utils.NotFound
//		}
//		return *res.Item, nil
//	}
//
// // TODO: use this
//
//	func (serv *AuthService) SessionForUserId(uid AlternateCollectionId) (session SessionId, err error) {
//		serv.RLock()
//		defer serv.RUnlock()
//
//		sess, exists := serv.UserSessionMap[uid.asBase58()]
//		if !exists {
//			return "", utils.NotFound
//		}
//		return sess, nil
//	}
//
// // TODO: USE THIS
//
//	func (serv *AuthService) TryToAuthUserPass(ctx context.Context, username, hashedPass string) (sessionId SessionId, usernOut string, err error) {
//		usernOut = username
//		var u User
//		err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(userCollName).FindOne(ctx, bson.M{
//			"_id": username}).Decode(&u)
//		if err != nil {
//			return "", "", err // TODO: ok?
//		}
//		username = u.Username
//		if u.HashedPass == nil || u.Salt == nil {
//			return "", "", errors.New("user does not have a password set up")
//		}
//		hashed, err := HashPassword(*u.Salt, hashedPass) // TODO: ensure ok
//		if *u.HashedPass != hashed {
//			return "", "", errors.New("password mismatch")
//		}
//		sessionId, _, err = serv.registerSessionAndResolvePerms(ctx, u)
//		return
//	}
//
// // TODO: USE THIS
//
//	func (serv *AuthService) TryToAuthGoogle(ctx context.Context, finalEndpoint, googleId string) (sessionId SessionId, username string, err error) {
//		var u User // TODO: get this from db
//		// TODO: DO THIS WHOLE THING
//		sessionId, _, err = serv.registerSessionAndResolvePerms(ctx, u)
//		return
//	}
//
//	func generateSessionId() (SessionId, error) {
//		b := make([]byte, 32)
//		_, err := io.ReadFull(rand.Reader, b)
//		return SessionId(b), err
//	}
//
// // TODO: ADD ABILITY FOR PROJECTS CHANGING TO CHANGE USERS
// // TODO: MODIFY SO THAT WE CAN GRAB OLD PERMS IF THEY EXIST
//
//	func (serv *AuthService) registerSessionAndResolvePerms(ctx context.Context, usr User) (sessionId SessionId, auths genericsessions.Session[AuthInfo], err error) {
//		// TODO: RESOLVE ALL USER INFO FOR PROJECTS
//		// Check if user already has a session
//		sessId, err := serv.SessionForUserId(usr.Id)
//		if err == nil {
//			// User already exists! Remove old one
//			serv.deleteSession(sessId, usr.Id.asBase58())
//		}
//		sessId, err = generateSessionId() // TODO: make sure does not already exist!
//		if err != nil {
//			return
//		}
//		// Resolve auth info
//		resolvedPerms, err := usr.ResolvePerms(ctx)
//		if err != nil {
//			return "", auths, err
//		}
//		authInfo := AuthInfo{
//			Id:   usr.Id,
//			Opts: resolvedPerms,
//		}
//		authsResult := serv.addSessionIfNotExists(sessId, authInfo)
//		if authsResult.Err != nil {
//			return sessId, auths, authsResult.Err
//		}
//		serv.UserSessionMap[usr.Id.asBase58()] = sessId
//		return sessId, *authsResult.Item, nil
//	}
//
// // TODO: ON PROJECT PERMS CHANGE, OR ENTRY PERMS CHANGE, that affect each user, modify user session perms
// func (serv *AuthService) changeSessionProjectPerms(projName projectName, newProjectPerms map[Base58Str]perms.Perm) { // TODO: USE THIS
//
//		for b58User, newPerms := range newProjectPerms {
//			serv.RLock()
//			sessId, exists := serv.UserSessionMap[b58User]
//			if !exists {
//				serv.RUnlock()
//				continue
//			}
//			session, exists := serv.sessMap[sessId]
//			serv.RUnlock()
//			if !exists {
//				continue
//			}
//			if session.Data.Opts == nil {
//				continue // TODO: ok?
//			}
//			// TODO: do we need to create projects in opts if it does not exist?
//			serv.Lock()
//			if newPerms == perms.None {
//				delete(session.Data.Opts.Projects, projName)
//			} else {
//				session.Data.Opts.Projects[projName] = canWriteBoolForPerm(newPerms)
//			}
//			serv.Unlock()
//
//		}
//		//// TODO: delete all below?
//		//
//		//// TODO: SESSION ITERATOR
//		//authInfo := AuthInfo{} // TODO: if project is in session and perms do not match, change
//		//// TODO: make sure we arent mistakenly modifying the map somewhere else at the same time (rwMutex on authInfo?)
//		//if authInfo.Opts != nil {
//		//	for proj, newPerm := range newProjectPerms {
//		//		canWrite, exists := (*authInfo.Opts).Projects[proj]
//		//		if !exists {
//		//			continue
//		//		}
//		//		if newPerm == perms.None {
//		//			delete((*authInfo.Opts).Projects, proj)
//		//			continue
//		//		}
//		//		if canWrite != canWriteBoolForPerm(newPerm) {
//		//			(*authInfo.Opts).Projects[proj] = !canWrite
//		//		}
//		//	}
//		//}
//	}
type AuthInfo struct {
	Id   AlternateCollectionId // ID in database for User
	Opts *UserPermsResolved    //admin?canCreateAccounts. // missing means no perms
}

//
//	func (auth AuthInfo) isAdmin() bool {
//		return auth.Opts != nil && (*auth.Opts).Admin != nil && *auth.Opts.Admin
//	}
//
//	func (auth AuthInfo) lowestPermBetweenEntries(entryPermsets ...Permissioned) perms.Perm {
//		out := perms.Write
//		for _, item := range entryPermsets {
//			thisPerm := item.Permissions().PermissionFor(auth)
//			if thisPerm < out {
//				if thisPerm == perms.None {
//					return perms.None
//				}
//				out = thisPerm
//			}
//		}
//		return out
//	}
//
// var (
//
//	ErrBlankSessionKey = errors.New("blank session key")
//
// )
//
// const (
//
//	AuthPermsContextHeaderKey = "Auth-Info"
//	AuthSessionCookieKey      = "SessionId"
//
// )
//
// // authSplitterMiddleware , if the user does not supply a session, will need to handle GetSessionCookie returning http.ErrNoCookie
// //
// // handleAuthErr is typically just a http.Handler that will redirect when err==http.ErrNoCookie and truly error out if err is !=nil otherwise
//
//	func authSplitterMiddleware(masterUser, masterPass string) func(http.Handler, http.Handler, func(error) http.Handler) http.Handler {
//		return func(onAuthed, handleNoSessionCookie http.Handler, handleAuthErr func(error) http.Handler) http.Handler {
//			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				ctx := r.Context()
//				svc, err := GetAuthService(ctx)
//				if err != nil {
//					err = errors.Join(errors.New("no auth service found. Endpoint unavailable"), err)
//					handleAuthErr(err).ServeHTTP(w, r)
//					return
//				}
//				var sessionId SessionId
//				cookie, err := GetSessionCookie(r)
//				if err == nil {
//					sessionId, err = SessionIdFromCookie(cookie)
//				}
//				if err != nil {
//					if errors.Is(err, http.ErrNoCookie) {
//						// TODO: remove cookie if needed?
//						handleNoSessionCookie.ServeHTTP(w, r)
//					} else {
//						handleAuthErr(err).ServeHTTP(w, r)
//					}
//					return
//				}
//
//				sess, err := svc.TryToReAuth(sessionId)
//				if err != nil {
//					handleAuthErr(err).ServeHTTP(w, r)
//					return
//				}
//				cookie.Expires = sess.Expiry
//				http.SetCookie(w, cookie)
//
//				onAuthed.ServeHTTP(w, r.WithContext(SetAuthInfo(r.Context(), sess.Data)))
//			})
//		}
//	}
//func GetAuthInfo(ctx context.Context) (AuthInfo, error) {
//	usr, ok := ctx.Value(AuthPermsContextHeaderKey).(AuthInfo)
//	if !ok {
//		return AuthInfo{}, errors.New("no auth info on context")
//	}
//
//	return usr, nil
//}

//func SetAuthInfo(ctxIn context.Context, info AuthInfo) context.Context {
//	return context.WithValue(ctxIn, AuthPermsContextHeaderKey, info)
//}
//
//func (serv *AuthService) necessaryFirstMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		next.ServeHTTP(w, r.WithContext(serv.OnContext(r.Context())))
//	})
//}
//
//func (serv *AuthService) AuthOrRedirectMiddleware(redirectUrl, masterUser, masterPass string) func(http.Handler) http.Handler {
//	var redirectHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// TODO: PUT PAGE FROM IN THE DATA SO WE CAN PULL IT LATER?
//		http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
//	})
//
//	//var redirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	//	// TODO: PUT PAGE FROM IN THE DATA SO WE CAN PULL IT LATER?
//	//	http.Redirect(w, r, redirectUrl, http.StatusSeeOther) // TODO: status ok?
//	//})
//	return func(nextHandler http.Handler) http.Handler {
//		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(nextHandler, redirectHandler, customDenyHandler))
//	}
//}
//
//func (serv *AuthService) BasicSplitterMiddleware(masterUser, masterPass string) func(http.Handler, http.Handler) http.Handler {
//	return func(authSuccess, authFailure http.Handler) http.Handler {
//		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(authSuccess, authFailure, func(error) http.Handler {
//			return authFailure
//		}))
//	}
//}
//
//var denyHandler = customDenyHandler(errors.New("Forbidden"))
//
//func customDenyHandler(err error) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		http.Error(w, err.Error(), http.StatusForbidden)
//		return
//	})
//}
//
//func (serv *AuthService) AuthOrDenyMiddleware(masterUser, masterPass string) func(http.Handler) http.Handler {
//	return func(nextHandler http.Handler) http.Handler {
//		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(nextHandler, denyHandler, func(error) http.Handler { return denyHandler }))
//	}
//}
//
//func (serv *AuthService) OnContext(ctx context.Context) context.Context {
//	return context.WithValue(ctx, authServiceContextKey, serv)
//}
//
//// TODO: use
//func setSessionCookie(w http.ResponseWriter, r *http.Request, sessionId string, session genericsessions.Session[AuthInfo]) {
//	http.SetCookie(w, &http.Cookie{
//		Name:    AuthSessionCookieKey,
//		Value:   sessionId,
//		Quoted:  false,
//		Path:    r.URL.Path, // TODO: ok?
//		Domain:  r.URL.Host,
//		Expires: session.Expiry,
//		//RawExpires:  "",    // TODO: ????????????????
//		MaxAge:   0, // TODO: ????????????????
//		Secure:   true,
//		HttpOnly: false,
//		SameSite: http.SameSiteNoneMode, // TODO: ok?
//		//Partitioned: false,                // TODO: ????????????????
//		//Raw:         "",                   // TODO: ????????????????
//		//Unparsed:    nil,                  // TODO: ????????????????
//	})
//}
//
//// TODO: RENAME AND USE
//func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
//	out, err := r.Cookie(AuthSessionCookieKey)
//	if err != nil {
//		return nil, errors.Join(http.ErrNoCookie, err)
//	}
//	return out, err
//}
//
//func SessionIdFromCookie(c *http.Cookie) (SessionId, error) {
//	if c == nil {
//		return "", errors.Join(errors.New("session cookie did not exist"), http.ErrNoCookie)
//	}
//	sessionName := c.Value
//	if sessionName == "" {
//		return "", errors.Join(errors.New("session cookie was empty"), http.ErrNoCookie)
//	}
//	return SessionId(sessionName), nil
//}
