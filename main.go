package main

import (
	"compress/gzip"
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
	"github.com/reeceappling/mushDb/rfid"
	"github.com/reeceappling/mushDb/rfid/pics"
	"github.com/reeceappling/mushDb/rfid/request"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/shared"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var baseApiUrl = "https://mush.appli.ng" // Overwritten main() early // TODO: likely set a different default

func setupDb(ctxIn context.Context) (ctx context.Context, client *mongo.Client, err error) {
	dbHostName := os.Getenv("DB_HOST_NAME")
	dbUser := os.Getenv("MONGO_INITDB_USERNAME")
	dbPass := os.Getenv("MONGO_INITDB_PASSWORD")
	dbSetupUser := os.Getenv("MONGO_INITDB_SETUP_USERNAME")
	dbSetupPass := os.Getenv("MONGO_INITDB_SETUP_PASSWORD")
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
	if dbPass == "" {
		err = errors.New("no MONGO_INITDB_PASSWORD env var found")
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

var oauthConfig *oauth2.Config

const defaultHttpPort = 8080

func main() {
	// TODO: need to clear out the db for actually using it!
	//env := os.Getenv("ENVIRONMENT") // TODO: set this!
	picsPath := os.Getenv("PICS_PATH")
	if picsPath == "" {
		panic("env var missing for PICS_PATH")
	}
	ctx := pics.SetFilePath(context.Background(), picsPath)
	//ctx = rfid.SetEnv(ctx, env == "prod")
	var err error
	authSvc := rfid.NewAuthService(utils.Pointer(2*time.Minute), utils.Pointer(1*time.Hour))
	ctx = authSvc.OnContext(ctx)

	dbUser := os.Getenv("MONGO_INITDB_USERNAME")
	dbPass := os.Getenv("MONGO_INITDB_PASSWORD")
	//dbSetupUser := os.Getenv("MONGO_INITDB_SETUP_USERNAME")
	//dbSetupPass := os.Getenv("MONGO_INITDB_SETUP_PASSWORD")
	// adminEmail := os.Getenv("ADMIN_EMAIL")

	// TODO: make sure logger is set up correctly
	log := logging.LoggerFactoryFor("mush-api-go") // TODO: ok name?
	ctx = logging.SetLogger(ctx, log)
	ctx = logging.SetSugaredLogger(ctx, log.Sugar())

	// Get non-db env vars
	rfidRegistrySecret := os.Getenv("RFID_SECRET")
	if rfidRegistrySecret == "" {
		panic("env var missing for RFID_SECRET")
	}
	googId, googSecret := os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")
	apiProtocol := strings.ToLower(os.Getenv("MAIN_API_EXTERNAL_PROTOCOL"))
	MAIN_API_EXTERNAL_HOST := os.Getenv("MAIN_API_EXTERNAL_HOST")
	apiPort, err := strconv.Atoi(os.Getenv("API_PORT"))
	if err != nil {
		fmt.Printf(`No api port configured, defaulting to port %d`, defaultHttpPort)
		apiPort = defaultHttpPort
	}
	tempPortStr := ""
	if (apiProtocol == "https" && apiPort != 443) || (apiProtocol == "http" && (apiPort != 80 && apiPort != defaultHttpPort)) { // TODO: 80 AND 8080 ok here?
		tempPortStr = fmt.Sprintf(`:%d`, apiPort)
	}
	apiHostPort := MAIN_API_EXTERNAL_HOST + tempPortStr
	baseApiUrl = fmt.Sprintf("%s://%s", apiProtocol, apiHostPort)
	oauthConfig = &oauth2.Config{
		ClientID:     googId,
		ClientSecret: googSecret,
		RedirectURL:  baseApiUrl + "/auth/google/callback",
		// TODO: works but want to use other... RedirectURL: "http://mush.appli.ng/auth/google/callback", // TODO: ensure fixed
		Scopes:   []string{"email", "profile", "openid"},
		Endpoint: google.Endpoint,
	}

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

	webHostName := envVarOrDefault("WEB_HOST_INTERNAL", "web") // Can have port if not hosting on 80      // TODO: CONFIGURE (localhost default or web?)

	// Set up server
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(apiPort),
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
	googleAuthProvider := google2.New(googId, googSecret, baseApiUrl+"/auth/google/callback", "email", "profile", "openid") // TODO: ensure callback is ok

	goth.UseProviders(googleAuthProvider)
	//gothic.Store = customSessionStore{} // TODO: REENABLE????

	http.HandleFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
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
	const loginPath = "/login"
	cleanupFreq := 2 * time.Minute
	// TODO: rfid sessions mock readers?
	mgr := websocketSessions.NewSessionManager(&cleanupFreq, rfidRegistrySecret)
	defer mgr.Cleanup()
	// Start generating mainCollectionIds
	rfid.StartGeneratingMCIDs(ctx, 12)
	internalOnlyMiddleware := internalOnlyMiddlewareCreator("api", apiPort)
	ctx, rateLimiter, rfidMiddleware, webAuthMiddleware, _ /*internalAuthMiddleware*/, ctxInternalAuthMiddleware, ctxMiddleware, ctxRfidMiddleware, err := setupMiddlewares(ctx, mgr, loginPath, dbUser, dbPass)
	if err != nil {
		panic("Error setting up middleware: " + err.Error())
	}

	// RFID HANDLERS
	println("Defining endpoints")
	println("Defining RFID endpoints")
	// Must be publicly available. (external)
	http.HandleFunc("/rfid/ws", ctxRfidMiddleware(http.HandlerFunc(websocketSessions.ServerHandler)))
	// Must be internal to docker network
	http.HandleFunc("/rfid/read/{readerName}", ctxRfidMiddleware(internalOnlyMiddleware(rfidReadHandler)))   //  OUTPUT IS BASE 2! // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/write/{writerName}", ctxRfidMiddleware(internalOnlyMiddleware(rfidWriteHandler))) // INPUT IS BASE58. OUTPUT IS BASE 2! // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/readers", ctxRfidMiddleware(internalOnlyMiddleware(getRfidReaderNamesHandler)))   // DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)

	// SERVER HANDLERS! (PASSTHROUGH) view, new, import
	webHostPort := 3000
	webHostPortStr := os.Getenv("WEB_HOST_INTERNAL_PORT")
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

	println("Defining webserver passthrough endpoints (external)")
	webProxyHandler := newPassthroughHandler(passthroughConfig)
	unAuthedProxied := ctxMiddleware(webProxyHandler)
	authedProxied := ctxMiddleware(webAuthMiddleware(webProxyHandler))

	// handle login
	http.Handle("/login", CorsMiddleware(rateLimiter(ctxMiddleware(handleLogin(webProxyHandler, dbUser, dbPass)))))
	http.Handle("/logout", CorsMiddleware(rateLimiter(ctxMiddleware(handleLogout))))
	http.Handle("/guestLogin", rateLimiter(ctxMiddleware(handleGuestLogin(baseApiUrl))))
	http.Handle("/auth/{provider}", CorsMiddleware(rateLimiter(ctxMiddleware(handleAuthProvider()))))
	http.Handle("/auth/{provider}/callback", CorsMiddleware(rateLimiter(ctxMiddleware(handleAuthCallback()))))

	// Proxied to react
	http.Handle("/_next", unAuthedProxied)
	http.Handle("/", unAuthedProxied)

	http.Handle("/import/{variant}", authedProxied)         // GET import is here (import item page)
	http.Handle("/new/{variant}", authedProxied)            // GET new item is here (new item page)
	http.Handle("/view/{variant}/{entryId}", authedProxied) // GET view item is here (view item page)
	http.Handle("/list/{variant}", authedProxied)           // GET list is here (list items page)
	http.Handle("/error/{errTxt}", webProxyHandler)         // TODO: rate limit???? ctx middleware? auth middleware?
	http.Handle("/testpage", webProxyHandler)               // GET testpage is here (test page)       // TODO: REMOVE

	println("Defining sensor data endpoints")
	// TODO: this
	// Sensor data endpoints
	//http.Handle("/sensorData/{nodeName}", rfid.GetSensorDataHandler())           // TODO: middleware?
	//http.Handle("/sensorDataSince/{nodeName}", rfid.GetSensorDataSinceHandler()) // TODO: middleware?
	//http.Handle("/addSensorData/{nodeName}", rfid.AddSensorDataHandler())        // TODO: middleware?

	println("Defining admin endpoints")
	// TODO: ADMIN STUFF
	// TODO: user-viewer/editor for admin
	// TODO: need to be able to create new users

	println("Defining db interaction endpoints")
	http.Handle("/db/pathFor/{id}", rateLimiter(ctxInternalAuthMiddleware(rfid.GetPageForIdHandler())))
	//http.Handle("/db/get/rfid/{id}", getRfidHandler())             // TODO: GET RID OF???             // TODO: ensure this works for base58s
	http.Handle("/db/get/{variant}/{id}", rateLimiter(ctxInternalAuthMiddleware(getAnyCollectionHandler())))
	http.Handle(fmt.Sprintf(`%s{%s...}`, imagesEndpoint, imageSubPathKey), ctxInternalAuthMiddleware(getImageHandler()))
	// Creation handlers
	http.Handle("/db/create/{variant}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.HandleCreate()))))
	// update handlers
	http.Handle("/db/update/{endpt}/{id}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.UpdateById()))))
	// import handlers
	http.Handle("/db/import/{endpt}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.ImportHandler()))))
	// List handlers
	http.Handle("/db/list/{variant}", ctxInternalAuthMiddleware(rfid.ListEntriesHandler()))
	http.Handle("/subspeciesFor/{variant}", ctxInternalAuthMiddleware(rfid.ListSubspeciesHandler()))
	http.Handle("/sessionUserProjects", ctxInternalAuthMiddleware(rfid.SessionUserProjectsHandler()))

	println("Defining simple api endpoints")
	http.Handle("/options/{optionsType}", internalOnlyMiddleware(rfid.GetOptionsHandler))

	if err = srv.ListenAndServe(); err != nil {
		panic("failed to listen and serve for http: " + err.Error())
	}
	if err != nil {
		panic("ERROR CLOSING SERVER " + err.Error())
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

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func setupMiddlewares(ctxIn context.Context, mgr *websocketSessions.SessionManager, loginPath, dbUser, dbPass string) (
	ctx context.Context,
	rateLimiter, rfidMiddleware, webAuthMiddleware, internalAuthMiddleware func(http.Handler) http.Handler,
	ctxInternalAuthMiddleware, ctxMiddleware, ctxRfidMiddleware func(http.Handler) http.HandlerFunc,
	err error) {
	ctx = ctxIn
	//rateLimitCount, rateLimitPeriod := int64(150), 5*time.Minute // TODO: reenable for prod
	rateLimitCount, rateLimitPeriod := int64(150), time.Minute
	// Rate limiter
	rate := limiter.Rate{ // TODO: TURN
		Period: rateLimitPeriod,
		Limit:  rateLimitCount,
	}
	rateLimiterStorage := memory.NewStore()
	rateLimiterMiddleware := stdlib.NewMiddleware(limiter.New(rateLimiterStorage, rate, limiter.WithTrustForwardHeader(true)))
	rateLimiter = rateLimiterMiddleware.Handler
	// PicsPath and rfid middleware
	// Pics path middleware
	picsPath := os.Getenv("PICS_PATH")
	if picsPath == "" {
		panic("env var missing for PICS_PATH")
	}
	ctx = pics.SetFilePath(ctx, picsPath)
	ctxMiddleware = func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx)) // TODO: ok?
		}
	}
	rfidMiddleware = mgr.Middleware()
	// Auth Middleware
	svc := rfid.GetAuthService(ctx)
	webAuthMiddleware = svc.AuthOrRedirectMiddleware(loginPath)
	internalAuthMiddleware = svc.AuthOrDenyMiddleware()

	ctxRfidMiddleware = func(next http.Handler) http.HandlerFunc {
		return ctxMiddleware(rfidMiddleware(next))
	}
	ctxInternalAuthMiddleware = func(next http.Handler) http.HandlerFunc {
		return ctxMiddleware(internalAuthMiddleware(next))
	}

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

func handleGuestLogin(defaultUrl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		println("creating new session id for guest")
		sessId, err := rfid.GetAuthService(ctx).SigninGuestUser(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		println("adding guest session to goth store")
		session, err := gothic.Store.New(r, string(sessId))
		if err != nil {
			// TODO: delete guest session
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		println("adding session id to session")
		err = gothic.StoreInSession(rfid.SessionIdKey, string(sessId), r, w)
		if err != nil {
			// TODO: del sess?
			err = errors.Join(errors.New("sessId storage fail"), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		println("saving guest session")
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectToBasePage(r, w)
	})
}

func handleLogin(viewHandler http.HandlerFunc, rootUser, rootPass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			viewHandler.ServeHTTP(w, r)
		case http.MethodPost:
			time.Sleep(3 * time.Second) // TODO: Make user wait for login, lower likelihood of attack
			println("SHOULD NEVER BE HIT")
		default:
			http.Error(w, "Unsupported http request method: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

	})
}

func handleUserAuthedViaGoth(ctx context.Context, user goth.User, w http.ResponseWriter, r *http.Request) error {
	authSvc := rfid.GetAuthService(ctx)
	sessId, _, err := authSvc.SigninGoogleAuthedUser(ctx, user)
	if err != nil {
		return errors.Join(errors.New("authSvc signin fail"), err)
	}
	err = gothic.StoreInSession(rfid.SessionIdKey, string(sessId), r, w)
	if err != nil {
		return errors.Join(errors.New("sessId storage fail"), err)
	}
	return nil
}

func getSessionValue(session *sessions.Session, key string) (string, error) {
	value := session.Values[key]
	if value == nil {
		return "", fmt.Errorf("could not find a matching session for this request")
	}

	rdata := strings.NewReader(value.(string))
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return "", err
	}
	s, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

func adminLogin(w http.ResponseWriter, r *http.Request) {

}

func handleAuthProvider() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("RECEIVED REQUEST at /auth/google!!!")
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			println("failed to read body", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		println(string(bs))

		user, err := gothic.CompleteUserAuth(w, r)
		if err == nil {
			println("authed, handling authentication") // TODO: del
			// TODO: why does this and the callback both call Complete?
			// TODO: do we only need one????
			err = handleUserAuthedViaGoth(r.Context(), user, w, r)
			if err != nil {
				println("failed to handle user auth: " + err.Error())
				//http.Error(w, err.Error(), http.StatusInternalServerError)

			}
			w.Write([]byte("OK!"))
			return
			// t, _ := template.New("foo").Parse(userTemplate)
			//t.Execute(w, gothUser)
		}
		println("not authed, beginning authentication by redirecting") // TODO: del

		//gothic.BeginAuthHandler(w, r)
		url, err := gothic.GetAuthURL(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

type SessionCookieStore struct {
	internal *rfid.AuthService
	sessions.Store
}

const sessionCookieName = "SessionId"

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

func handleAuthCallback() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		println("hit callback")
		// Clear cookies if needed
		if len(r.CookiesNamed("_gothic_session")) > 1 {
			c, _ := r.Cookie("_gothic_session")
			c.MaxAge = -1
			http.SetCookie(w, c)
			http.Redirect(w, r, "/auth/google/callback", http.StatusTemporaryRedirect)
			return
		}
		c, err := r.Cookie("_gothic_session")
		if err != nil {
			println("no gothic session cookie")
		} else {
			println("session cookie", c.Name, c.Value)
		}

		// TODO: ensure goth storage expirations match our storage expirations
		// TODO: OVERHAUL THE COOKIE STORAGE???
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			println("failed to auth in callback", err.Error())
			if strings.Contains(err.Error(), "state token mismatch") {
				err = gothic.Logout(w, r)
				if err != nil {
					println("failed to logout: " + err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/auth/google/callback", http.StatusTemporaryRedirect)
				return
			}
		}

		// Loop back on initial failure
		//reqState := gothic.GetState(r)
		//
		//originalState := authURL.Query().Get("state")
		//if originalState != "" && (originalState != reqState) {
		//	gothic.BeginAuthHandler(w, r)
		//	return
		//}
		err = handleUserAuthedViaGoth(ctx, user, w, r)
		if err != nil {
			http.Error(w, "failed in handleUserAuthedViaGoth: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = r.Cookie("_gothic_session")
		if err != nil {
			println("no cookie after auth", err.Error()) // TODO: ensure ok
			//http.Error(w, "no cookie after auth", http.StatusInternalServerError)
			//return
		}
		//http.SetCookie(w, c) // TODO: if this works do it everywhere
		// redirect to original page
		redirectToBasePage(r, w)
	})
}

func redirectToBasePage(r *http.Request, w http.ResponseWriter) {
	dst := r.URL.Query().Get("destination")
	if dst == "" {
		dst = baseApiUrl
	}
	http.Redirect(w, r, dst, http.StatusTemporaryRedirect)
}

var handleLogout = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	authSvc := rfid.GetAuthService(r.Context())
	err := gothic.Logout(w, r)
	if err != nil {
		http.Error(w, "logout fail: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sessId, err := rfid.SessionIdFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = authSvc.LogoutSession(sessId)
	if err != nil {
		http.Error(w, "failed to log out of session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/login")
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

const imageSubPathKey = "imageSubPath"
const imagesEndpoint = "/db/images/" // MUST match PicsEndpoint in PicWithNotes.tsx
func getImageHandler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: make sure authed?
		ctx := r.Context()
		imgSubPath := r.PathValue(imageSubPathKey)
		if imgSubPath == "" {
			http.Error(w, "image name must not be blank", http.StatusBadRequest)
			return
		}
		println("trying to read picture file: " + filepath.Join(pics.GetFilePath(ctx), imgSubPath)) // TODO: DEL
		bytes, err := os.ReadFile(filepath.Join(pics.GetFilePath(ctx), imgSubPath))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				println("file does not exist!") // TODO: fix
				http.Error(w, "image not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error retrieving image. "+err.Error(), http.StatusInternalServerError)
		}
		//bytes, err := pics.GetFile(ctx, imgSubPath)
		//if err != nil {
		//	if errors.Is(err, pics.ErrNotFound) {
		//		http.Error(w, "image not found", http.StatusNotFound)
		//		return
		//	}
		//	http.Error(w, "Error retrieving image. "+err.Error(), http.StatusInternalServerError)
		//}
		_, err = w.Write(bytes)
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}
	})
	return rfid.GetPermsMiddleware(handler)
}

var rootHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("an apple a day: from " + r.URL.Path)) //nolint:errcheck // TODO: DO WE EVEN WANT THE ROOT TO RESPOND?
	if err != nil {
		rfid.HandleHttpWriteError(err)
	}
})

// TODO: use this somewhere!
// getItemTypeForId request body is just a base58 string of the mainCollectionId
func getItemTypeForId() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req := rfid.MainCollectionId{}
		if err = json.Unmarshal(bs, &req); err != nil {
			http.Error(w, "Error parsing body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		itemType, err := rfid.FindItemTypeForId(r.Context(), req)
		if err != nil {
			http.Error(w, "Error finding item type: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(itemType.EntryType()))
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}

	})
}

func getRfidHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		idBytes := []byte(id)
		if len(idBytes) != rfid.RfidByteSize {
			http.Error(w, "invalid id format. Must be 8 bytes", http.StatusBadRequest)
			return
		}
		item, err := rfid.GetMainCollectionItemWithId(ctx, [rfid.RfidByteSize]byte(idBytes))
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "not found: "+err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, "failed to retrieve item: "+err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := rfid.GetAuthInfo(ctx)
		if err != nil {
			http.Error(w, "failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Validate user can read this entry
		if item.Permissions().HighestPermFor(user) == nil {
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}
		out, err := json.Marshal(item)
		if err != nil {
			http.Error(w, "failed to marshal item", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(out)
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}
	})
}

// TODO: FIXME!!!!
func getAnyCollectionHandler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("GETTING ITEM") // TODO: del
		ctx := r.Context()
		// TODO: ?????? id := strings.ReplaceAll(r.PathValue("id"), "_", " ") // TODO: replace all underscores with spaces, for things like "chicken of the woods"
		//id := strings.ReplaceAll(r.PathValue("id"), "_", " ") // TODO: replace all underscores with spaces, for things like "chicken of the woods"
		id, err := rfid.UrlDecodeString(r.PathValue("id"))
		if err != nil {
			println("failed to url decode string") // TODO; DEL
			http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
			return
		}
		println("URL ID", id) // TODO: DELETEME

		entryType := r.PathValue("variant")
		var bytes []byte
		switch entryType {
		case "project", "species", "subspecies": // Items with possible spaces in names
			// TODO: ensure to convert id from url format to server format
			println("getting " + entryType + " " + id) // TODO: DEL
			out, err := rfid.GetAltCollectionItem[rfid.AltCollectionItem[string]](ctx, id, map[string]rfid.AltCollectionItem[string]{
				"project":    &rfid.Project{},
				"species":    &rfid.Species{},
				"subspecies": &rfid.Subspecies{},
			}[strings.ToLower(entryType)]) // TODO: validate works
			if err != nil {
				println("err getting altCollType " + err.Error()) // TODO: DEL
				http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// TODO: ENSURE USER HAS CORRECT PERMS
			//authinfo, err := rfid.GetAuthInfo(ctx)
			//if err != nil {
			//	http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
			//	return
			//}
			println("marshalling bytes")
			bytes, err = json.Marshal(out)
			if err != nil {
				http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
				return
			}
			//println("marshalling indented bytes")
			//tempBs, err = json.MarshalIndent(out, "", "  ") // TODO: del
			//if err != nil {
			//	http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			//	return
			//}
		case "user": // User (can have @)
			// TODO: ensure to convert id from url format to server format!!!!!
			decodedId, err := rfid.UrlDecodeString(id)
			if err != nil {
				http.Error(w, "failed to decode user email: "+err.Error(), http.StatusInternalServerError)
				return
			}
			out, err := rfid.GetAltCollectionItem(ctx, decodedId, &rfid.User{})
			if err != nil {
				http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
				return
			} // TODO: validate works
			bytes, err = json.Marshal(out)
			if err != nil {
				http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
				return
			}
			//tempBs, err = json.MarshalIndent(out, "", "  ") // TODO: del
			//if err != nil {
			//	http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			//	return
			//}
		// Cases which are alt colls with base58->binary ids
		case "agarBatch", "agarRecipe", "jarRecipe", "grainBatch", "lcRecipe", "pcRun", "sale", "substrateRecipe", "substrateBatch", "transfer":
			// TODO: maybe de-urlencode here to account for named recipes?
			altId, err := rfid.StandardizeAltCollectionId(id)
			if err != nil {
				http.Error(w, "failed to standardize alt coll id: "+err.Error(), http.StatusInternalServerError)
				return
			}
			//altId := [12]byte{}
			//if len(id) < 12 {
			//	for i, byt := range altId {
			//		altId[12-len(altId)+i] = byt
			//	}
			//} else {
			//	temp := []byte(id)
			//	altId = [12]byte(temp)
			//}
			baseItem, exists := map[string]rfid.PermissionedAltCollectionItem[rfid.AlternateCollectionId]{
				"agarBatch":       &rfid.AgarBatch{},
				"agarRecipe":      &rfid.AgarRecipe{}, // TODO: handle recipe name?
				"grainBatch":      &rfid.GrainBatch{},
				"jarRecipe":       &rfid.JarRecipe{}, // TODO: handle recipe name?
				"lcRecipe":        &rfid.LcRecipe{},  // TODO: handle recipe name?
				"pcRun":           &rfid.PCRun{},
				"sale":            &rfid.Sale{},
				"substrateBatch":  &rfid.SubstrateBatch{},
				"substrateRecipe": &rfid.SubstrateRecipe{}, // TODO: handle recipe name?
				"transfer":        &rfid.Transfer{},
			}[entryType]
			if !exists {
				http.Error(w, "invalid entry type in getAnyCollHandler: "+entryType, http.StatusBadRequest)
			}

			out, err := rfid.GetAltCollectionItem(ctx, rfid.AlternateCollectionId(altId[:]), baseItem)
			if err != nil {
				http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
				return
			}
			authinfo, err := rfid.GetAuthInfo(ctx)
			if err != nil {
				http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if out.Permissions().HighestPermFor(authinfo) == nil {
				println("user does not have permission", err.Error()) // TODO; del
				http.Error(w, "item requested cannot be read by this user: "+err.Error(), http.StatusForbidden)
				return
			}
			bytes, err = json.Marshal(out)
			if err != nil {
				http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
				return
			}
			//tempBs, err = json.MarshalIndent(out, "", "  ") // TODO: del
			//if err != nil {
			//	http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
			//	return
			//}
		default: // Main collection ids
			if mainCollItem, exists := map[string]rfid.MainCollectionItem{
				"bag":             &rfid.Bag{}, // can only go to fruits
				"fruit":           &rfid.Fruit{},
				"fruitingChamber": &rfid.FruitingChamber{}, // can only go to fruits
				"jar":             &rfid.GrainJar{},        // can go anywhere (in theory) except MSS
				"lc":              &rfid.LiquidCulture{},   // can go anywhere (in theory) except MSS
				"lcSyringe":       &rfid.LcSyringe{},
				"mss":             &rfid.MSS{},   // generally only goes to plate
				"plate":           &rfid.Plate{}, // can go anywhere (in theory) except MSS
				"plugs":           &rfid.PlugsJar{},
				"slant":           &rfid.Slant{}, // generally only goes to plate
				"sporePrint":      &rfid.SporePrint{},
				"sporeSwab":       &rfid.SporeSwab{},
				"stasisTube":      &rfid.StasisTube{}, // generally only goes to plate
				"waterJar":        &rfid.WaterJar{},
			}[entryType]; exists {
				// ensure id is in correct format
				mainCollId, err := rfid.StandardizeMainCollectionId(id)
				if err != nil {
					println("failed to standardize main collection id: " + err.Error()) // TODO: del
					http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
					return
				}
				out, err := rfid.GetMainCollectionItem(ctx, *mainCollId, mainCollItem)
				if err != nil {
					println("failed to get mainCollItem for "+string(mainCollId.ToBinaryCollectionId().ToBase58Bytes()), err.Error()) // TODO: del
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
					http.Error(w, "item requested cannot be read by this user: "+err.Error(), http.StatusForbidden)
					return
				}
				can := "read"
				if *userPermOnEntry == true {
					can = "write"
				}
				println("user got item and can " + can) // TODO: del?
				bytes, err = json.Marshal(out)
				if err != nil {
					http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
					return
				}
				//tempBs, err = json.MarshalIndent(out, "", "  ") // TODO: del
				//if err != nil {
				//	http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
				//	return
				//}
			}
		}
		_, err = w.Write(bytes)
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}
		return
	})
	return rfid.GetPermsMiddleware(handler)
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
	totalSessions = withGoodBadTestWriters(totalSessions) // TODO: remove after testing
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
	if readerName == goodTestRfid {
		_, err := w.Write([]byte(rfid.EmptyTestPlateBinaryId().AsBase58()))
		if err != nil {
			println("failed to write internal result", err)
		}
		return
	} else if readerName == badTestRfid {
		http.Error(w, "bad test rfid reader/writer selected", http.StatusInternalServerError)
		return
	}
	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method != "POST" {
		http.Error(w, unsupportedHttpMethod, http.StatusBadRequest)
		return
	}
	headers := r.Header
	if headers.Get("Content-Type") != "text/html" {
		http.Error(w, unacceptableContentType, http.StatusBadRequest)
		return
	}

	if r.Header.Get("Accept") != "text/html" {
		http.Error(w, invalidAcceptHeader, http.StatusBadRequest)
		return
	}

	bodyIn, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !mgr.SecretValid(string(bodyIn)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	binaryUID, err := mgr.ReadRfid(readerName)
	if err != nil {
		// TODO: what type of error?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(binaryUID[:])
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
	println("Received rfid write request for: ", r.URL.String()) // TODO: del
	if r.Header.Get("Accept") != "text/html" {
		http.Error(w, invalidAcceptHeader, http.StatusBadRequest)
		return
	}

	bodyIn, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := writeTagRequest{}
	if err = json.Unmarshal(bodyIn, &req); err != nil {
		http.Error(w, "invalid request body structure", http.StatusBadRequest)
		return
	}
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
	if writerName == goodTestRfid { // TODO: del
		_, err = w.Write(req.Data) // TODO: is this still ok if incoming was base58?
		if err != nil {
			println("failed to write internal result", err)
		}
		return
	} else if writerName == badTestRfid {
		http.Error(w, "bad test rfid reader/writer selected", http.StatusInternalServerError)
		return
	}

	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: what if this is something like id==1????
	if len(req.Data) != 8 { // TODO: this is a base58 string, shouldnt it always be that?
		// could be base58str
		req.Data, err = rfid.Base58Str(req.Data).Base2Bytes()
		if err != nil || len(req.Data) != 8 {
			http.Error(w, "invalid request body data: "+string(req.Data), http.StatusBadRequest)
			return
		}
	}
	if !mgr.SecretValid(req.Secret) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	toWrite := [8]byte(req.Data) // TODO: ensure base2?
	if err = mgr.WriteRfid(writerName, toWrite); err != nil {
		// TODO: what type of error?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(req.Data) // TODO: is this still ok if incoming was base58? Probably want to unmarshal into a binary one instead
	if err != nil {
		println("failed to write internal result", err)
	}
}

// TODO: everything below this is for guest sessions, consider moving to its own file

var _ goth.Provider = &guestLoginProvider{}
var _ goth.Session = &guestSession{}

type guestLoginProvider struct {
	providerName string
}

func (p *guestLoginProvider) Name() string {
	return p.providerName
}

func (p *guestLoginProvider) SetName(name string) {
	p.providerName = name
}

func (p *guestLoginProvider) BeginAuth(state string) (goth.Session, error) {
	return &guestSession{
		AuthURL: "", // TODO: FIXME!
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
	accessToken := rand.Text()
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
