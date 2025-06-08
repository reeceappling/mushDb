package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/reeceappling/mushDb/rfid"
	"github.com/reeceappling/mushDb/rfid/pics"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	dbHostPort, err := strconv.Atoi(os.Getenv("DB_HOST_PORT"))
	if err != nil {
		println(errors.Join(errors.New("no db port configured on env var DB_HOST_PORT"), err).Error())
		dbHostPort = 0
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
	println("Initializing DB") // TODO: ok?
	if err = rfid.Initialize(ctx); err != nil {
		return ctx, nil, errors.Join(errors.New("failed to initialize database"), err)
	}
	println("DB setup and connection complete!")
	return ctx, rfid.GetMongoClient(ctx), nil
}

var conf *oauth2.Config

func main() {
	var err error
	ctx := context.Background()

	// TODO: setup logger

	// Get non-db env vars
	clusterSecret := os.Getenv("RFID_SECRET")
	if clusterSecret == "" {
		panic("env var missing for RFID_SECRET")
	}
	conf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),         // TODO: CONFIGURE IN HELM
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),     // TODO: CONFIGURE IN  HELM
		RedirectURL:  "http://localhost:3728/auth/callback", // TODO: fixme?
		Scopes:       []string{"email", "profile"},          // TODO: one more?
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

	webHostName := envVarOrDefault("WEB_HOST_INTERNAL", "localhost") // Can have port if not hosting on 80      // TODO: CONFIGURE
	ingressPort, err := strconv.Atoi(os.Getenv("API_PORT"))
	if err != nil {
		println("No api port configured, defaulting to port 80")
		ingressPort = 80
	}

	// Set up server
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(ingressPort),
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
	cleanupFreq := 2 * time.Minute
	mgr := websocketSessions.NewSessionManager(&cleanupFreq, clusterSecret)
	defer mgr.Cleanup()
	rfidMiddleware := mgr.Middleware()
	picPathMiddleware := setupFilePathMiddleware("/pics") // TODO: ensure matches k8s
	// SETUP AUTH MIDDLEWARE!!!!!
	webAuthMiddleware, internalAuthMiddleware := placeholderMiddleware, placeholderMiddleware // TODO: what is internal even doing???
	ctxMiddleware := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx)) // TODO: ok?
		}
	}

	//const loginPath = "/login" // TODO: REENABLE
	//ctx = rfid.SetupAuthenticatorOnContext(ctx, utils.Pointer(time.Minute*5), utils.Pointer(time.Hour*2), dbUser, dbPass)
	//svc, _ := rfid.GetAuthService(ctx) // TODO: reenable
	//webAuthMiddleware := svc.AuthOrRedirectMiddleware(loginPath, dbUser, dbPass) // TODO: reenable
	//internalAuthMiddleware := svc.AuthOrDenyMiddleware(dbUser, dbPass) // TODO: reenable

	// RFID HANDLERS

	println("Defining endpoints")
	// TODO: DO THESE NEED ANY AUTHENTICATION?
	ctxRfidMiddleware := func(next http.Handler) http.HandlerFunc {
		return ctxMiddleware(rfidMiddleware(next))
	}
	// Must be publicly available?
	http.HandleFunc("/rfid/ws", ctxRfidMiddleware(wsServerHandler))
	// Can be internal to docker network
	http.HandleFunc("/rfid/read/{readerName}", ctxRfidMiddleware(rfidReadHandler))   // TODO: OUTPUT IS BASE 2! // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/write/{writerName}", ctxRfidMiddleware(rfidWriteHandler)) // TODO: INPUT IS BASE58(?) OUTPUT IS BASE 2! // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)
	http.HandleFunc("/rfid/readers", ctxRfidMiddleware(getRfidReaderNamesHandler))   // TODO: DONT ALLOW USERS TO DIRECTLY HIT THIS (req should come from webserver)

	// SERVER HANDLERS! (PASSTHROUGH) view, new, import
	passthroughConfig := newPassthroughHandlerConfig().
		useHttps(false).
		withHost(webHostName).
		withCookies(true).
		withHeaders(true)
	webHostPortStr := os.Getenv("WEB_HOST_INTERNAL_PORT") // TODO: CONFIGURE
	webHostPort := 3000
	if webHostPortStr != "" {
		webHostPort, err = strconv.Atoi(webHostPortStr)
		if err != nil {
			panic("invalid internal web host port: " + webHostPortStr)
		}
		passthroughConfig.withPort(webHostPort)
	} else {
		println("No/invalid web host port specified, defaulting to 3000")
	}

	http.Handle("/", rootHandler)
	// TODO: maybe create a more readable middleware setup???

	webProxyHandler := newPassthroughHandler(passthroughConfig)
	http.Handle("/login", ctxMiddleware(handleLogin(webProxyHandler))) // GET=view, POST=do
	http.Handle("/_next", ctxMiddleware(webProxyHandler))              // TODO: CHANGE ROOT?
	http.Handle("/", ctxMiddleware(webProxyHandler))                   // TODO: CHANGE ROOT?
	ctxWebAuthMiddleware := func(next http.Handler) http.HandlerFunc {
		return ctxMiddleware(webAuthMiddleware(next))
	}
	http.Handle("/import/{variant}", ctxWebAuthMiddleware(webProxyHandler))
	http.Handle("/new/{variant}", ctxWebAuthMiddleware(webProxyHandler))
	http.Handle("/view/{variant}/{entryId}", ctxWebAuthMiddleware(webProxyHandler))

	ctxInternalAuthMiddleware := func(next http.Handler) http.HandlerFunc {
		return ctxMiddleware(internalAuthMiddleware(next))
	}
	//http.Handle("/db/get/rfid/{id}", getRfidHandler())             // TODO: GET RID OF???             // TODO: ensure this works for base58s
	http.Handle("/db/get/{endpt}/{id}", ctxInternalAuthMiddleware(getAnyCollectionHandler())) // TODO: GET RID OF??? // TODO: make this work for base58 mains as well
	http.Handle(fmt.Sprintf(`%s{%s...}`, imagesEndpoint, imageSubPathKey), ctxInternalAuthMiddleware(picPathMiddleware(getImageHandler())))
	// Creation handlers
	http.Handle("/db/create/{variant}", ctxInternalAuthMiddleware(rfidMiddleware(rfid.HandleCreate())))
	// update handlers
	http.Handle("/db/update/{endpt}/{id}", ctxInternalAuthMiddleware(rfidMiddleware(rfid.UpdateById())))
	// import handlers
	http.Handle("/db/import/{endpt}", ctxInternalAuthMiddleware(rfidMiddleware(rfid.ImportHandler())))
	// List handlers
	http.Handle("/db/list/{variant}", ctxInternalAuthMiddleware(rfid.ListEntriesHandler())) // TODO: needs fixing

	// lastN handlers
	//http.Handle("/db/list/latest/{variant}", rfid.ListNewestEntriesHandler()) // TODO: maybe unnecessary?
	// listAllStandard handlers
	//http.Handle("/db/list/standard/{variant}", rfid.ListStandardEntriesHandler()) // TODO: maybe unnecessary?
	if err = srv.ListenAndServe(); err != nil {
		panic("failed to listen and serve for http: " + err.Error())
	}
	//println("Listening on port " + strconv.Itoa(ingressPort))
	//if doTLS {
	//	if err = srv.ListenAndServeTLS("/tls/"+certFileName, "/tls/"+keyFileName); err != nil {
	//		panic("failed to listen and serve TLS: " + err.Error())
	//	}
	//} else {
	//	if certFileName+keyFileName != "" {
	//		println("certFile or keyFile exists for TLS, but not both! Falling back to http (but still port 443)") // TODO: ok?
	//	}
	//
	//}
	if err != nil {
		panic("ERROR CLOSING SERVER " + err.Error())
	}
}

func envVarOrDefault(varName, defaultResult string) string {
	result := os.Getenv(varName)
	if result == "" {
		println("env var " + varName + " missing, defaulting to " + defaultResult)
		return defaultResult
	}
	return result
}

func handleLogin(viewHandler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			viewHandler.ServeHTTP(w, r)
		case http.MethodPost:
			// TODO: HANDLE USERNAME/PASSWORD
			// TODO: HANDLE GOOGLE
			// TODO: THIS! USER LOGS IN!
			http.Error(w, "NOT IMPLEMENTED YET IN handleLogin", http.StatusServiceUnavailable)
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
		ctx := r.Context()
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
		// SET AUTH PERMS HEADER IF EXISTS
		perms := rfid.GetAuthInfo(ctx)
		if len(perms.Opts) != 0 {
			req.Header.Set(rfid.AuthPermsContextHeaderKey, perms.Opts.OptsAsString())
		}
		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to do "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// RELAY HEADERS AS NEEDED
		// TODO: REMOVE AUTH PERMS HEADERS
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}

		//http.SetCookie(w, &http.Cookie{
		//	Name:        "ApplingSession",
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
			http.Error(w, "Failed to read from http "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(out)
		if err != nil {
			panic("failed to write response")
		}
	}
}

func placeholderMiddleware(next http.Handler) http.HandlerFunc { // TODO: DELETEME!!!!!!!
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set(rfid.AuthPermsContextHeaderKey, rfid.MaxAuthInfo.Opts.OptsAsString()) // TODO: necessary?
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), rfid.AuthPermsContextHeaderKey, rfid.MaxAuthInfo.Opts.OptsAsString())))
	}
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

var rootHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("an apple a day: from " + r.URL.Path)) //nolint:errcheck // TODO: DO WE EVEN WANT THE ROOT TO RESPOND?
	if err != nil {
		rfid.HandleHttpWriteError(err)
	}
})

//func getRfidHandler() http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		ctx := r.Context()
//		id := r.PathValue("id")
//		idBytes := []byte(id)
//		if len(idBytes) != rfid.RfidByteSize {
//			http.Error(w, "invalid id format. Must be 8 bytes", http.StatusBadRequest)
//			return
//		}
//		item, err := rfid.GetMainCollectionItem(ctx, [rfid.RfidByteSize]byte(idBytes), nil) // TODO: WONT WORK
//		if err != nil {
//			if errors.Is(err, mongo.ErrNoDocuments) {
//				http.Error(w, "not found", http.StatusNotFound)
//				return
//			}
//			http.Error(w, "failed to retrieve item", http.StatusInternalServerError)
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

func getAnyCollectionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := strings.ReplaceAll(r.PathValue("id"), "_", " ") // TODO: replace all underscores with spaces, for things like "chicken of the woods"
		endpt := r.PathValue("endpt")
		var bytes []byte
		if item, exists := map[string]rfid.AltCollectionItem{
			"agarBatch":  rfid.AgarBatch{},
			"agarRecipe": rfid.AgarRecipe{},
			"fruit":      rfid.Fruit{},
			"jarRecipe":  rfid.JarRecipe{},
			"lcRecipe":   rfid.LCRecipe{},
			//"user":"", // TODO: probably don't need
			"pcRun":           rfid.PCRun{},
			"project":         rfid.Project{},
			"sale":            rfid.Sale{},
			"species":         rfid.Species{},
			"sporePrint":      rfid.SporePrint{},
			"subspecies":      rfid.Subspecies{},
			"substrateRecipe": rfid.SubstrateRecipe{},
			"transfer":        rfid.Transfer{},
		}[endpt]; exists {
			out, err := rfid.GetAltCollectionItem(ctx, id, item) // TODO: modify output for special species
			if err != nil {
				http.Error(w, "failed to get alt collection item: "+err.Error(), http.StatusInternalServerError)
				return
			}
			bytes, err = json.Marshal(out)
			if err != nil {
				http.Error(w, "failed to marshal item", http.StatusInternalServerError)
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
			}[endpt]; exists {
				// ensure id is in correct format
				mainCollId, err := rfid.StandardizeMainCollectionId(id)
				if err != nil {
					http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
					return
				}
				out, err := rfid.GetMainCollectionItem(ctx, *mainCollId, mainCollItem)
				if err != nil {
					http.Error(w, "failed to get main collection item: "+err.Error(), http.StatusInternalServerError)
					return
				}
				bytes, err = json.Marshal(out)
				if err != nil {
					http.Error(w, "failed to marshal item: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			// If not a main collection item, try for alt
		}
		_, err := w.Write(bytes)
		if err != nil {
			rfid.HandleHttpWriteError(err)
		}
		return
	})
}

var wsServerHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mgr := websocketSessions.GetSessionManager(r.Context())
	if mgr == nil {
		http.Error(w, websocketSessions.ErrNoSessionManager.Error(), http.StatusInternalServerError)
		return
	}
	conn, errUpgr := upgrader.Upgrade(w, r, nil)
	if errUpgr != nil {
		fmt.Println("Error upgrading connection:", errUpgr)
	}
	defer conn.Close()

	//try to read signup message
	msgType, msgBytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading from websocket:", err)
		return
	}
	if msgType != websocket.BinaryMessage {
		fmt.Println("Error reading from websocket: message type was not binary:", msgType)
		return
	}
	msg := websocketSessions.SocketMessage{}
	if err = json.Unmarshal(msgBytes, &msg); err != nil {
		fmt.Println("Error unmarshalling from websocket:", err)
		return
	}
	if msg.Type != websocketSessions.MessageTypeSignup {
		fmt.Println("Error signing up, initial message received is of wrong type")
		return
	}
	req := &websocketSessions.SignupRequest{}
	err = json.Unmarshal(msg.Data, req)
	if err != nil {
		fmt.Println("Error unmarshalling signup request on server:", err)
		return
	}
	timeBtwnChecks := 30 * time.Second                        // TODO: ok?
	maxFailures := 1                                          // TODO: ok?
	requestTimeout := 10 * time.Second                        // TODO: ok?
	sessionTimeout := 5 * time.Minute                         // TODO: ok?
	ctx, sessionCancelFunc := context.WithCancel(r.Context()) // TODO: ensure this cancelFunc is being used correctly
	newSession := websocketSessions.NewSession(conn, &sessionTimeout, &requestTimeout, &timeBtwnChecks, &maxFailures)
	err = mgr.Add(sessionCancelFunc, newSession, *req)
	if err != nil {
		fmt.Println("Error adding websocket session:", err)
		return
	}
	_ = <-ctx.Done() // Keep request alive until ready to close connection
})

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
	readerName := websocketSessions.RfidReaderName(r.PathValue("readerName"))
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
	Data   []byte
}

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

	writerName := websocketSessions.RfidReaderName(r.PathValue("writerName"))
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

	toWrite := [8]byte(req.Data)

	if err = mgr.WriteRfid(writerName, toWrite); err != nil {
		// TODO: what type of error?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html") // TODO: ensure caught on other side
	_, err = w.Write(req.Data)                  // TODO: is this still ok if incoming was base58?
	if err != nil {
		println("failed to write writer result", err)
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
