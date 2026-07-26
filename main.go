package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	google2 "github.com/markbates/goth/providers/google"
	"github.com/reeceappling/goUtils/v2/logging"
	"github.com/reeceappling/goUtils/v2/utils"
	rfid "github.com/reeceappling/mushDb/api"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/shared"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	extProto   string
	extHost    string
	baseApiUrl string
)

func init() {
	extProto = os.Getenv("MAIN_API_EXTERNAL_PROTOCOL")
	extHost = os.Getenv("MAIN_API_EXTERNAL_HOST")
	baseApiUrl = fmt.Sprintf("%s://%s", extProto, extHost)
}

func setupDb(ctxIn context.Context) (ctx context.Context, client *mongo.Client, err error) {
	dbHostName := os.Getenv("DB_HOST_NAME")
	dbUser := os.Getenv("MONGO_INITDB_USERNAME")
	dbPass, err := getSecret("MONGO_INITDB_PASSWORD")
	if err != nil {
		panic("failed to resolve secret MONGO_INITDB_PASSWORD: " + err.Error())
	}
	dbSetupPass, err := getSecret("MONGO_INITDB_SETUP_PASSWORD")
	if err != nil {
		panic("failed to resolve secret MONGO_INITDB_SETUP_PASSWORD: " + err.Error())
	}
	dbSetupUser := os.Getenv("MONGO_INITDB_SETUP_USERNAME")
	dbHostPortStr := os.Getenv("DB_HOST_PORT")
	dbHostPort, err := strconv.Atoi(dbHostPortStr)
	if err != nil {
		println(errors.Join(errors.New("no db port configured on env var DB_HOST_PORT. Defaulting to 27017"), err).Error())
		dbHostPort = 27017
	}

	if dbUser == "" {
		err = errors.New("no MONGO_INITDB_USERNAME env var found")
		return
	}

	println("Connecting to database as initializer")
	ctxA, clientInitial, errA := rfid.NewMongoDbClient(ctxIn, dbSetupUser, dbSetupPass, dbHostName, dbHostPort)
	if errA != nil {
		return ctxA, nil, errors.Join(errors.New("failed to create MongoDB creation client"), errA)
	}
	println("Initializing DB")
	if err = rfid.Initialize(ctxA); err != nil {
		return ctxA, nil, errors.Join(errors.New("failed to initialize database"), err)
	}
	if err = clientInitial.Disconnect(ctxA); err != nil {
		return ctxA, nil, errors.Join(errors.New("failed to disconnect database on start"), err)
	}
	println("Connecting to database standard application user")
	ctx, client, err = rfid.NewMongoDbClient(ctxIn, dbUser, dbPass, dbHostName, dbHostPort)
	if err != nil {
		return ctx, nil, errors.Join(errors.New("failed to create MongoDB application client"), err)
	}
	println("DB setup and connection complete!")
	return ctx, rfid.GetMongoClient(ctx), nil
}

//var oauthConfig *oauth2.Config // TODO: ????

const defaultHttpPort = 8080
const loginPath = "/login"

func main() {
	ctx := context.Background()
	var envir string
	switch os.Getenv("ENVIRONMENT") { // TODO: set this!
	case "prod":
		envir = env.Prod
	case "cert":
		envir = env.Cert
	case "qual", "dev", "devl":
		envir = env.Dev
	default:
		panic("invalid environment: " + os.Getenv("ENVIRONMENT"))
	}
	ctx = env.SetEnv(ctx, envir)
	picsPath := os.Getenv("PICS_PATH")
	if picsPath == "" {
		panic("env var missing for PICS_PATH")
	}
	ctx = pics.SetFilePath(ctx, picsPath)
	if err := resolveGothicSessionSecret(); err != nil {
		panic("failed to resolve gothic session secret: " + err.Error())
	}
	var err error
	authSvc := rfid.NewAuthService(utils.Pointer(2*time.Minute), utils.Pointer(1*time.Hour))
	ctx = authSvc.OnContext(ctx)

	// adminEmail := os.Getenv("ADMIN_EMAIL") // TODO: use?

	// TODO: make sure logger is set up correctly
	log := logging.LoggerFactoryFor("mush-api-go") // TODO: ok name?
	ctx = logging.SetLogger(ctx, log)
	ctx = logging.SetSugaredLogger(ctx, log.Sugar())

	// Get non-db env vars
	rfidRegistrySecret, err := getSecret("RFID_SECRET")
	if err != nil {
		panic("failed to resolve secret RFID_SECRET: " + err.Error())
	}
	googId, err := getSecret("GOOGLE_CLIENT_ID")
	if err != nil {
		panic("failed to resolve secret GOOGLE_CLIENT_ID: " + err.Error())
	}
	googSecret, err := getSecret("GOOGLE_CLIENT_SECRET")
	if err != nil {
		panic("failed to resolve secret GOOGLE_CLIENT_SECRET: " + err.Error())
	}
	apiPort, err := strconv.Atoi(os.Getenv("API_PORT")) // TODO: ok?
	if err != nil {
		fmt.Printf(`No api port configured, defaulting to port %d`, defaultHttpPort)
		apiPort = defaultHttpPort
	}
	// TODO: api is hosted internally on 8080, but the actual site on cloudflare uses 443! Web is on 3000
	//tempPortStr := ""
	//if (extProto == "https" && apiPort != 443) || (extProto == "http" && (apiPort != 80 && apiPort != defaultHttpPort)) { // TODO: 80 AND 8080 ok here?
	//	tempPortStr = fmt.Sprintf(`:%d`, apiPort)
	//}
	//apiHostPort := extHost + tempPortStr
	//baseApiUrl = fmt.Sprintf("%s://%s", extProto, apiHostPort)
	//oauthConfig = &oauth2.Config{
	//	ClientID:     googId,
	//	ClientSecret: googSecret,
	//	RedirectURL:  extHost + "/auth/google/callback",
	//	Scopes:       []string{"email", "profile", "openid"},
	//	Endpoint:     google.Endpoint,
	//}

	ctx, client, err := setupDb(ctx)
	if err != nil {
		panic("Error setting up db: " + err.Error())
	}
	defer func() {
		err = client.Disconnect(ctx)
		if err != nil {
			panic("db failed to disconnect: " + err.Error()) // TODO: ok?
		}
	}()

	webHostName := envVarOrDefault("WEB_HOST_INTERNAL", "web") // Can have port if not hosting on 80

	// Set up server
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(apiPort), // TODO: will this work for websockets?
		ReadHeaderTimeout: 10 * time.Second,
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		switch v := <-sigterm; {
		case v == syscall.SIGTERM:
			println("Server received SIGTERM Signal")
		case v == syscall.SIGINT:
			println("Server received SIGINT Signal")
		default:
			println("Server Received Signal", "signal", v)
		}
		println("Shutting Down Server")

		srvCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // default ecs shutdown is 30 seconds
		defer cancel()

		if err = srv.Shutdown(srvCtx); err != nil {
			println("server shutdown with unhandled error: " + err.Error())
		}

		if err = srv.Close(); err != nil {
			println("server closed with unhandled error: " + err.Error())
			return
		}
		os.Exit(42) // weird code so we can tell at a glance that we shut it down
	}()
	googleAuthProvider := google2.New(googId, googSecret, authCallbackUrlFor("google"), "email", "profile", "openid") // TODO: ensure callback is ok
	guestCallbackUrl := authCallbackUrlFor("guest")                                                                   // TODO: set this up!
	guestAuthProvider := NewGuestProvider(guestCallbackUrl, "guest")
	goth.UseProviders(googleAuthProvider, guestAuthProvider)
	//gothic.Store = customSessionStore{} // TODO: REENABLE????

	http.HandleFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, errr := w.Write([]byte("Hello World!"))
		if errr != nil {
			println("/test failure: " + errr.Error())
		}
	}))

	//providersMap := map[string]string{
	//	"google": "google",
	//}
	//var keys []string
	//for k := range providersMap {
	//	keys = append(keys, k)
	//}
	////sort.Strings(keys)
	////providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: providersMap}
	println("Creating Middlewares")
	// TODO: NEXT.JS LOGS! WEBPACK IS GETTING TRAPPED BY MAIN SERVER! WEBSERVER ENV VARS!
	// Setup middlewares

	cleanupFreq := 2 * time.Minute
	// TODO: rfid sessions mock readers?
	mgr := websocketSessions.NewSessionManager(&cleanupFreq, rfidRegistrySecret)
	defer mgr.Cleanup()
	// Start generating mainCollectionIds
	rfid.StartGeneratingMCIDs(ctx, 12)
	ctx, rateLimiter, rfidMiddleware, internalOnlyMiddleware, webAuthMiddleware, authOrDenyMiddleware, ctxMiddleware, err := setupMiddlewares(ctx, mgr, loginPath, apiPort)
	if err != nil {
		panic("Error setting up middleware: " + err.Error())
	}

	// RFID HANDLERS
	println("Defining endpoints")
	println("Defining RFID endpoints")
	// Must be publicly available. (external)
	// TODO: ADD OPTIONS method to all endpoints!
	http.HandleFunc("/rfid/ws", Middlewares(ctxMiddleware, rfidMiddleware)(http.HandlerFunc(websocketSessions.ServerHandler)).ServeHTTP) // TODO: was /rfid/ws // TODO: OPTIONS
	// Must be internal to docker network
	http.HandleFunc("/rfid/read/{readerName}", Middlewares(ctxMiddleware, rfidMiddleware, authOrDenyMiddleware)(rfidReadHandler).ServeHTTP)        // TODO: OPTIONS // TODO: internal only?   //  OUTPUT IS BASE 2! // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/write/{writerName}", Middlewares(ctxMiddleware, rfidMiddleware, authOrDenyMiddleware)(rfidWriteHandler).ServeHTTP)      // TODO: OPTIONS // INPUT IS BASE58. OUTPUT IS BASE 2! // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/readers", Middlewares(ctxMiddleware, rfidMiddleware, internalOnlyMiddleware)(getRfidReaderNamesHandler).ServeHTTP)      // TODO: OPTIONS // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/clear/{writerName}", Middlewares(ctxMiddleware, rfidMiddleware, internalOnlyMiddleware)(clearRfidTagHandler).ServeHTTP) // TODO: OPTIONS // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)

	// SERVER HANDLERS! (PASSTHROUGH) view, new, import
	webHostPort := 3000
	webHostPortStr := os.Getenv("WEB_HOST_INTERNAL_PORT") // only for running outside of docker compose
	if webHostPortStr != "" {
		webHostPort, err = strconv.Atoi(webHostPortStr)
		if err != nil {
			panic("invalid internal web host port: " + webHostPortStr)
		}
	} else {
		println("No/invalid web host port specified, defaulting to 3000")
	}
	passthroughConfig := newPassthroughHandlerConfig().
		useHttps(false).
		withHost(webHostName).
		withCookies(true).
		withHeaders(true).
		withPort(webHostPort)

	// Proxy combo middlewares
	webAuthAdminMiddleware := rfid.GetAuthService(ctx).AuthAdminOrDenyMiddleware
	webProxyHandler := newPassthroughHandler(passthroughConfig)
	unAuthedProxied := ctxMiddleware(webProxyHandler)
	authedProxied := ctxMiddleware(webAuthMiddleware(webProxyHandler))
	adminProxied := ctxMiddleware(webAuthAdminMiddleware(webProxyHandler))

	println("Defining webserver passthrough endpoints (external)")
	OptionsGetPost := OptionsMiddleware(http.MethodGet, http.MethodPost)
	OptionsGetOnly := OptionsMiddleware(http.MethodGet)
	OptionsPostOnly := OptionsMiddleware(http.MethodPost)

	// handle login

	rateLimitCtxMiddleware := func(next http.Handler) http.Handler {
		return rateLimiter(ctxMiddleware(next))
	}
	corsAuthRateLimit := Middlewares(CorsAuthMiddleware, rateLimitCtxMiddleware)
	http.Handle(loginPath /* /login */, Middlewares(OptionsGetPost, corsAuthRateLimit, handleLoginMiddleware)(webProxyHandler))
	http.Handle("/logout", Middlewares(OptionsGetOnly, corsAuthRateLimit)(handleLogout)) // TODO: make logout button in ts!
	http.Handle("/guestLogin", Middlewares(OptionsPostOnly, rateLimitCtxMiddleware)(handleGuestLogin))
	// TODO: reenable for testing only! http.Handle("/testLogin/{emailEncoded}", rateLimitCtxMiddleware(OptionsPostOnly(handleTestLogin))) // TODO: remove later
	http.Handle("/auth/{provider}", Middlewares(OptionsGetOnly, corsAuthRateLimit)(authProviderHandler)) // TODO: getOnly ok?
	http.Handle("/auth/{provider}/callback", Middlewares(corsAuthRateLimit)(authCallbackHandler))        // TODO: GET or POST?
	// Biometrics endpoints TODO: (maybe add biometric provider?)
	//http.Handle("/biometrics/fingerprint/register-start", CorsAuthMiddleware(rateLimitCtxMiddleware(http.HandlerFunc(BeginFingerprintRegistration))))   // TODO: change middlewares and handler
	//http.Handle("/biometrics/fingerprint/register-end", CorsAuthMiddleware(rateLimitCtxMiddleware(http.HandlerFunc(FinishFingerprintRegistration))))    // TODO: change middlewares and handler
	//http.Handle("/biometrics/fingerprint/register-challenge", CorsAuthMiddleware(rateLimitCtxMiddleware(biometricFingerprintRegisterChallengeHandler))) // TODO: change middlewares and handler
	//http.Handle("/biometrics/fingerprint/verify-challenge", CorsAuthMiddleware(rateLimitCtxMiddleware(biometricFingerprintVerifyChallengeHandler)))     // TODO: change middlewares and handler

	// Proxied to react
	// Generalized react endpoints
	http.Handle("/_next", unAuthedProxied)
	http.Handle("/", unAuthedProxied)

	// Specific React/Next pages
	http.Handle("/import/{variant}", authedProxied)         // GET import is here (import item page)
	http.Handle("/new/{variant}", authedProxied)            // GET new item is here (new item page)
	http.Handle("/view/{variant}/{entryId}", authedProxied) // GET view item is here (view item page)
	http.Handle("/list/{variant}", authedProxied)           // GET list is here (list items page)
	// Error and test pages
	http.Handle("/error/{errTxt}", webProxyHandler) // TODO: rate limit???? ctx middleware? auth middleware?
	http.Handle("/testpage", webProxyHandler)       // GET testpage is here (test page)       // TODO: REMOVE
	// Admin pages
	http.Handle("/whitelistUser", adminProxied) // TODO: THIS!

	println("Defining sensor data endpoints")
	// TODO: this
	// Sensor data endpoints
	//http.Handle("/sensorData/{nodeName}", rfid.GetSensorDataHandler())           // TODO: middleware?
	//http.Handle("/sensorDataSince/{nodeName}", rfid.GetSensorDataSinceHandler()) // TODO: middleware?
	//http.Handle("/addSensorData/{nodeName}", rfid.AddSensorDataHandler())        // TODO: middleware?

	println("Defining admin endpoints")
	http.Handle("/admin/whitelistUser", Middlewares(ctxMiddleware, webAuthAdminMiddleware)(whitelistUserHandler))

	// TODO: ADMIN STUFF
	// TODO: user-viewer/editor for admin
	// TODO: need to be able to create new users

	println("Defining db interaction endpoints")
	// TODO: CORS db middlewares?
	// Resolving Types
	http.Handle("/db/pathFor/{id}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(rfid.GetPageForIdHandler)) // TODO: DenyGuestMiddleware?
	// Get handlers
	// TODO: ??? http.Handle("/db/get/rfid/{id}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(getRfidHandler()) // TODO: DenyGuestMiddleware?             // TODO: GET RID OF???             // TODO: ensure this works for base58s
	// TODO: ??? http.Handle("/db/get/rfid/{id}", rateLimitedWithCtxAndInternalAuth(getRfidHandler()) // TODO: DenyGuestMiddleware?             // TODO: GET RID OF???             // TODO: ensure this works for base58s
	http.Handle("/db/get/{variant}/{id}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(getAnyCollectionHandler))
	http.Handle("/db/images/{imageSubPath...}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(getImageHandler)) // TODO: rate limiter ok here?
	// Creation handlers
	http.Handle("/db/create/{variant}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware, rfidMiddleware, rfid.DenyGuestMiddleware)(rfid.CreateHandler))
	// TODO: chain spore print handler?
	// update handlers
	http.Handle("/db/update/{variant}/{id}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware, rfid.DenyGuestMiddleware)(rfid.UpdateHandler)) // TODO: no rfid?
	// import handlers
	http.Handle("/db/import/{variant}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware, rfidMiddleware, rfid.DenyGuestMiddleware)(rfid.ImportHandler))

	// delete handlers
	// TODO: enable! http.Handle("/db/delete/{endpt}/{id}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware, rfidMiddleware, rfid.AdminOnlyMiddleware)(rfid.DeleteHandler))
	// List handlers
	http.Handle("/db/list/{variant}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(rfid.ListEntriesHandler))
	http.Handle("/subspeciesFor/{variant}", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware)(rfid.ListSubspeciesHandler))
	http.Handle("/sessionUserProjects", Middlewares(rateLimiter, ctxMiddleware, authOrDenyMiddleware, rfid.DenyGuestMiddleware)(rfid.SessionUserProjectsHandler)) // TODO: DenyGuestMiddleware? Will guests only have public projects???
	// Next endpt needs no authorization, but does have a rate limiter?? // TODO: rl?
	http.Handle("/options/{optionsType}", Middlewares(rateLimiter, internalOnlyMiddleware)(rfid.GetOptionsHandler)) // TODO: DenyGuestMiddleware? Guests should not be changing anything...

	if err = srv.ListenAndServe(); err != nil {
		panic("failed to listen and serve for http: " + err.Error())
	}
	if err != nil {
		panic("ERROR CLOSING SERVER " + err.Error())
	}
}

func getSecret(secretName string) (string, error) {
	path := filepath.Join("/run/secrets", secretName)
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	//// Trim trailing newlines or whitespace added by file editors // TODO: ????
	//return string(bytes.TrimSpace(payload)), nil
	return string(payload), nil
}

func resolveGothicSessionSecret() error {
	sessionSecretName := "SESSION_SECRET"
	sec, err := getSecret(sessionSecretName) // TODO: unnecessary, sessionsecret is pulled in during init of gothic
	if err != nil {
		return err
	}
	println("Gothic secret used: ", sec)
	return os.Setenv(sessionSecretName, sec)
}

func OptionsMiddleware(acceptableMethods ...string) SingleMiddleware {
	methods := strings.Join(acceptableMethods, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.Header().Set("Allow", methods)
				w.Header().Set("Access-Control-Allow-Methods", methods) // TODO: only if CORS enabled?
				w.WriteHeader(http.StatusOK)                            // TODO: ok?
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type customSessionStore struct { // TODO: use or delete
	svc *rfid.AuthService
}

func (c customSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	for _, cookie := range r.Cookies() {
		println(cookie.Name)
		if cookie.Name == name {
			// TODO: FIX THIS IF NEEDED
		}
	}
	return nil, errors.New("implement me") // TODO: fix?
}

func (c customSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	//TODO implement me
	panic("implement me")
	return nil, errors.New("implement me") // TODO: fix?
}

func (c customSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	//TODO implement me
	panic("implement me") // TODO: fix?
	return errors.New("implement me")
}

func internalOnlyMiddlewareCreator(validDomain string, expectedPort int) func(handler http.Handler) http.Handler {
	expDomain := fmt.Sprintf("%s:%d", validDomain, expectedPort)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqHost := r.Host
			if len(strings.Split(r.Host, ":")) == 1 {
				if r.TLS != nil {
					reqHost = r.Host + ":443"
				} else {
					reqHost = fmt.Sprintf(`%s:%d`, r.Host, defaultHttpPort)
				}
			}
			if reqHost != expDomain {
				http.Error(w, "Internal requests only expected at this endpoint. Invalid host or port", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ReqTrackingMiddleWare(handler http.Handler) http.Handler { // TODO: USE OR DELETE
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logging.GetLogger(ctx)
		requestPath := r.URL.Path
		var requestId string
		// Set request path on request

		ctx = request.SetPath(ctx, requestPath)
		log = log.WithStringKVP(request.Path, requestPath)
		// Set request and trace ids (trace only if span exists)
		log, requestId = log.Fork(request.Id)
		if span, ok := tracer.SpanFromContext(ctx); ok {
			span.SetTag(request.Id, requestId)
			span.SetBaggageItem(request.Id, requestId)
			traceId := span.Context().TraceID()
			log = log.WithTraceId(ctx, strconv.Itoa(int(traceId)))
		}

		defer log.Sync() //nolint:errcheck

		handler.ServeHTTP(w, r.WithContext(logging.SetLogger(ctx, log)))

		//now := time.Now()
		//// TODO: ensure works properly
		//wrappedWriter := NewResponseWriterWrapper(w, func() {
		//	w.Header().Set("reqTimeMs", strconv.Itoa(int(time.Since(now).Milliseconds())))
		//}) // reqTimeMs header will exist on responses
		//handler.ServeHTTP(wrappedWriter, r.WithContext(logging.SetLogger(ctx, log)))
	})
}

//
//func NewResponseWriterWrapper(w http.ResponseWriter, onRespond func()) WrappedResponseWriter {
//	return WrappedResponseWriter{
//		TwoHundredOrAboveHeaderWritten: false,
//		internal:                       w,
//		onRespond:                      onRespond,
//	}
//}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

//type WrappedResponseWriter struct {
//	TwoHundredOrAboveHeaderWritten bool
//	internal                       http.ResponseWriter
//	onRespond                      func()
//}
//
//func (w WrappedResponseWriter) Header() http.Header {
//	return w.internal.Header()
//}
//func (w WrappedResponseWriter) Write(bs []byte) (int, error) {
//	// Will call WriteHeader(200)before bytes are written
//	if !w.TwoHundredOrAboveHeaderWritten {
//		w.onRespond()
//		w.TwoHundredOrAboveHeaderWritten = true
//	}
//
//	return w.internal.Write(bs)
//}
//func (w WrappedResponseWriter) WriteHeader(statusCode int) {
//	if w.TwoHundredOrAboveHeaderWritten {
//		return
//	}
//	if statusCode >= 200 {
//		w.onRespond()
//		w.TwoHundredOrAboveHeaderWritten = true
//	}
//	w.internal.WriteHeader(statusCode) // TODO: handle 1XX/2XX/???
//}

func CorsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: fix?
		next.ServeHTTP(w, r)
	})
}

func setupMiddlewares(ctxIn context.Context, mgr *websocketSessions.SessionManager, loginPath string, apiPort int) (
	ctx context.Context,
	rateLimiter, rfidMiddleware, internalOnlyMiddleware, webAuthMiddleware, internalAuthMiddleware func(http.Handler) http.Handler,
	ctxMiddleware func(http.Handler) http.Handler,
	err error) {
	// PicsPath and rfid middleware
	// Pics path middleware
	picsPath := os.Getenv("PICS_PATH")
	if picsPath == "" {
		panic("env var missing for PICS_PATH")
	}
	ctx = pics.SetFilePath(ctxIn, picsPath)
	ctxMiddleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx)) // TODO: ok?
		})
	}
	rfidMiddleware = mgr.Middleware()

	// Rate limiter
	//rateLimitCount, rateLimitPeriod := int64(150), 5*time.Minute // TODO: reenable for prod
	rateLimitCount, rateLimitPeriod := int64(150), time.Minute
	// Rate limiter
	rate := limiter.Rate{ // TODO: TURN
		Period: rateLimitPeriod,
		Limit:  rateLimitCount,
	}
	rateLimiterStorage := memory.NewStore()                                                              // TODO: ok?
	rateLimiterUnderlying := limiter.New(rateLimiterStorage, rate, limiter.WithTrustForwardHeader(true)) // TODO: trust header?
	rateLimiter = stdlib.NewMiddleware(rateLimiterUnderlying).Handler

	// Direction and auth middleware
	// Direction middleware
	internalOnlyMiddleware = internalOnlyMiddlewareCreator("api", apiPort)
	// Auth Middleware
	svc := rfid.GetAuthService(ctx)
	webAuthMiddleware = svc.AuthOrRedirectMiddleware(loginPath)
	internalAuthMiddleware = svc.AuthOrDenyMiddleware

	return
}

func envVarOrDefault(varName, defaultResult string) string {
	result := os.Getenv(varName)
	if result == "" {
		println("env var " + varName + " missing, defaulting to " + defaultResult)
		return defaultResult
	}
	return result
}

var handleGuestLogin http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	println("signing in as guest user...") // TODO: del!
	sessId, err := rfid.GetAuthService(ctx).SigninGuestUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("created new session id for guest: " + sessId)
	session, err := gothic.Store.New(r, string(sessId))
	if err != nil {
		// TODO: delete guest session?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("adding session id to session") // TODO: del
	err = gothic.StoreInSession(rfid.SessionIdKey, string(sessId), r, w)
	if err != nil {
		// TODO: delete guest session? If not already saved, then no?
		//session.Options.MaxAge = -1 // delete session
		err = errors.Join(errors.New("sessId storage fail"), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = gothic.Store.Save(r, w, session); err != nil {
		http.Error(w, "failed so save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	redirectToBasePage(r, w)
}
var handleTestLogin http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email, err := rfid.UrlDecodeString(r.PathValue("emailEncoded"))
	if err != nil {
		http.Error(w, "failed to decode email from request path: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessId, err := rfid.GetAuthService(ctx).SigninTestUser(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("created new session id (" + string(sessId) + ") for test user: " + email)
	session, err := gothic.Store.New(r, string(sessId))
	if err != nil {
		// TODO: delete test session?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = gothic.StoreInSession(rfid.SessionIdKey, string(sessId), r, w)
	if err != nil {
		// TODO: delete test session? If not already saved, then no?
		//session.Options.MaxAge = -1 // delete session
		err = errors.Join(errors.New("sessId storage fail"), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = gothic.Store.Save(r, w, session); err != nil {
		http.Error(w, "failed so save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	redirectToBasePage(r, w)
}

func handleLoginMiddleware(viewHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			viewHandler.ServeHTTP(w, r)
		case http.MethodPost:
			time.Sleep(1500 * time.Millisecond) // Make user wait for login, lower likelihood of attack
			println("SHOULD NEVER BE HIT")
		default:
			http.Error(w, "Unsupported http request method: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

	})
}

func handleUserAuthedViaGoth(ctx context.Context, user goth.User, w http.ResponseWriter, r *http.Request) error {
	authSvc := rfid.GetAuthService(ctx)
	provider := r.PathValue("provider")
	var sessId rfid.SessionId
	var err error
	switch provider {
	case "google":
		sessId, _, err = authSvc.SigninGoogleAuthedUser(ctx, user) // TODO: ???
		if err != nil {
			return errors.Join(errors.New("authSvc google signin fail"), err)
		}
	default:
		sessId, err = authSvc.SigninGuestUser() // TODO: ???
		if err != nil {
			return errors.Join(errors.New("authSvc guest signin fail"), err)
		}
	}
	// TODO: WHAT ABOUT GUESTS?

	err = gothic.StoreInSession(rfid.SessionIdKey, string(sessId), r, w)
	if err != nil {
		return errors.Join(errors.New("sessId storage fail"), err)
	}
	return nil
}

//func getSessionValue(session *sessions.Session, key string) (string, error) {
//	value := session.Values[key]
//	if value == nil {
//		return "", fmt.Errorf("could not find a matching session for this request")
//	}
//
//	rdata := strings.NewReader(value.(string))
//	r, err := gzip.NewReader(rdata)
//	if err != nil {
//		return "", err
//	}
//	s, err := io.ReadAll(r)
//	if err != nil {
//		return "", err
//	}
//
//	return string(s), nil
//}
//
//func adminLogin(w http.ResponseWriter, r *http.Request) {
//	// TODO: implement!
//}

func getDestinationFromAuthRequest(r *http.Request) string {
	if dst := r.URL.Query().Get("destination"); dst != "" {
		return dst
	}
	return "/?noDst=true"
}

var authProviderHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	println("RECEIVED REQUEST at " + r.URL.String()) // TODO: del?
	if providerName == "guest" {
		handleGuestLogin(w, r)
		return
	}

	//bs, err := io.ReadAll(r.Body)
	//if err != nil {
	//	println("failed to read body", err.Error())
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//println(string(bs)) // TODO: del?
	q := r.URL.Query()
	q.Add("provider", "google")
	customParamsFromMap := func(m map[string]string) string {
		sb := strings.Builder{}
		ct := 0
		for k, v := range m {
			if ct != 0 {
				sb.Write([]byte("|"))
			}
			sb.Write([]byte(fmt.Sprintf("%s=%s", k, v)))
			ct++
		}
		return sb.String()
	}
	customParams := customParamsFromMap(map[string]string{
		"destination": getDestinationFromAuthRequest(r),
		"provider":    "google",
	})
	q.Add("state", customParams) // Combine
	r.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(w, r)
	if err == nil {

		if err = handleUserAuthedViaGoth(r.Context(), user, w, r); err != nil {
			println("failed to handle user auth: " + err.Error()) // TODO: del?
			http.Error(w, "failed to handle user auth: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("OK!"))
		return
		// t, _ := template.New("foo").Parse(userTemplate)
		//t.Execute(w, gothUser)
	}
	println("not authed, beginning authentication by redirecting") // TODO: del?

	//gothic.BeginAuthHandler(w, r)
	url, err := gothic.GetAuthURL(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "error in gothic.GetAuthURL: ", err) // TODO: handle err?
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect) // TODO: permanent redirect??
}

//type SessionCookieStore struct {
//	internal *rfid.AuthService
//	sessions.Store
//}
//
//const sessionCookieName = "SessionId"

//func (store SessionCookieStore) Get(r *http.Request, name string) (*sessions.Session, error){
//	c, err := r.Cookie(sessionCookieName)
//	if err != nil {
//		return nil, err
//	}
//	res := store.internal.GetSession(rfid.SessionId(c.Value), true)
//	if res.Err != nil {
//		return nil, res.Err
//	}
//	sess := sessions.NewSession(store, c.Value)
//	return &sessions.Session{
//		ID:      c.Value,
//	}, nil
//}
//
//// New should create and return a new session.
////
//// Note that New should never return a nil session, even in the case of
//// an error if using the Registry infrastructure to cache the session.
//func (store SessionCookieStore) New(r *http.Request, name string) (*sessions.Session, error){
//	c, err := r.Cookie(sessionCookieName)
//	if err != nil {
//		return nil, err
//	}
//	store.internal.SessionForEmail()
//	res := store.internal.GetSession(rfid.SessionId(c.Value), true)
//	if res.Err != nil {
//		return nil, res.Err
//	}
//	return &sessions.Session{
//		ID:      c.Value,
//	}, nil
//}
//
//// Save should persist session to the underlying store implementation.
//func (store SessionCookieStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
//	// TODO: this
//}

func addDestinationToUrl(basePath string, destination string) string {
	if destination == "" {
		return basePath
	}
	q := url.Values{}
	q.Set("destination", destination)
	return basePath + "?" + q.Encode()

}

type stateMap map[string]string

func (sm stateMap) updateRequest(r *http.Request) error {
	if r == nil {
		return errors.New("request cannot be nil")
	}
	if r.URL == nil {
		return errors.New("url cannot be nil")
	}
	// TODO: update url?
	// Update query params based on state param returned from google
	updatedQuery := r.URL.Query()
	for stateKey, queryKey := range map[string]string{
		"destination": "destination",
		"provider":    "provider",
	} {
		if stateVal, exists := sm[stateKey]; exists {
			updatedQuery.Set(queryKey, stateVal)
		}
	}
	r.URL.RawQuery = updatedQuery.Encode()
	// TODO: update body?
	// TODO: update method?
	// TODO: update host?
	// TODO: update headers?
	// TODO: update cookies?
	return nil
}

func getStateMap(r *http.Request) (stateMap, error) {
	returnedState := r.URL.Query().Get("state")
	// Split out state and custom parameters
	parts := strings.Split(returnedState, "|")
	sm := make(map[string]string, len(parts))
	for _, part := range parts {
		keyval := strings.Split(part, "=") // TODO: ensure no equals in strings otherwise!
		if len(keyval) != 2 {
			println("malformed state for part: " + part)
			continue
		}
		sm[keyval[0]] = keyval[1]
	}
	return sm, nil
}

func getStateMapAndUpdateRequest(r *http.Request) error {
	sm, err := getStateMap(r)
	if err != nil {
		return err
	}
	return sm.updateRequest(r)

}

func removeCookie(c *http.Cookie, w http.ResponseWriter) {
	c.MaxAge = -1 // TODO: is this ok?
	http.SetCookie(w, c)
}

var authCallbackHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	println("hit callback. request was " + r.Method)
	// check for state params (from google) and update the request (params, etc) as needed
	err := getStateMapAndUpdateRequest(r)
	if err != nil {
		http.Error(w, "failed to get or utilize stateMap returned to auth callback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clear cookies if needed
	provider := r.PathValue("provider")
	callbackPath := addDestinationToUrl(fmt.Sprintf(`/auth/%s/callback`, provider), r.URL.Query().Get("destination"))
	// If too many cookies, remove the cookies
	sessCookies := r.CookiesNamed("_gothic_session")
	c, err := r.Cookie("_gothic_session")
	if err != nil {
		println("no gothic session cookie")
	} else {
		//println("session cookie", c.Name, c.Value)
	}
	if len(sessCookies) > 1 { // TODO: or >0? should probably be 0
		if err != nil {
			println("not finding cookie should never happen here!")
			http.Error(w, "not finding cookie should never happen here!", http.StatusInternalServerError)
			return
		}
		removeCookie(c, w)
		http.Redirect(w, r, callbackPath, http.StatusTemporaryRedirect)
		return
	}

	// TODO: ensure goth storage expirations match our storage expirations
	// TODO: OVERHAUL THE COOKIE STORAGE???
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		println("failed to auth in callback", err.Error()) // TODO: del
		if strings.Contains(err.Error(), "state token mismatch") {
			err = gothic.Logout(w, r)
			if err != nil {
				println("failed to logout: " + err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, callbackPath, http.StatusTemporaryRedirect)
			return
		}
	}

	// Loop back on initial failure
	switch provider {
	case "google":
		err = handleUserAuthedViaGoth(ctx, user, w, r)
		if err != nil {
			http.Error(w, "failed in handleUserAuthedViaGoth: "+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		// TODO: ?????
	}

	_, err = r.Cookie("_gothic_session")
	if err != nil {
		println("no cookie after auth", err.Error()) // TODO: ensure ok
		//http.Error(w, "no cookie after auth", http.StatusInternalServerError) // TODO: fix?
		//return
	}
	//http.SetCookie(w, c) // TODO: if this works do it everywhere? May not be needed!
	redirectToBasePage(r, w)
}

func redirectToBasePage(r *http.Request, w http.ResponseWriter) {
	ctx := r.Context()
	dst := r.URL.Query().Get("destination")
	if dst == "" {
		env.LogIfDev(ctx, "dst not on query")
		dst = baseApiUrl
	}
	env.LogIfDev(ctx, "redirecting to: "+dst)
	http.Redirect(w, r, dst, http.StatusTemporaryRedirect) // TODO: REDIRECT!
}

var handleLogout = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	sessId, err := rfid.SessionIdFromRequest(r)
	if err != nil {
		http.Error(w, "failed to get session id from request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err = gothic.Logout(w, r); err != nil {
		http.Error(w, "logout goth fail: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = rfid.GetAuthService(r.Context()).LogoutSession(sessId) // TODO: should this go before the goth stuff?
	if err != nil {
		http.Error(w, "failed to log out of rfid session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", loginPath) // TODO: set to login!
	w.WriteHeader(http.StatusTemporaryRedirect)
})

type passthroughHandlerConfig struct {
	redirectWithHttps bool
	host              string
	port              int
	copyCookies       bool
	copyHeaders       bool
}

func newPassthroughHandlerConfig() passthroughHandlerConfig {
	return passthroughHandlerConfig{
		redirectWithHttps: false,
		host:              "localhost",
		port:              -1,
		copyHeaders:       true,
		copyCookies:       true,
	}
}

func (conf passthroughHandlerConfig) withHost(hostName string) passthroughHandlerConfig {
	conf.host = hostName
	return conf
}
func (conf passthroughHandlerConfig) withPort(port int) passthroughHandlerConfig {
	conf.port = port
	return conf
}
func (conf passthroughHandlerConfig) useHttps(useHttps bool) passthroughHandlerConfig {
	conf.redirectWithHttps = useHttps
	return conf
}
func (conf passthroughHandlerConfig) withCookies(retainCookies bool) passthroughHandlerConfig {
	conf.copyCookies = retainCookies
	return conf
}
func (conf passthroughHandlerConfig) withHeaders(retainHeaders bool) passthroughHandlerConfig {
	conf.copyHeaders = retainHeaders
	return conf
}

func newPassthroughHandler(config passthroughHandlerConfig) http.HandlerFunc {
	protocol := map[bool]string{true: "https", false: "http"}[config.redirectWithHttps]
	finalHost := config.host
	if config.port > 0 {
		finalHost = fmt.Sprintf("%s:%d", config.host, config.port)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()
		defer r.Body.Close()
		weburl := protocol + "://" + finalHost + r.URL.Path
		println("passing through to " + weburl)
		req, err := http.NewRequest(r.Method, weburl, r.Body)
		if err != nil {
			http.Error(w, "Failed to create req: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
		//for key, vals := range r.Header {
		//	req.Header.Set(key, vals[0])
		//}
		// Set session header if exists

		//perms, err := rfid.GetAuthInfo(ctx)
		//if err == nil && perms.Opts != nil { // TODO: THIS IS EW
		//	optsBytes, err := json.Marshal(perms.Opts)
		//	if err != nil {
		//		// TODO: something here?
		//	}
		//	req.Header.Set(rfid.AuthPermsContextHeaderKey, string(optsBytes))
		//}
		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			println("passthrough failure, " + err.Error())
			http.Error(w, "Failed to do "+err.Error(), http.StatusInternalServerError)
			return
		}
		out, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := "Failed to read from http " + err.Error()
			println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			println(fmt.Sprintf("passthrough failure, status: %s, status code: %d", resp.Status, resp.StatusCode))
		}

		// RELAY HEADERS AS NEEDED
		// TODO: REMOVE AUTH PERMS HEADERS
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}

		//http.SetCookie(w, &http.Cookie{
		//	Username:        "ApplingSession",
		//	Value:       "",
		//	Quoted:      false,
		//	Path:        "",
		//	Domain:      "",
		//	Expires:     time.Time{},
		//	RawExpires:  "",
		//	MaxAge:      0,
		//	Secure:      false,
		//	HttpOnly:    false,
		//	SameSite:    0,
		//	Partitioned: false,
		//	Raw:         "",
		//	Unparsed:    nil,
		//})
		// TODO: THIS w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)

		if _, err = w.Write(out); err != nil {
			println("failed to write response")
		}
	}
}

//func placeholderMiddleware(next http.Handler) http.Handler { // TODO: DELETEME!!!!!!!
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		next.ServeHTTP(w, r)
//		// TODO: reenable
//		//next.ServeHTTP(w, r.WithContext(rfid.SetAuthInfo(r.Context(), rfid.AuthInfo{
//		//	Email: rfid.AlternateCollectionId(primitive.ObjectID{}),
//		//	Opts: &rfid.UserPermsResolved{
//		//		Admin:    utils.Pointer(true),
//		//		Projects: nil,
//		//	},
//		//})))
//	})
//}

func setupFilePathMiddleware(filePath string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(pics.SetFilePath(r.Context(), filePath)))
		})
	}
}

const imagesEndpoint = "/db/images/" // MUST match PicsEndpoint in PicWithNotes.tsx // TODO: use?
var getImageHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	// TODO: make sure authed?
	ctx := r.Context()
	imgSubPath := r.PathValue("imageSubPath")
	if imgSubPath == "" {
		http.Error(w, "image name must not be blank", http.StatusBadRequest)
		return
	}
	//println("trying to read picture file: " + filepath.Join(pics.GetFilePath(ctx), imgSubPath)) // TODO: DEL
	bytes, err := os.ReadFile(filepath.Join(pics.GetFilePath(ctx), imgSubPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//println("file does not exist!") // TODO: fix
			http.Error(w, "image not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving image. "+err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(bytes)
	if err != nil {
		rfid.HandleHttpWriteError(err)
	}
}

//var rootHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	_, err := w.Write([]byte("an apple a day: from " + r.URL.Path)) //nolint:errcheck // TODO: DO WE EVEN WANT THE ROOT TO RESPOND?
//	if err != nil {
//		rfid.HandleHttpWriteError(err)
//	}
//})

//// TODO: use this somewhere!
//// getItemTypeForId request body is just a base58 string of the mainCollectionId
//func getItemTypeForId() http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		defer r.Body.Close()
//		bs, err := io.ReadAll(r.Body)
//		if err != nil {
//			http.Error(w, "Error reading body: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		req := rfid.MainCollectionId{}
//		if err = json.Unmarshal(bs, &req); err != nil {
//			http.Error(w, "Error parsing body: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		itemType, err := rfid.FindItemTypeForId(r.Context(), req)
//		if err != nil {
//			http.Error(w, "Error finding item type: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		_, err = w.Write([]byte(itemType.EntryType()))
//		if err != nil {
//			rfid.HandleHttpWriteError(err)
//		}
//
//	})
//}
//
//func getRfidHandler() http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		ctx := r.Context()
//		id := r.PathValue("id")
//		idBytes := []byte(id)
//		if len(idBytes) != rfid.RfidByteSize {
//			http.Error(w, "invalid id format. Must be 8 bytes", http.StatusBadRequest)
//			return
//		}
//		item, err := rfid.GetMainCollectionItemWithId(ctx, [rfid.RfidByteSize]byte(idBytes))
//		if err != nil {
//			if errors.Is(err, mongo.ErrNoDocuments) {
//				http.Error(w, "not found: "+err.Error(), http.StatusNotFound)
//				return
//			}
//			http.Error(w, "failed to retrieve item: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		user, err := rfid.GetAuthInfo(ctx)
//		if err != nil {
//			http.Error(w, "failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		// Validate user can read this entry
//		if item.Permissions().HighestPermFor(user) == nil {
//			http.Error(w, "permission denied", http.StatusForbidden)
//			return
//		}
//		out, err := json.Marshal(item)
//		if err != nil {
//			http.Error(w, "failed to marshal item", http.StatusInternalServerError)
//			return
//		}
//
//		_, err = w.Write(out)
//		if err != nil {
//			rfid.HandleHttpWriteError(err)
//		}
//	})
//}

var getAnyCollectionHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := rfid.UrlDecodeString(r.PathValue("id"))
	if err != nil {
		env.LogIfDev(ctx, "failed to url decode string")
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}

	entryType := r.PathValue("variant")
	var bytes []byte
	env.LogIfDev(ctx, "getting "+entryType+" "+id)
	switch entryType {
	case "project": // Items with possible spaces in names but abnormal perms
		out, err := rfid.GetAltCollectionItem[*rfid.Project](ctx, id, &rfid.Project{})
		if err != nil {
			env.LogIfDev(ctx, "failed to get project: "+err.Error())
			http.Error(w, "failed to get project: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO: PERMS!
		user, err := rfid.GetAuthInfo(ctx)
		if err != nil {
			env.LogIfDev(ctx, "Failed to get authinfo: "+err.Error())
			http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var uats = "guest"      // TODO: del
		uat := user.AccountType // TODO: del
		if uat.IsAdmin() {      // TODO: del
			uats = "Admin" // TODO: del
		} else { // TODO: del
			if uat.IsRegular() { // TODO: del
				uats = "Regular user" // TODO: del
			} // TODO: del
		} // TODO: del
		env.LogIfDev(ctx, fmt.Sprintf(`Getting page for user %s, who is %s`, user.Email, uats)) // TODO: del
		if !user.IsAdmin() && out.Private {
			if user.AccountType.IsGuest() {
				env.LogIfDev(ctx, "guests are not authorized to view this project") // TODO: reveals too much info, given we already grabbed the project
				http.Error(w, "permission denied for project", http.StatusForbidden)
				return
			}
			if !out.Perms.ForUser(user.Email).CanRead() {
				env.LogIfDev(ctx, "permission denied to user for project") // TODO: reveals too much info, given we already grabbed the project
				http.Error(w, "perm denied to user for project", http.StatusForbidden)
				return
			}
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
			return
		}
		tempBs, err := json.MarshalIndent(out, "", "  ") // TODO: del
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType: "+err.Error())
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		env.LogIfDev(ctx, "returning item: "+string(tempBs))
	case "species", "subspecies": // Items with possible spaces in names but which have normal perms
		out, err := rfid.GetAltCollectionItem[rfid.PermissionedAltCollectionItem[string]](ctx, id, map[string]rfid.PermissionedAltCollectionItem[string]{
			"species":    &rfid.Species{},
			"subspecies": &rfid.Subspecies{},
		}[strings.ToLower(entryType)])
		if err != nil {
			http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// PERMS!
		user, err := rfid.GetAuthInfo(ctx)
		if err != nil {
			http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if out.Permissions().HighestPermFor(user) == nil { // TODO: validate working ok
			http.Error(w, "permission denied for sp/subsp", http.StatusForbidden)
			return
		}

		bytes, err = json.Marshal(out)
		if err != nil {
			http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
			return
		}
		tempBs, err := json.MarshalIndent(out, "", "  ") // TODO: del
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType: "+err.Error())
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		env.LogIfDev(ctx, "returning item: "+string(tempBs))
	case "user": // User (can have @)
		// TODO: ensure admin?????
		decodedId, err := rfid.UrlDecodeString(id)
		if err != nil {
			http.Error(w, "failed to decode user email: "+err.Error(), http.StatusInternalServerError)
			return
		}
		out, err := rfid.GetAltCollectionItem(ctx, decodedId, &rfid.User{})
		if err != nil {
			http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO: any permissions? validate user is admin?
		bytes, err = json.Marshal(out)
		if err != nil {
			http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
			return
		}
		tempBs, err := json.MarshalIndent(out, "", "  ") // TODO: del
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType: "+err.Error())
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		env.LogIfDev(ctx, "returning item: "+string(tempBs))
	// Cases which are alt colls with base58->binary ids
	case "agarBatch", "agarRecipe", "jarRecipe", "grainBatch", "lcRecipe", "pcRun", "sale", "substrateRecipe", "substrateBatch", "transfer":
		// TODO: maybe de-urlencode here to account for named recipes?
		altId, err := rfid.StandardizeAltCollectionId(id)
		if err != nil {
			http.Error(w, "failed to standardize alt coll id: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var out rfid.PermissionedAltCollectionItem[rfid.AlternateCollectionId]
		switch entryType {
		case "agarBatch":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.AgarBatch{})
			out, err = temp, errr
		case "agarRecipe":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.AgarRecipe{}) // TODO: handle recipe name?
			out, err = temp, errr
		case "grainBatch":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.GrainBatch{})
			out, err = temp, errr
			//out = temp
		case "jarRecipe":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.JarRecipe{}) // TODO: handle recipe name?
			out, err = temp, errr
		case "lcRecipe":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.LcRecipe{}) // TODO: handle recipe name?
			out, err = temp, errr
		case "pcRun":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.PCRun{})
			out, err = temp, errr
		case "sale":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.Sale{})
			out, err = temp, errr
		case "substrateBatch":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.SubstrateBatch{})
			out, err = temp, errr
		case "substrateRecipe":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.SubstrateRecipe{}) // TODO: handle recipe name?
			out, err = temp, errr
		case "transfer":
			temp, errr := rfid.GetAltCollectionItem(ctx, *altId, &rfid.Transfer{})
			out, err = temp, errr
		default:
			env.LogIfDev(ctx, "invalid entry type in getAnyCollHandler: "+entryType)
			http.Error(w, "invalid entry type in getAnyCollHandler: "+entryType, http.StatusBadRequest)
			return
		}
		if err != nil {
			env.LogIfDev(ctx, "failed to get alt collection itemType: "+err.Error())
			http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		authinfo, err := rfid.GetAuthInfo(ctx)
		if err != nil {
			env.LogIfDev(ctx, "Failed to get authinfo: "+err.Error())
			http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if out.Permissions().HighestPermFor(authinfo) == nil {
			env.LogIfDev(ctx, "item requested cannot be read by this user: "+err.Error())
			http.Error(w, "item requested cannot be read by this user: "+err.Error(), http.StatusForbidden)
			return
		}
		bytes, err = json.Marshal(out)
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType")
			http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
			return
		}
		tempBs, err := json.MarshalIndent(out, "", "  ") // TODO: del
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType: "+err.Error())
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		env.LogIfDev(ctx, "returning item: "+string(tempBs))
	default: // Main collection ids
		mainCollItem, err := rfid.MainCollItemForEntryType(entryType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// ensure id is in correct format
		mainCollId, err := rfid.StandardizeMainCollectionId(id)
		if err != nil {
			http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
			return
		}
		out, err := rfid.GetMainCollectionItem(ctx, *mainCollId, mainCollItem)
		if err != nil {
			env.LogIfDev(ctx, "failed to get mainCollItem for "+string(mainCollId.ToBinaryCollectionId().ToBase58Bytes())+": "+err.Error())
			http.Error(w, "failed to get main collection itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := rfid.GetAuthInfo(r.Context())
		if err != nil {
			http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
			return
		}
		userPermOnEntry := out.Permissions().HighestPermFor(user)
		if userPermOnEntry == nil {
			http.Error(w, "item requested cannot be read by this user", http.StatusForbidden)
			return
		}
		can := "read"
		if *userPermOnEntry == true {
			can = "write"
		}
		env.LogIfDev(ctx, "user "+user.Email+" got item and can "+can) // TODO: del?
		bytes, err = json.Marshal(out)
		if err != nil {
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tempBs, err := json.MarshalIndent(out, "", "  ") // TODO: del
		if err != nil {
			env.LogIfDev(ctx, "failed to marshal itemType: "+err.Error())
			http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			return
		}
		env.LogIfDev(ctx, "returning item: "+string(tempBs))
	}
	_, err = w.Write(bytes)
	if err != nil {
		rfid.HandleHttpWriteError(err)
	}
	return
}
var whitelistUserHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	var email string
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(bs, &email); err != nil {
		http.Error(w, "failed to unmarshal email: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: whitelist the user!
	rfid.UserWhitelist.Add(email)
}

//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//		//return r.Header.Get("Origin") == "<http://yourdomain.com>" // TODO: this to protect against Cross-Site websocket hijacking (CSWSH)
//	},
//}

// TODO: below this is rfid stuff! CONSIDER MOVING!

const goodTestRfid = "goodTestRfid"
const badTestRfid = "badTestRfid"

func withGoodBadTestWriters(actualWriters []string) []string {
	out := make([]string, 2, len(actualWriters)+2)
	out[0] = goodTestRfid
	out[1] = badTestRfid
	out = append(out, actualWriters...)
	return out
}

var getRfidReaderNamesHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	rfidReaderSessions := mgr.Sessions()
	totalSessions := rfidReaderSessions
	_ = env.IfNotProd(r.Context(), func() error { // TODO: del later
		totalSessions = withGoodBadTestWriters(totalSessions)
		return nil
	})

	out, err := json.Marshal(totalSessions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if _, err = w.Write(out); err != nil {
		fmt.Println("error writing getRfidReaderNamesHandler response:", err)
	}
}

var (
	invalidAcceptHeader     = "invalid Accept header"
	unsupportedHttpMethod   = "unsupported http method"
	unacceptableContentType = "invalid Content-Type"
)

var rfidReadHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	readerName := shared.RfidReaderName(r.PathValue("readerName"))
	println("trying to read from reader: " + readerName)
	ctx := r.Context()
	err := env.IfNotProd(ctx, func() error { // TODO: del later?
		if readerName == goodTestRfid { // TODO: remove later
			// TODO: multiple? not just one id?
			_, err := w.Write([]byte(rfid.EmptyTestPlateBinaryId().AsBase58()))
			if err != nil {
				println("failed to write internal result", err)
			}
			return errors.New("wrote")
		} else if readerName == badTestRfid {
			http.Error(w, "bad test rfid reader/writer selected", http.StatusInternalServerError)
			return errors.New("wrote")
		}
		return nil
	})
	if err != nil {
		return
	}

	mgr := websocketSessions.GetSessionManager(ctx)
	if mgr == nil {
		println("no session mgr found?") // TODO: del
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodGet {
		println("invalid method") // TODO: del
		http.Error(w, unsupportedHttpMethod, http.StatusBadRequest)
		return
	}
	headers := r.Header
	if headers.Get("Content-Type") != "text/html" {
		println("invalid content type header") // TODO: del
		http.Error(w, unacceptableContentType, http.StatusBadRequest)
		return
	}

	if r.Header.Get("Accept") != "text/html" {
		println("invalid accept header") // TODO: del
		http.Error(w, invalidAcceptHeader, http.StatusBadRequest)
		return
	}

	//bodyIn, err := io.ReadAll(r.Body)
	//if err != nil {
	//	http.Error(w, "unable to read request body: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//if !mgr.SecretValid(string(bodyIn)) {
	//	http.Error(w, "forbidden", http.StatusForbidden)
	//	return
	//}
	println("trying to read rfid") // TODO: del
	binaryUID, err := mgr.ReadRfid(ctx, readerName)
	if err != nil {
		println("read error " + err.Error()) // TODO: del
		// TODO: what type of error?
		http.Error(w, "failed to read rfid: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(rfid.MainCollectionId(binaryUID).AsBase58().Bytes())
	if err != nil {
		println("failed to write reader result", err)
	}
}

type writeTagRequest struct {
	Secret string
	Data   []byte // TODO: make sure we're ok with this being []byte? In-transit should be base58, but immediately become base2 in memory?
}

// TODO: consider moving to reader/internal side?
var rfidWriteHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	if r.Header.Get("Accept") != "text/html" {
		http.Error(w, invalidAcceptHeader, http.StatusBadRequest)
		return
	}

	toWrite, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	toWriteB58 := rfid.Base58Str(toWrite) // TODO: this is base58?!
	toWriteBytes, err := toWriteB58.ToMainCollectionId()
	if err != nil {
		http.Error(w, "unable to read request body. Invalid base58: "+err.Error(), http.StatusBadRequest)
	}
	//req := writeTagRequest{}
	//if err = json.Unmarshal(bodyIn, &req); err != nil {
	//	http.Error(w, "invalid request body structure", http.StatusBadRequest)
	//	return
	//}
	if r.Method != "POST" {
		http.Error(w, unsupportedHttpMethod, http.StatusBadRequest)
		return
	}
	headers := r.Header
	if headers.Get("Content-Type") != "application/json" {
		http.Error(w, unacceptableContentType, http.StatusBadRequest)
		return
	}
	writerName := shared.RfidReaderName(r.PathValue("writerName"))
	err = env.IfNotProd(r.Context(), func() error { // TODO: del later?
		if writerName == goodTestRfid {
			_, err = w.Write(toWriteB58.Bytes())
			if err != nil {
				println("failed to write internal result", err)
			}
			return errors.New("wrote")
		} else if writerName == badTestRfid {
			http.Error(w, "bad test rfid reader/writer selected", http.StatusInternalServerError)
			return errors.New("wrote")
		}
		return nil
	})
	if err != nil {
		return
	}

	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	//// TODO: fix to either use secret and internal, or no secret but external auth!
	//// TODO: what if this is something like id==1????
	//if len(toWri) != shared.RfidByteSize { // TODO: use constant for length! // TODO: this is a base58 string, shouldnt it always be that?
	//	// could be base58str
	//	req.Data, err = rfid.Base58Str(toWrite).Base2Bytes()
	//	if err != nil || len(req.Data) != shared.RfidByteSize {
	//		http.Error(w, "invalid request body data: "+string(req.Data), http.StatusBadRequest)
	//		return
	//	}
	//}
	//if !mgr.SecretValid(req.Secret) {
	//	http.Error(w, "forbidden", http.StatusForbidden)
	//	return
	//}

	if err = mgr.WriteRfid(ctx, writerName, toWriteBytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(toWriteB58)) // TODO: is this still ok if incoming was base58? Probably want to unmarshal into a binary one instead
	if err != nil {
		println("failed to write internal result", err)
	}
}

var clearRfidTagHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete { // TODO: implement
		http.Error(w, unsupportedHttpMethod, http.StatusBadRequest)
		return
	}
	if r.Header.Get("Accept") != "text/html" { // TODO: implement
		http.Error(w, invalidAcceptHeader, http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	toWriteBytes := [8]byte{0, 0, 0, 0, 0, 0, 0, 0} // TODO: ok?
	writerName := shared.RfidReaderName(r.PathValue("writerName"))
	validResponse := []byte("Cleared") // TODO: ok?
	mgr := websocketSessions.GetSessionManager(ctx)
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	err := env.IfNotProd(r.Context(), func() error { // TODO: del later?
		if writerName == goodTestRfid {
			_, err := w.Write(validResponse)
			if err != nil {
				println("failed write response", err)
			}
			return errors.New("wrote")
		} else if writerName == badTestRfid {
			http.Error(w, "bad test rfid reader/writer selected", http.StatusInternalServerError)
			return errors.New("wrote")
		}
		return nil
	})
	if err != nil {
		return
	}

	if err := mgr.WriteRfid(ctx, writerName, toWriteBytes); err != nil {
		http.Error(w, "failed to write rfid: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(validResponse)
	if err != nil {
		println("failed to clear rfid tag on writer: "+writerName, err)
	}
}

// TODO: everything below this is for guest sessions, consider moving to its own file!

func authUrlFor(providerName string) string {
	return fmt.Sprintf(`%s/auth/%s`, baseApiUrl, providerName)
}
func authCallbackUrlFor(providerName string) string {
	return fmt.Sprintf(`%s/auth/%s/callback`, baseApiUrl, providerName) // TODO: add state to the auth callback url to ensure we have destination params?
}

var _ goth.Provider = &guestLoginProvider{}
var _ goth.Session = &guestSession{}

func NewGuestProvider(callbackUrl string, providerName ...string) *guestLoginProvider {
	name := "guest"
	if len(providerName) > 0 {
		name = providerName[0]
	}
	return &guestLoginProvider{
		CallbackUrl:  callbackUrl,
		providerName: &name,
	}
}

type guestLoginProvider struct { // TODO: USE THIS?
	CallbackUrl  string
	providerName *string
	Config       oauth2.Config
}

func (p *guestLoginProvider) Name() string {
	if p.providerName != nil {
		return *p.providerName // TODO: ok?
	}
	return "guestProvider"
}

func (p *guestLoginProvider) SetName(name string) {
	newName := name
	p.providerName = &newName
}

func (p *guestLoginProvider) BeginAuth(state string) (goth.Session, error) {
	return &guestSession{
		AuthURL:      p.Config.AuthCodeURL(state), // TODO: fix the state
		AccessToken:  "",                          // TODO: fixme!
		RefreshToken: "",                          // TODO: fixme!
		ExpiresAt:    time.Time{},                 // TODO: fixme!
		IDToken:      "",                          // TODO: fixme!
	}, nil
}

func (p *guestLoginProvider) UnmarshalSession(data string) (goth.Session, error) {
	// UnmarshalSession will unmarshal a JSON string into a session.
	sess := &guestSession{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(sess)
	return sess, err
}

func (p *guestLoginProvider) FetchUser(session goth.Session) (goth.User, error) {
	sess, ok := session.(*guestSession)
	if !ok {
		return goth.User{}, errors.New("was not guest session")
	}
	user := goth.User{
		AccessToken:  sess.AccessToken,
		Provider:     p.Name(),
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt,
		IDToken:      sess.IDToken,
	}

	// Extract the user data we got from Google into our goth.User.
	user.Name = "guest"
	user.FirstName = ""
	user.LastName = ""
	user.NickName = "guest"
	user.Email = ""
	user.AvatarURL = ""
	user.UserID = ""
	return user, nil
}

func (p *guestLoginProvider) Debug(b bool) {
	// No-op
}

func (p *guestLoginProvider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	//ts := p.config.TokenSource(goth.ContextForClient(http.DefaultClient), token)
	//newToken, err := ts.Token()
	var err error
	newToken := token
	if err != nil {
		return nil, err
	}
	return newToken, err
}

func (p *guestLoginProvider) RefreshTokenAvailable() bool {
	return true
}

type guestSession struct {
	AuthURL      string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	IDToken      string
}

func (s guestSession) GetAuthURL() (string, error) {
	if s.AuthURL == "" {
		return "", errors.New(goth.NoAuthUrlErrorMessage)
	}
	return s.AuthURL, nil
}

func (s guestSession) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *guestSession) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	//p := provider.(*guestLoginProvider)
	accessToken := rand.Text() // TODO: fix?
	refreshToken := rand.Text()
	//token, err := p.config.Exchange(goth.ContextForClient(p.Client()), params.Get("code")) // TODO: reenable
	//if err != nil {
	//	return "", err
	//}
	//
	//if !token.Valid() {
	//	return "", errors.New("Invalid token received from provider")
	//}

	s.AccessToken = accessToken
	s.RefreshToken = refreshToken
	s.ExpiresAt = time.Now().Add(1 * time.Hour)
	return accessToken, nil
}

var corsAllowedOrigins = []string{
	"https://mush.appli.ng",
	"http://web", // TODO: ok?
	"http://api", // TODO: likely don't need
}

// TODO: USE THIS WHERE NEEDED
func enableCors(w *http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "*" {
		// TODO: what here?
	} else if slices.Contains(corsAllowedOrigins, origin) {
		(*w).Header().Set("Access-Control-Allow-Origin", origin)
	}

	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS") // TODO: which
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")    // TODO: ok? probably need more
	(*w).Header().Set("Access-Control-Max-Age", "Content-Type, Authorization")                  // Specifies how long preflight results can be cached            // TODO: ok? probably need more
	(*w).Header().Set("Access-Control-Expose-Headers", "Accept, Content-Type, Authorization")   // Lists headers accessible to JavaScript            // TODO: fix
	/*
		Access-Control-Allow-Origin: Specifies allowed origins or *.
		Access-Control-Allow-Methods: Lists allowed HTTP methods.
		Access-Control-Allow-Headers: Lists allowed HTTP headers.
		Access-Control-Allow-Credentials: Indicates if credentials are permitted.
		Access-Control-Expose-Headers: Lists headers accessible to JavaScript.
		Access-Control-Max-Age: Defines the preflight cache duration.
	*/
}

//var (
//	webAuthn *webauthn.WebAuthn
//	_        PasskeyUser = &rfid.User{} // TODO: delete later
//)
//func init(){
//	origin := "mush.appli.ng"
//	wconfig := &webauthn.Config{
//		RPDisplayName: "Go Webauthn",    // Display Name for your site  // TODO: fix
//		RPID:          "mush.appli.ng",             // Generally the FQDN for your site    // TODO: fix
//		RPOrigins:     []string{origin}, // The origin URLs allowed for WebAuthn    // TODO: fix?
//	}
//	var err error = nil
//	if webAuthn, err = webauthn.New(wconfig); err != nil {
//		fmt.Printf("[FATA] %s", err.Error())
//		os.Exit(1)
//	}
//
//}
//
//type PasskeyUser interface {
//	webauthn.User
//	AddCredential(*webauthn.Credential)
//	UpdateCredential(*webauthn.Credential)
//}
//
//type PasskeyStore interface {
//	GetUser(userName string) PasskeyUser
//	SaveUser(PasskeyUser)
//	GetSession(token string) webauthn.SessionData
//	SaveSession(token string, data webauthn.SessionData)
//	DeleteSession(token string)
//}
//
//func BeginFingerprintRegistration(w http.ResponseWriter, r *http.Request) {
//	userPerms, err := rfid.GetAuthInfo(r.Context())
//	if err != nil {
//		http.Error(w, "failed to get user for request: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	user, err := userPerms.GetUser(r.Context())
//	if err != nil {
//		http.Error(w, "failed to get user for request: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	//authSvc := rfid.GetAuthService(r.Context())
//
//	options, session, err := webAuthn.BeginRegistration(user)
//	if err != nil {
//		msg := fmt.Sprintf("can't begin registration: %s", err.Error())
//		http.Error(w, msg, http.StatusBadRequest)
//		return
//	}
//
//	// Make a session key and store the sessionData values
//	t := uuid.New().String()
//	datastore.SaveSession(t, *session) // TODO: save the session????
//
//	JSONResponse(w, t, options, http.StatusOK) // return the options generated with the session key
//	// options.publicKey contain our registration options
//}
//
//func FinishFingerprintRegistration(w http.ResponseWriter, r *http.Request) {
//	// Get the session key from the header
//	t := r.Header.Get("Session-Key")
//	// Get the session data stored from the function above
//	session := datastore.GetSession(t) // FIXME: cover invalid session
//
//	// In out example username == userID, but in real world it should be different   user := datastore.GetUser(string(session.UserID)) // Get the user
//
//	credential, err := webAuthn.FinishRegistration(user, session, r)
//	if err != nil {
//		msg := fmt.Sprintf("can't finish registration: %s", err.Error())
//		l.Printf("[ERRO] %s", msg)
//		JSONResponse(w, "", msg, http.StatusBadRequest)
//
//		return
//	}
//
//	// If creation was successful, store the credential object
//	user.AddCredential(credential)
//	datastore.SaveUser(user)
//	// Delete the session data
//	datastore.DeleteSession(t)
//
//	l.Printf("[INFO] finish registration ----------------------/")
//	JSONResponse(w, "", "Registration Success", http.StatusOK) // Handle next steps
//}
//
//type challengePayload struct {
//	Challenge string
//	// TODO: any more?
//}
//
//func generateRegistrationOptions(rpID string, rpName string, userName string) (payload challengePayload, err error) {
//	// TODO: DO THIS!
//}
//
//var biometricFingerprintRegisterChallengeHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	options, sessionData, err := BeginRegistration
//	config, err := configuration.Load([]string{"config.yaml"}, true, nil) // TODO: fix
//	if err != nil {
//		panic(err)
//	}
//
//	authorizer := webauthn.New()
//	//const { generateRegistrationOptions, verifyRegistrationResponse } = require('@simplewebauthn/server');
//	rpId := "localhost"        // TODO: change!
//	rpName := "My Application" // TODO: change!
//	userName := "emailHere"    // TODO: change!
//	payload, err := generateRegistrationOptions(rpId, rpName, userName)
//	if err != nil {
//		http.Error(w, "failed to generate registration options: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	bs, err := json.Marshal(payload)
//	if err != nil {
//		http.Error(w, "failed to marshal registration options: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	_, err = w.Write(bs)
//	rfid.HandleHttpWriteError(err)
//	//try {
//	//	const challengePayload = await generateRegistrationOptions({
//	//	rpID: 'localhost',
//	//	rpName: 'My Application',
//	//	userName: req.user.id,
//	//});
//	//	if (challengePayload) {
//	//	res.json({ code: 200, data: challengePayload, status: 'success' });
//	//} else {
//	//	res.json({ code: 500, message: 'Something went wrong. Please contact your administrator.', status: 'error' });
//	//}
//	//} catch (err) {
//	//	console.error(err);
//	//	res.status(500).json({ message: 'Internal Server Error' });
//	//}
//	// TODO: OVERHAUL!
//}
//
//type verifResp struct {
//	RegistrationInfo RegistrationInfo
//	Verified         bool
//}
//type RegistrationInfo struct {
//	// TODO: fix!
//}
//
//func verifyRegistrationResponse(expectedChallenge, expectedOrigin, expectedRPID string, response any /* TODO: what type?*/) (verifResp, error) {
//	// TODO: this!
//	return verifResp{}, errors.New("not implemented")
//}
//func saveFingerprint(userId string, registrationInfo RegistrationInfo) error {
//	// TODO: this!
//	return errors.New("not implemented")
//}
//
//var biometricFingerprintVerifyChallengeHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	expChallenge := ""      // req.body.challenge // TODO: fix
//	expOrigin := ""         // TODO: fix
//	expRPID := ""           // TODO: fix
//	reqBodyCredential := "" // req.body.credential // TODO: fix!
//	verificationResult, err := verifyRegistrationResponse(expChallenge, expOrigin, expRPID, reqBodyCredential)
//	if err != nil {
//		http.Error(w, "failed to verify registration response: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	user, err := rfid.GetAuthInfo(ctx) // TODO: ensure ok here...
//	if err != nil {
//		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	// Save the fingerprint information to the database
//	if err = saveFingerprint(user.Email, verificationResult.RegistrationInfo); err != nil {
//		http.Error(w, "failed to save fingerprint: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	_, err = w.Write([]byte("Biometric Registration Successful!"))
//	rfid.HandleHttpWriteError(err)
//	// TODO: OVERHAUL!
//	//const { generateRegistrationOptions, verifyRegistrationResponse } = require('@simplewebauthn/server');
//	//try {
//	//	const verificationResult = await verifyRegistrationResponse({
//	//	expectedChallenge: req.body.challenge,
//	//	expectedOrigin: 'https://myapp.example.com',
//	//	expectedRPID: 'myapp.example.com',
//	//	response: req.body.credential,
//	//});
//	//	if (verificationResult.verified) {
//	//	// Save the fingerprint information to the database
//	//	await saveFingerprint(req.user.id, verificationResult.registrationInfo);
//	//	res.json({ code: 200, status: 'success', message: 'Biometric Registration Successful!' });
//	//} else {
//	//	res.json({ code: 500, message: 'Invalid user!', status: 'error' });
//	//}
//	//} catch (err) {
//	//	console.error(err);
//	//	res.status(500).json({ message: 'Internal Server Error' });
//	//}
//	// TODO: OVERHAUL!
//}

func Middlewares(mws ...SingleMiddleware) SingleMiddleware {
	return func(next http.Handler) http.Handler {
		out := next
		for _, mw := range slices.Backward(mws) {
			out = mw(out)
		}
		return out
	}
}

type SingleMiddleware func(http.Handler) http.Handler
