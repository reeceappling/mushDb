package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
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
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func setupDb(ctxIn context.Context) (ctx context.Context, client *mongo.Client, err error) {
	dbHostName := os.Getenv("DB_HOST_NAME")
	dbUser := os.Getenv("MONGO_INITDB_USERNAME")
	dbPass := os.Getenv("MONGO_INITDB_PASSWORD")
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

	println("Connecting to database")
	ctx, client, err = rfid.NewMongoDbClient(ctxIn, dbUser, dbPass, dbHostName, dbHostPort)
	if err != nil {
		return ctx, nil, errors.Join(errors.New("failed to create MongoDB client"), err)
	}
	println("Initializing DB")
	if err = rfid.Initialize(ctx); err != nil {
		return ctx, nil, errors.Join(errors.New("failed to initialize database"), err)
	}
	println("DB setup and connection complete!")
	return ctx, rfid.GetMongoClient(ctx), nil
}

var conf *oauth2.Config

func main() {
	// TODO: MAKE SURE TO STORE USERID ON COOKIES AS WELL?
	var err error
	ctx := context.Background()

	dbUser := os.Getenv("MONGO_INITDB_USERNAME")
	dbPass := os.Getenv("MONGO_INITDB_PASSWORD")

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
		println("No api port configured, defaulting to port 80")
		apiPort = 80
	}
	tempPortStr := "" // TODO: ensure ok
	if (apiProtocol == "https" && apiPort != 443) || (apiProtocol == "http" && apiPort != 80) {
		tempPortStr = fmt.Sprintf(`:%d`, apiPort)
	}
	conf = &oauth2.Config{
		ClientID:     googId,
		ClientSecret: googSecret,
		RedirectURL:  apiProtocol + "://" + MAIN_API_EXTERNAL_HOST + tempPortStr + "/auth/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	ctx, client, err := setupDb(ctx)
	if err != nil {
		panic("Error setting up db: " + err.Error()) // TODO: ok?
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
		ReadHeaderTimeout: 10 * time.Second, // TODO: ensure ok, was 30?
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

		srvCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // default ecs shutdown is 30 seconds // TODO: change
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

	println("Creating Middleware")
	// TODO: NEXT.JS LOGS! WEBPACK IS GETTING TRAPPED BY MAIN SERVER! WEBSERVER ENV VARS!
	// Setup middlewares
	const loginPath = "/login" // TODO: REENABLE
	cleanupFreq := 2 * time.Minute
	mgr := websocketSessions.NewSessionManager(&cleanupFreq, rfidRegistrySecret)
	defer mgr.Cleanup()
	ctx, rateLimiter, rfidMiddleware, picPathMiddleware, webAuthMiddleware, _ /*internalAuthMiddleware*/, ctxInternalAuthMiddleware, ctxMiddleware, ctxRfidMiddleware, err := setupMiddlewares(ctx, mgr, loginPath, dbUser, dbPass)
	if err != nil {
		panic("Error setting up middleware: " + err.Error())
	}

	// RFID HANDLERS
	println("Defining endpoints")
	println("Defining RFID endpoints")
	// Must be publicly available. (external)
	http.HandleFunc("/rfid/ws", ctxRfidMiddleware(http.HandlerFunc(websocketSessions.ServerHandler)))
	// Can be internal to docker network
	// TODO: are these internal or external?
	http.HandleFunc("/rfid/read/{readerName}", ctxRfidMiddleware(rfidReadHandler))   // TODO: OUTPUT IS BASE 2! // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/write/{writerName}", ctxRfidMiddleware(rfidWriteHandler)) // TODO: INPUT IS BASE58. OUTPUT IS BASE 2! // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	// internal
	http.HandleFunc("/rfid/readers", ctxRfidMiddleware(getRfidReaderNamesHandler)) // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)

	// SERVER HANDLERS! (PASSTHROUGH) view, new, import
	webHostPort := 3000
	webHostPortStr := os.Getenv("WEB_HOST_INTERNAL_PORT") // TODO: configure
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
	// TODO: maybe create a more readable middleware setup???
	webProxyHandler := newPassthroughHandler(passthroughConfig)
	unAuthedProxied := ctxMiddleware(webProxyHandler)
	authedProxied := ctxMiddleware(webAuthMiddleware(webProxyHandler))
	// TODO: handle login????
	http.Handle("/login", rateLimiter(ctxMiddleware(handleLogin(webProxyHandler, dbUser, dbPass))))         // GET=view, POST=do
	http.Handle("/signup", rateLimiter(ctxMiddleware(rfid.SignupHandler(webProxyHandler, dbUser, dbPass)))) // GET=view, POST=do
	http.Handle("/confirmSignup/{token}", rateLimiter(ctxMiddleware(rfid.ConfirmSignupHandler)))
	// TODO: auth callback? /auth/callback
	http.Handle("/_next", unAuthedProxied) // TODO: CHANGE ROOT?
	http.Handle("/", unAuthedProxied)      // TODO: CHANGE ROOT?

	http.Handle("/import/{variant}", authedProxied)         // TODO: GET import is here
	http.Handle("/new/{variant}", authedProxied)            // TODO: GET new item is here
	http.Handle("/view/{variant}/{entryId}", authedProxied) // TODO: GET view item is here
	http.Handle("/list/{variant}", authedProxied)
	http.Handle("/error/{errTxt}", webProxyHandler) // TODO: rate limit???? ctx middleware? auth middleware?
	http.Handle("/testpage", webProxyHandler)       // TODO: REMOVE

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
	//http.Handle("/db/get/rfid/{id}", getRfidHandler())             // TODO: GET RID OF???             // TODO: ensure this works for base58s
	http.Handle("/db/get/{variant}/{id}", rateLimiter(ctxInternalAuthMiddleware(getAnyCollectionHandler()))) // TODO: GET RID OF??? // TODO: make this work for base58 mains as well
	http.Handle(fmt.Sprintf(`%s{%s...}`, imagesEndpoint, imageSubPathKey), ctxInternalAuthMiddleware(picPathMiddleware(getImageHandler())))
	// Creation handlers
	http.Handle("/db/create/{variant}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.HandleCreate()))))
	// update handlers
	http.Handle("/db/update/{endpt}/{id}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.UpdateById())))) // TODO: THIS SHOULD BE PATCH REQUEST?
	// import handlers
	http.Handle("/db/import/{endpt}", rateLimiter(ctxInternalAuthMiddleware(rfidMiddleware(rfid.ImportHandler()))))
	// List handlers
	http.Handle("/db/list/{variant}", ctxInternalAuthMiddleware(rfid.ListEntriesHandler())) // TODO: needs fixing
	//http.Handle("/sessionUserProjects", ctxInternalAuthMiddleware(rfid.SessionUserProjectsHandler())) // TODO: GetPermsMiddleware?
	//http.Handle("/userIdFor", ctxInternalAuthMiddleware(rfid.UserIdForNameOrEmail()))                 // TODO: GetPermsMiddleware?

	println("Defining simple api endpoints")
	http.Handle("/options/{optionsType}", rfid.GetOptionsHandler) // TODO: any options here?

	// lastN handlers
	//http.Handle("/db/list/latest/{variant}", rfid.ListNewestEntriesHandler()) // TODO: maybe unnecessary?
	// listAllStandard handlers
	//http.Handle("/db/list/standard/{variant}", rfid.ListStandardEntriesHandler()) // TODO: maybe unnecessary?
	if err = srv.ListenAndServe(); err != nil {
		panic("failed to listen and serve for http: " + err.Error())
	}
	if err != nil {
		panic("ERROR CLOSING SERVER " + err.Error())
	}
}

func ReqTrackingMiddleWare(handler http.Handler) http.Handler {
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

func NewResponseWriterWrapper(w http.ResponseWriter, onRespond func()) WrappedResponseWriter {
	return WrappedResponseWriter{
		TwoHundredOrAboveHeaderWritten: false,
		internal:                       w,
		onRespond:                      onRespond,
	}
}

type WrappedResponseWriter struct {
	TwoHundredOrAboveHeaderWritten bool
	internal                       http.ResponseWriter
	onRespond                      func()
}

func (w WrappedResponseWriter) Header() http.Header {
	return w.internal.Header()
}
func (w WrappedResponseWriter) Write(bs []byte) (int, error) {
	// Will call WriteHeader(200)before bytes are written
	if !w.TwoHundredOrAboveHeaderWritten {
		w.onRespond()
		w.TwoHundredOrAboveHeaderWritten = true
	}

	return w.internal.Write(bs)
}
func (w WrappedResponseWriter) WriteHeader(statusCode int) {
	if w.TwoHundredOrAboveHeaderWritten {
		return
	}
	if statusCode >= 200 {
		w.onRespond()
		w.TwoHundredOrAboveHeaderWritten = true
	}
	w.internal.WriteHeader(statusCode) // TODO: handle 1XX/2XX/???
}

func setupMiddlewares(ctxIn context.Context, mgr *websocketSessions.SessionManager, loginPath, dbUser, dbPass string) (
	ctx context.Context,
	rateLimiter, rfidMiddleware, picsPathMiddleware, webAuthMiddleware, internalAuthMiddleware func(http.Handler) http.Handler,
	ctxInternalAuthMiddleware, ctxMiddleware, ctxRfidMiddleware func(http.Handler) http.HandlerFunc,
	err error) {
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

	rfidMiddleware = mgr.Middleware()
	// Pics path middleware
	picsPath := os.Getenv("PICS_PATH")
	if picsPath == "" {
		panic("env var missing for PICS_PATH")
	}
	picsPathMiddleware = setupFilePathMiddleware(picsPath)
	// Auth Middleware
	adminUsername := os.Getenv("ADMIN_USER")
	if adminUsername == "" {
		panic("ADMIN_USER env var not set")
	}
	adminPassword := os.Getenv("ADMIN_PASS")
	if adminPassword == "" {
		panic("ADMIN_PASS env var not set")
	}
	webAuthMiddleware, internalAuthMiddleware = placeholderMiddleware, placeholderMiddleware // TODO: what is internal even doing???
	// TODO: DO THESE NEED ANY AUTHENTICATION?
	//ctx = rfid.SetupAuthenticatorOnContext(ctx, utils.Pointer(time.Minute*5), utils.Pointer(time.Hour*2), dbUser, dbPass)
	//svc, _ := rfid.GetAuthService(ctx)                                               // TODO: reenable
	//webAuthMiddleware := svc.AuthOrRedirectMiddleware(loginPath, dbUser, dbPass) // TODO: reenable
	//internalAuthMiddleware := svc.AuthOrDenyMiddleware(dbUser, dbPass) // TODO: reenable
	ctxMiddleware = func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx)) // TODO: ok?
		}
	}
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

type loginRequest struct {
	Username   string `json:"username"`
	HashedPass string `json:"hashedPass"`
}

func handleLogin(viewHandler http.HandlerFunc, rootUser, rootPass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			viewHandler.ServeHTTP(w, r)
		case http.MethodPost:
			time.Sleep(3 * time.Second) // TODO: Make user wait for login, lower likelihood of attack
			//viewHandler.ServeHTTP(w, r) // TODO: is this ok?
			defer r.Body.Close()
			bs, err := io.ReadAll(r.Body)
			if err != nil {
				// TODO: something here
			}
			var userInfo loginRequest
			err = json.Unmarshal(bs, &userInfo)
			if err != nil {
				// TODO: SOMETHING HERE
			}
			var sessionId rfid.SessionId
			// If user is rootUser, try to login as root admin
			if userInfo.Username == rootUser {
				// TODO: reenable
				//hashedRootPass, err := rfid.HashPassword("", rootPass)
				//if err != nil {
				//	http.Error(w, "invalid login credentials: "+err.Error(), http.StatusInternalServerError)
				//	return
				//}
				//if hashedRootPass == userInfo.HashedPass {
				//	// Login as root
				//	// TODO: reenable
				//	//sessionId, err = rfid.LoginRoot(r.Context())
				//	sessionId = "testRootSessionId"
				//} else {
				//	err = errors.New("invalid credentials")
				//}
				sessionId = "testRootSessionId" // TODO: del
			} else {
				// TODO: reenable
				//sessionId, err = rfid.LoginUserPass(r.Context(), userInfo.Username, userInfo.HashedPass)
				sessionId = "testSessionId"
			}
			if err != nil {
				http.Error(w, "invalid login credentials: "+err.Error(), http.StatusForbidden)
				return // TODO: handle this
			}
			// TODO: HANDLE GOOGLE
			// RETURN NEW SESSION ID AS TEXT RESPONSE WITH 201 status!
			http.SetCookie(w, &http.Cookie{
				Name:        "",
				Value:       "",
				Quoted:      false,
				Path:        "",
				Domain:      "",
				Expires:     time.Time{},
				RawExpires:  "",
				MaxAge:      0,
				Secure:      false,
				HttpOnly:    false,
				SameSite:    0,
				Partitioned: false,
				Raw:         "",
				Unparsed:    nil,
			})
			w.WriteHeader(http.StatusAccepted)
			_, err = w.Write([]byte(sessionId))
			if err != nil {
				println("failed to write response: " + err.Error())
			}
		default:
			http.Error(w, "Unsupported http request method: "+http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

	})
}

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
		for key, vals := range r.Header {
			req.Header.Set(key, vals[0])
		}
		// Set session header if exists
		// TODO: reenable
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
		out, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := "Failed to read from http " + err.Error()
			println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		if _, err = w.Write(out); err != nil {
			println("failed to write response")
		}
	}
}

func placeholderMiddleware(next http.Handler) http.Handler { // TODO: DELETEME!!!!!!!
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		// TODO: reenable
		//next.ServeHTTP(w, r.WithContext(rfid.SetAuthInfo(r.Context(), rfid.AuthInfo{
		//	UserId: rfid.AlternateCollectionId(primitive.ObjectID{}),
		//	Opts: &rfid.UserPermsResolved{
		//		Admin:    utils.Pointer(true),
		//		Projects: nil,
		//	},
		//})))
	})
}

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
		bytes, err := os.ReadFile(filepath.Join(pics.GetFilePath(ctx), imgSubPath))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "image not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error retrieving image. "+err.Error(), http.StatusInternalServerError)
			return
		}
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

//	func getRfidHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			ctx := r.Context()
//			id := r.PathValue("id")
//			idBytes := []byte(id)
//			if len(idBytes) != rfid.RfidByteSize {
//				http.Error(w, "invalid id format. Must be 8 bytes", http.StatusBadRequest)
//				return
//			}
//			item, err := rfid.GetMainCollectionItem(ctx, [rfid.RfidByteSize]byte(idBytes), nil) // TODO: WONT WORK
//			if err != nil {
//				if errors.Is(err, mongo.ErrNoDocuments) {
//					http.Error(w, "not found", http.StatusNotFound)
//					return
//				}
//				http.Error(w, "failed to retrieve item", http.StatusInternalServerError)
//				return
//			}
//			out, err := json.Marshal(item)
//			if err != nil {
//				http.Error(w, "failed to marshal item", http.StatusInternalServerError)
//				return
//			}
//
//			_, err = w.Write(out)
//			if err != nil {
//				rfid.HandleHttpWriteError(err)
//			}
//		})
//	}
type clientServerStringConverter struct {
	toClient, toServer func(client string) utils.ErrAnd[string]
}

var base58Converter = clientServerStringConverter{
	toClient: func(binaryIdStr string) utils.ErrAnd[string] {
		out, err := rfid.Base2BytesToBase58([]byte(binaryIdStr))
		if err != nil {
			return utils.TandErr("", err)
		}
		return utils.ErrAndT(string(out))
	},
	toServer: func(b58id string) utils.ErrAnd[string] {
		bs, err := rfid.Base58Str(b58id).Base2Bytes()
		if err != nil {
			return utils.TandErr("", err)
		}
		return utils.ErrAndT(string(bs))
	},
}
var spacedNameConverter = clientServerStringConverter{
	toClient: func(spaced string) utils.ErrAnd[string] {
		// Take spaces and make them underscores?
		return utils.ErrAndT(strings.ReplaceAll(spaced, " ", "_"))
	},
	toServer: func(underscored string) utils.ErrAnd[string] {
		// replace underscores with spaces
		return utils.ErrAndT(strings.ReplaceAll(underscored, "_", " "))
	},
}

type itemTypeWithIdConverter struct {
	item      rfid.AltCollectionItem
	converter clientServerStringConverter
}

func getAnyCollectionHandler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		//authinfo, err := rfid.GetAuthInfo(ctx)
		//if err != nil {
		//	http.Error(w, "Failed to get authinfo: "+err.Error(), http.StatusInternalServerError)
		//	return
		//}
		id := strings.ReplaceAll(r.PathValue("id"), "_", " ") // TODO: replace all underscores with spaces, for things like "chicken of the woods"
		entryType := r.PathValue("variant")
		var bytes []byte
		if itemType, exists := map[string]itemTypeWithIdConverter{
			"agarBatch":       {rfid.AgarBatch{}, base58Converter},
			"agarRecipe":      {rfid.AgarRecipe{}, base58Converter}, // TODO: what about search by name?
			"fruit":           {rfid.Fruit{}, base58Converter},
			"jarRecipe":       {rfid.JarRecipe{}, base58Converter}, // TODO: what about search by name?
			"lcRecipe":        {rfid.LcRecipe{}, base58Converter},  // TODO: what about search by name?
			"pcRun":           {rfid.PCRun{}, base58Converter},
			"project":         {rfid.Project{}, spacedNameConverter},
			"sale":            {rfid.Sale{}, base58Converter},
			"species":         {rfid.Species{}, spacedNameConverter}, // TODO: search by other names?
			"sporePrint":      {rfid.SporePrint{}, base58Converter},
			"subspecies":      {rfid.Subspecies{}, spacedNameConverter},  // TODO: search by other names?
			"substrateRecipe": {rfid.SubstrateRecipe{}, base58Converter}, // TODO: what about search by name?
			"transfer":        {rfid.Transfer{}, base58Converter},
			// TODO: reenable "user":            {rfid.User{}, base58Converter}, // TODO: MAKE SURE USER IS ONLY GOTTEN BY ADMIN // TODO: search by other things than id?
		}[entryType]; exists {
			serverFormattedId := itemType.converter.toServer(id)
			if serverFormattedId.Err != nil {
				http.Error(w, "failed to convert id: "+serverFormattedId.Err.Error(), http.StatusBadRequest)
				return
			}
			out, err := rfid.GetAltCollectionItem(ctx, rfid.AlternateCollectionId([]byte(serverFormattedId.Item)), itemType.item) // TODO: fix
			if err != nil {
				http.Error(w, "failed to get alt collection itemType: "+err.Error(), http.StatusInternalServerError)
				return
			}
			bytes, err = json.Marshal(out)
			if err != nil {
				http.Error(w, "failed to marshal itemType", http.StatusInternalServerError)
				return
			}
		} else {
			if mainCollItem, exists := map[string]rfid.MainCollectionItem{
				"bag":             rfid.Bag{},             // can only go to fruits
				"fruitingChamber": rfid.FruitingChamber{}, // can only go to fruits
				"jar":             rfid.GrainJar{},        // can go anywhere (in theory) except MSS
				"lc":              rfid.LiquidCulture{},   // can go anywhere (in theory) except MSS
				"mss":             rfid.MSS{},             // generally only goes to plate
				"plate":           rfid.Plate{},           // can go anywhere (in theory) except MSS
				"slant":           rfid.Slant{},           // generally only goes to plate
				"stasis":          rfid.StasisTube{},      // generally only goes to plate
			}[entryType]; exists {
				// ensure id is in correct format
				mainCollId, err := rfid.StandardizeMainCollectionId(id)
				if err != nil {
					http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
					return
				}
				out, err := rfid.GetMainCollectionItem(ctx, *mainCollId, mainCollItem)
				if err != nil {
					http.Error(w, "failed to get main collection itemType: "+err.Error(), http.StatusInternalServerError)
					return
				}
				// TODO: reenable
				//if out.Permissions().PermissionFor(authinfo) == perms.None {
				//	http.Error(w, "no perms on mainCollItem for user", http.StatusForbidden)
				//	return
				//}
				bytes, err = json.Marshal(out)
				if err != nil {
					http.Error(w, "failed to marshal itemType: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			// If not a main collection itemType, try for alt
		}
		_, err := w.Write(bytes)
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}
		return
	})
	return rfid.GetPermsMiddleware(handler)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
		//return r.Header.Get("Origin") == "<http://yourdomain.com>" // TODO: this to protect against Cross-Site websocket hijacking (CSWSH)
	},
}

var getRfidReaderNamesHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(mgr.Sessions())
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
	readerName := shared.RfidReaderName(r.PathValue("readerName"))
	binaryUID, err := mgr.ReadRfid(readerName)
	if err != nil {
		// TODO: what type of error?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html") // TODO: content type ok?
	_, err = w.Write(binaryUID[:])
	if err != nil {
		println("failed to write reader result", err)
	}
}

type writeTagRequest struct {
	Secret string
	Data   []byte // TODO: make sure we're ok with this being []byte
}

// TODO: consider moving to reader/internal side?
var rfidWriteHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	if headers.Get("Content-Type") != "application/json" {
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

	req := writeTagRequest{}
	if err = json.Unmarshal(bodyIn, &req); err != nil {
		http.Error(w, "invalid request body structure", http.StatusBadRequest)
		return
	}
	if len(req.Data) != 8 {
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

	writerName := shared.RfidReaderName(r.PathValue("writerName"))
	toWrite := [8]byte(req.Data)
	if err = mgr.WriteRfid(writerName, toWrite); err != nil {
		// TODO: what type of error?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html") // TODO: ensure caught on other side
	_, err = w.Write(req.Data)                  // TODO: is this still ok if incoming was base58?
	if err != nil {
		println("failed to write internal result", err)
	}
}

func GoogleAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code") // TODO: CHANGE? NOT IN QUERY

	t, err := conf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	client := conf.Client(context.Background(), t)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	var v any // TODO: USE

	// Reading the JSON body using JSON decoder
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urlToRedir, _ := url.QueryUnescape(r.URL.Query().Get("state"))
	// TODO: fix
	//println("redirecting to " + r.Host + urlToRedir)
	//http.Redirect(w, r, r.Host+urlToRedir, http.StatusTemporaryRedirect) // TODO: fix
	println("redirecting to " + urlToRedir)
	http.Redirect(w, r, urlToRedir+"?loggedIn=true", http.StatusTemporaryRedirect) // TODO: fix
}
