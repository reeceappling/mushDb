package rfid

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"time"
)

const authServiceContextKey = "authenticationService"

// TODO: DONE ON CONTEXT WILL STOP CLEANUP?
func NewAuthService(sessionsCleanupFreq, sessionTTL *time.Duration) *AuthService {
	out := &AuthService{
		Map: sessions.NewMap[AuthInfo, string](utils.Default(sessionTTL, 2*time.Hour)),
	}
	// Start cleanup thread
	t := time.NewTicker(utils.Default(sessionsCleanupFreq, 5*time.Minute))
	go func() {
		select {
		case <-t.C:
			out.Map.ClearOldSessions()
		}
	}()
	return out
}

func GetAuthService(ctx context.Context) (*AuthService, error) {
	svc, ok := ctx.Value(authServiceContextKey).(*AuthService)
	if !ok {
		return nil, errors.New("auth service not found")
	}
	if svc == nil {
		return nil, errors.New("auth service is nil")
	}
	return svc, nil
}

type AuthService struct {
	Map sessions.Map[AuthInfo, string]
}

func (serv *AuthService) TryToReAuth(sessionKey string) (sessions.Session[AuthInfo], error) {
	if sessionKey == "" {
		return sessions.Session[AuthInfo]{}, ErrBlankSessionKey
	}
	res := serv.Map.GetSession(sessionKey)
	if res.Err != nil {
		return sessions.Session[AuthInfo]{}, utils.NotFound
	}
	return *res.Item, nil
}

type AuthInfo struct {
	Id   primitive.ObjectID // ID in database for User // TODO: OK?
	Opts UserPerms          //admin?canCreateAccounts.
}

type UserPerms utils.Set[string]

func (perms UserPerms) Contains(x string) bool {
	return utils.Set[string](perms).Contains(x)
}

func (perms UserPerms) ToSlice() []string {
	return utils.Set[string](perms).ToSlice()
}

func (perms UserPerms) OptsAsString() string {
	out := perms.ToSlice()
	if perms.Contains(MaxAuthKey) {
		out = []string{MaxAuthKey}
	}
	bs, _ := json.Marshal(out)
	return string(bs)
}

func (info AuthInfo) IsAdmin() bool {
	return info.Opts.Contains(MaxAuthKey)
}

var MaxAuthInfo = AuthInfo{
	Id:   primitive.NewObjectID(),
	Opts: UserPerms(utils.Set[string]{MaxAuthKey: {}}),
}

var (
	ErrBlankSessionKey = errors.New("blank session key")
)

// TODO: change master creds?
func SetupAuthenticatorOnContext(ctx context.Context, sessionsCleanupFreq, sessionTTL *time.Duration, masterUsername, masterPassword string) context.Context { // TODO: THIS!
	svc := NewAuthService(sessionsCleanupFreq, sessionTTL)
	//svc.Map.AddSession(masterUsername) // TODO: ADD MASTER USER TO DB?
	return context.WithValue(ctx, authServiceContextKey, svc)
}

const (
	AuthPermsContextHeaderKey = "Auth-Info"
	AuthSessionCookieKey      = "Session-Id"
)

func authSplitterMiddleware(masterUser, masterPass string) func(http.Handler, http.Handler) http.Handler {
	return func(onAuthed http.Handler, onAuthFail http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			svc, err := GetAuthService(ctx)
			if err != nil {
				http.Error(w, "NO ACTIVE AUTH SERVICE FOUND. ENDPOINT UNAVAILABLE: "+err.Error(), http.StatusInternalServerError)
			}
			cookie, err := GetSessionCookie(r)
			if err != nil {
				onAuthFail.ServeHTTP(w, r)
			}
			sessionId := cookie.Value
			sess, err := svc.TryToReAuth(sessionId)
			if err != nil {
				onAuthFail.ServeHTTP(w, r)
			}
			cookie.Expires = sess.Expiry
			http.SetCookie(w, cookie)
			ctxWithPerms := context.WithValue(r.Context(), AuthPermsContextHeaderKey, sess.Data)
			onAuthed.ServeHTTP(w, r.WithContext(ctxWithPerms))
		})
	}
}

func GetAuthInfo(ctx context.Context) AuthInfo { // TODO: USE!
	usr, ok := ctx.Value(AuthPermsContextHeaderKey).(AuthInfo)
	if !ok {
		return AuthInfo{
			Id:   primitive.ObjectID{},
			Opts: UserPerms(utils.Set[string]{}),
		}
	}
	return usr
}

// TODO: ALWAYS DO THIS FIRST!
func (serv *AuthService) necessaryFirstMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(serv.OnContext(r.Context())))
	})
}

func (serv *AuthService) AuthOrRedirectMiddleware(redirectUrl, masterUser, masterPass string) func(http.Handler) http.Handler {
	var redirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: PUT PAGE FROM IN THE DATA SO WE CAN PULL IT LATER?
		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
	})
	return func(nextHandler http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(nextHandler, redirectHandler))
	}
}

func (serv *AuthService) BasicSplitterMiddleware(masterUser, masterPass string) func(http.Handler, http.Handler) http.Handler {
	return func(authSuccess, authFailure http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(authSuccess, authFailure))
	}
}

func (serv *AuthService) AuthOrDenyMiddleware(masterUser, masterPass string) func(http.Handler) http.Handler {
	var denyHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	return func(nextHandler http.Handler) http.Handler {
		return serv.necessaryFirstMiddleware(authSplitterMiddleware(masterUser, masterPass)(nextHandler, denyHandler))
	}
}

func (serv *AuthService) OnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, authServiceContextKey, serv)
}

type loginItem struct { // TODO: FIX THIS FOR GOOGLE VS user/pass
	ID       string
	Password string
	// TODO: LOGIN! GOOGLE VS USER/PASS
}

func (serv *AuthService) LoginUserPassWithMasterUserPass(rootU, rootP string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unable to read body", http.StatusBadRequest)
			return
		}
		data := loginItem{}
		err = json.Unmarshal(bs, &data)
		if err != nil {
			http.Error(w, "invalid login struct", http.StatusBadRequest)
			return
		}
		var newSessionId = rand.Text() // TODO: GENERATE SESSION ID!
		var authinfo = AuthInfo{}
		if data.ID == rootU && data.ID == rootP {
			authinfo = MaxAuthInfo
		} else {
			usern := "" // TODO: GET USERNAME FROM REQUEST?
			passw := "" // TODO: GET PASSWORD FROM REQUEST?
			ctx := r.Context()
			userInfo := User{}
			err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(userCollName).FindOne(ctx, bson.M{"_id": usern}).Decode(&userInfo)
			if err != nil {
				http.Error(w, "no matching credentials found", http.StatusBadRequest)
				return
			}
			err = userInfo.validatePassword(passw)
			if err != nil {
				http.Error(w, "no matching credentials found.", http.StatusBadRequest) // TODO: add err here?
				return
			}
			authinfo = userInfo.perms()
		}
		sess, err := serv.Map.NewSession(newSessionId, authinfo) // TODO: ERR OK? MAYBE CREATE SESS ID HERE
		if err != nil {
			http.Error(w, "unable to create session", http.StatusInternalServerError)
			return
		}
		// Set session cookies and contexts
		ctx := context.WithValue(r.Context(), AuthPermsContextHeaderKey, sess.Data) // TODO ENSURE OK?
		setSessionCookie(w, r, newSessionId, sess)
		// TODO: Ensure proper redirection and that cookies stay
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

// TODO: use
func setSessionCookie(w http.ResponseWriter, r *http.Request, sessionId string, session sessions.Session[AuthInfo]) {
	http.SetCookie(w, &http.Cookie{
		Name:        AuthSessionCookieKey,
		Value:       sessionId,
		Quoted:      false,
		Path:        r.URL.Path,
		Domain:      r.URL.Host, // TODO: ok?
		Expires:     session.Expiry,
		RawExpires:  "",    // TODO: ????????????????
		MaxAge:      0,     // TODO: ????????????????
		Secure:      false, // TODO: ??????????????????????????????????
		HttpOnly:    false,
		SameSite:    http.SameSiteLaxMode, // TODO: ????????????????
		Partitioned: false,                // TODO: ????????????????
		Raw:         "",                   // TODO: ????????????????
		Unparsed:    nil,                  // TODO: ????????????????
	})
}

// TODO: THIS
func GetSessionCookie(r *http.Request) (*http.Cookie, error) {
	out, err := r.Cookie(AuthSessionCookieKey)
	if err != nil {
		if err != http.ErrNoCookie {
			return nil, err
		}
		return nil, nil
	}
	return out, nil
}
