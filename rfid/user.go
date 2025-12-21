package rfid

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/mail"
	"slices"
)

// User collection

const userCollName = "users" // TODO: readonly field

var ErrInTxnAlreadyTriedToWrite error = errors.New("transaction failed, response has been written already") // TODO: do already

type User struct {
	AlternateCollectionIdField
	Username   string     `bson:"username" json:"username"` // TODO: INDEX? MUST BE UNIQUE
	Email      string     `bson:"email" json:"email"`       // TODO: INDEX? MUST BE UNIQUE
	HashedPass *string    `bson:"password,omitempty" json:"password,omitempty"`
	Salt       *string    `bson:"salt,omitempty" json:"salt,omitempty"`
	GoogleId   *string    `bson:"googleId,omitempty" json:"googleId,omitempty"` // TODO: INDEX?
	Perms      *UserPerms `bson:"perms,omitempty" json:"perms,omitempty"`       // TODO: PROJECTS COME FROM PERMS
	// All can view?
	// TODO: GET ID TOKEN FROM USER, THEN VERIFY IT TO GET ACTUAL GOOGLE ID:
	// TODO: GET TOKEN: https://developers.google.com/identity/sign-in/web/sign-in
	// TODO: VERIFY TOKEN https://developers.google.com/identity/sign-in/web/backend-auth
	// TODO: more (google email, TOTP seed, etc)
}

// TODO: describe
func (u User) CleanBytes() ([]byte, error) {
	return json.Marshal(User{
		AlternateCollectionIdField: u.AlternateCollectionIdField,
		Username:                   u.Username,
		Email:                      u.Email,
		Perms:                      u.Perms, // TODO: ok?
	})
}

func (u User) EntryTypeField() *string {
	return nil
}

func (u User) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := u
	err := decodeItem(&out, encoded)
	return out, err
}

// TODO: USE THIS LATE IN THE SETUP!
func initializeUsers(ctx context.Context, usern, unhashedPass string) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(userCollName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("username", "username", false, false, true), // TODO: is true ok?
		newSimpleIndex("email", "email", false, false, true),       // TODO: is true ok?
		newSimpleIndex("googleId", "googleIde", false, true, true), // TODO: is true ok?
	})
	if err != nil {
		return err
	}
	rootUserId := altCollIdForint(0)

	root := User{}
	result := coll.FindOne(ctx, bson.M{"_id": rootUserId})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// create root user
			root.Email = "root@root.com"
			root.Username = usern
			salt, errr := generateUserSalt()
			if errr != nil {
				return errr
			}
			root.Salt = &salt
			h := sha256.New()

			if _, err = h.Write([]byte(unhashedPass)); err != nil {
				return err
			}
			onceHashedPass := string(h.Sum(nil))
			finalPassHash, err := HashPassword(salt, onceHashedPass)
			if err != nil {
				return err
			}
			root.HashedPass = &finalPassHash
			root.Perms = &UserPerms{
				Admin:    utils.Pointer(true), // TODO: instead of using this, can admins just have a nil perms field?
				Projects: nil,
			}
			// Insert root user to db
			res, err := coll.InsertOne(ctx, root)
			if err != nil {
				return err
			}
			if res.InsertedID.(AlternateCollectionId).asBase58() != rootUserId.asBase58() {
				// TODO: REMOVE FROM DB?
				return errors.New("inserted root user id did not match expected id")
			}
		} else {
			return result.Err()
		}
	}
	var actualRootUser User
	err = result.Decode(&actualRootUser)
	if err != nil {
		return errors.Join(errors.New("failed to decode current root user"), err)
	}
	if actualRootUser.Username != usern {
		return errors.New("current root user id and expected do not match")
	}
	if actualRootUser.Id.asBase58() != rootUserId.asBase58() {
		return errors.New("inserted root id did not match expected")
	}
	return nil
}

// GenerateSalt creates a random salt of the specified length.
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func generateUserSalt() (string, error) {
	bs, err := generateSalt(64)
	return string(bs), err
}

func HashPassword(salt, pw string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(salt + pw))
	if err != nil {
		return "", errors.New("unable to hash password. Should never happen")
	}
	return string(h.Sum(nil)), nil
}

func (u User) validatePassword(pw string) error { // TODO: USE
	if u.Salt == nil {
		return errors.New("salt is required")
	}
	if u.HashedPass == nil {
		return errors.New("hashed password is required")
	}
	toCheckAgainstHashed, err := HashPassword(*u.Salt, pw)
	if err != nil {
		return err
	}
	if toCheckAgainstHashed != *u.HashedPass {
		return errors.New("hashed password does not match")
	}
	return nil
}

func (u User) CollectionName() string {
	return userCollName
}

//func idFieldFor(item any) string { // use THIS ELSEWHERE IF NEEDED
//	itemType := reflect.TypeOf(item)
//	for i := 0; i < itemType.NumField(); i++ {
//		field := itemType.Field(i)
//		tag := field.Tag.Get("bson")
//		if tag == "_id" {
//			return field.Tag.Get("json")
//		}
//	}
//	return ""
//}

//type createNormalUserRequest struct {
//	Name          string
//	Pass          string // validate correct complexity
//	Admin         *bool  // ok?
//	UpdateEntries *bool
//}

// FIX SO THE ADMIN DOESNT KNOW THE USER'S PASSWORD!
//func createNormalUserHandler(w http.ResponseWriter, r *http.Request) { // USE BEHIND AUTH HANDLER!
//	// ENSURE THIS GOES OVER HTTPS!
//	ctx := r.Context()
//	defer r.Body.Close()
//	bs, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "unable to read request body", http.StatusBadRequest)
//		return
//	}
//	userInfo := GetAuthInfo(ctx)
//	if !userInfo.Opts.Contains(AdminKey) {
//		http.Error(w, "you do not have permission to create accounts", http.StatusForbidden)
//		return
//	}
//	req := createNormalUserRequest{}
//	err = json.Unmarshal(bs, &req)
//	if err != nil {
//		http.Error(w, "unable to parse request body", http.StatusBadRequest)
//		return
//	}
//	// Create user
//	salt, err := generateUserSalt()
//	if err != nil {
//		http.Error(w, "unable to generate salt", http.StatusInternalServerError)
//		return
//	}
//	hashedPass, err := HashPassword(salt, req.Pass)
//	if err != nil {
//		http.Error(w, "unable to generate hash", http.StatusInternalServerError)
//		return
//	}
//	newUser := User{
//		Id:         newAlternateCollectionId(),
//		Username:   req.Name,
//		HashedPass: &hashedPass,
//		Salt:       &salt,
//	}
//	if req.Admin != nil && *req.Admin == true {
//		newUser.Admin = req.Admin
//	}
//	if req.UpdateEntries != nil && *req.UpdateEntries == true {
//		newUser.UpdateEntries = req.UpdateEntries
//	}
//	_, err = ctx.Value(mongoClientContextKey).(*mongo.Client).
//		Database(dbName).
//		Collection(userCollName).
//		InsertOne(ctx, newUser)
//	if err != nil {
//		http.Error(w, "unable to create user", http.StatusInternalServerError)
//		return
//	}
//}

func (u User) ResolvePerms(ctx context.Context) (*UserPermsResolved, error) {
	var admin *bool = nil
	if u.Perms != nil {
		admin = u.Perms.Admin
	} else {
		admin = utils.Pointer(false) // TODO: do we want nil to be admin?
	}
	out := UserPermsResolved{
		Admin:    admin,
		Projects: nil,
	}
	if u.Perms == nil || (admin != nil && *admin) {
		return &out, nil
	}
	// Resolve project perms // TODO: MAKE SURE THIS WORKS
	cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(projectsCollectionName).
		Find(ctx, bson.M{"_id": bson.M{"$in": u.Perms.Projects}})
	if err != nil {
		return nil, errors.Join(errors.New("failed to get cursor for UserPerms projects"), err)
	}
	out.Projects = map[projectName]bool{}
	for cursor.Next(context.TODO()) {
		var project Project
		if err = cursor.Decode(&project); err != nil {
			return nil, errors.Join(errors.New("cursor decode error for UserPerms project"), err)
		}
		userIndex := slices.Index(sliceutils.Map(project.Perms.Users.Ids, func(ppuid ProjectPermUserId) AlternateCollectionId {
			return ppuid.Id
		}), u.Id)
		if userIndex != -1 {
			out.Projects[project.Name] = project.Perms.Users.CanWrite[userIndex]
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	return &out, nil
}

func (u User) humanReadableId() string { // TODO: ok?
	if u.GoogleId != nil {
		return u.Email
	}
	return u.Username
}

type UserPerms struct {
	Admin    *bool         `bson:"admin,omitempty" json:"admin,omitempty"`
	Projects []projectName `bson:"projects,omitempty" json:"projects,omitempty"`
}

// UserPermsResolved is the cached version of userPerms that has the read/write privileges from each project
type UserPermsResolved struct { // TODO: should we hold these in a cache?
	Admin    *bool
	Projects map[projectName]bool
}

// getUserPermsForRequest tries to get AuthInfo off of the request's context,
// otherwise it tries to resolve those perms and place them on the context to be returned
func getUserPermsForRequest(r *http.Request) (ctx context.Context, auth AuthInfo, err error) {
	ctx = r.Context()
	authInfo, err := GetAuthInfo(ctx)
	if err != nil {
		return getCookieAuthInfo(r)
	}
	return ctx, authInfo, nil
}

func GetPermsMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, _, err := getUserPermsForRequest(r)
		if err != nil {
			http.Error(w, "failed to get perms for request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getCookieAuthInfo(r *http.Request) (ctx context.Context, auth AuthInfo, err error) {
	cookie, err := GetSessionCookie(r)
	if err != nil {
		return ctx, AuthInfo{}, errors.Join(errors.New("no session on request"), err)
	}
	ctx = r.Context()
	session := SessionId(cookie.Value)
	svc, err := GetAuthService(ctx)
	if err != nil {
		return ctx, auth, err
	}
	sess, err := svc.TryToReAuth(session)
	if err != nil {
		return ctx, auth, err
	}
	ctx = SetAuthInfo(r.Context(), sess.Data)
	return ctx, sess.Data, nil
}

func LoginUserPass(ctx context.Context, username string, hashedPass string) (SessionId, error) {
	svc, err := GetAuthService(ctx)
	if err != nil {
		return "", err // TODO: ok?
	}
	sessId, _, err := svc.TryToAuthUserPass(ctx, username, hashedPass)
	return sessId, err
}

// TODO: root is all-powerful admin
func LoginRoot(ctx context.Context) (SessionId, error) {
	svc, err := GetAuthService(ctx)
	if err != nil {
		return "", err // TODO: ok?
	}
	sessId, err := generateSessionId()
	if err != nil {
		return sessId, err
	}
	_, err = svc.SessionToAuthMap.NewSession(sessId, AuthInfo{
		Id: altCollIdForint(0), // TODO: FIX ME?
		Opts: &UserPermsResolved{
			Admin:    utils.Pointer(true),
			Projects: nil,
		},
	})
	return sessId, err
}

// TODO: USE THIS
// TODO: return user ID?
func CreateUser(ctx context.Context, email string, username string, hashedPass, googleId *string) error {
	user := User{
		AlternateCollectionIdField: newAlternateCollectionId().asIdField(),
		Username:                   username,
		Email:                      email,
		GoogleId:                   googleId,
		Perms:                      nil,
	}
	// User/pw stuff
	if hashedPass != nil {
		salt, err := generateUserSalt()
		if err != nil {
			return err
		}
		user.Salt = &salt
		finalHashedPass, err := HashPassword(salt, *hashedPass)
		if err != nil {
			return err
		}
		user.HashedPass = &finalHashedPass
	}
	_, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
		Collection(userCollName).InsertOne(ctx, user)
	return err
}

func UserIdForNameOrEmail() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		valueBs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to get user from request", http.StatusBadRequest)
			return
		}
		usernameOrEmail := string(valueBs) // TODO: get via username or email
		result := getUserIdByNameOrEmail(r.Context(), usernameOrEmail)
		if result.Err != nil {
			http.Error(w, result.Err.Error(), http.StatusNotFound)
			return
		}
		bs, err := json.Marshal(result.Item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = w.Write(bs)
		if err != nil {
			HandleHttpWriteError(err)
		}
	})
}

func getUserIdByNameOrEmail(ctx context.Context, nameOrEmail string) utils.ErrAnd[AlternateCollectionId] {
	var filter bson.M
	_, err := mail.ParseAddress(nameOrEmail)
	isEmail := err == nil
	if isEmail {
		filter = bson.M{"email": nameOrEmail}
	} else {
		filter = bson.M{"username": nameOrEmail}
	}
	var u User
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(userCollName).FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return utils.TandErr(AlternateCollectionId{}, err)
	}
	return utils.ErrAndT(u.Id)
}

type updateUserRequest struct {
	Admin *bool `json:"admin,omitempty"`
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) { // TODO: MAKE SURE WE ARE ONLY UPDATING THE ADMIN PART
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateUserRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	authInfo, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get session permissions: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !authInfo.isAdmin() {
		http.Error(w, "only admins can modify user permissions", http.StatusForbidden)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(userCollName) // TODO: FIX EVERYTHING BELOW THIS
		existing, err := GetAltCollectionItemInTxn(ctx, id, User{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return DbTxnStdErr(w, err.Error(), stat)
		}
		doRemoveAdmin := req.Admin == nil || *req.Admin == false
		if existing.Id == authInfo.Id && doRemoveAdmin {
			return DbTxnStdErr(w, "cannot remove admin from self", http.StatusBadRequest)
		}
		mods := bson.D{} // TODO: modify
		if doRemoveAdmin {
			mods = append(mods, bson.E{"$unset", "perms.admin"})
		} else {
			mods = append(mods, bson.E{"$set", bson.D{{"perms.admin", true}}})
		}
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", id}}, mods)
		err = result.Err()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		resultUser, err := GetAltCollectionItemInTxn(ctx, id, User{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return DbTxnStdErr(w, err.Error(), stat)
		}
		userBs, err := resultUser.CleanBytes()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(userBs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}

}
