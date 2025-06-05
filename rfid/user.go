package rfid

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
)

// User collection

const userCollName = "users" // TODO: probably dont need a collection of users

const (
	MaxAuthKey = "ALL"
	ChangeKey  = "updateEntries"
)

type User struct {
	Id         alternateCollectionId `bson:"_id" json:"_id"`                               // TODO: k for main, r for backup (makes some less descriptive)
	Name       string                `bson:"name" json:"name"`                             // TODO: INDEX?
	HashedPass *string               `bson:"password,omitempty" json:"password,omitempty"` // TODO: DONT INDEX!
	Salt       *string               `bson:"salt,omitempty" json:"salt,omitempty"`
	GoogleId   *string               `bson:"googleId,omitempty" json:"googleId,omitempty"` // TODO: INDEX?
	Perms      UserPerms             `bson:"perms" json:"perms"`
	Admin      *bool                 `bson:"admin,omitempty" json:"admin,omitempty"`
	// All can view?
	UpdateEntries *bool `bson:"updateEntries,omitempty" json:"updateEntries,omitempty"`
	// TODO: GET ID TOKEN FROM USER, THEN VERIFY IT TO GET ACTUAL GOOGLE ID:
	// TODO: GET TOKEN: https://developers.google.com/identity/sign-in/web/sign-in
	// TODO: VERIFY TOKEN https://developers.google.com/identity/sign-in/web/backend-auth
	// TODO: more (google email, TOTP seed, etc)
}

func (u User) perms() AuthInfo { // TODO: PUBLIC OR NO?
	idOut := primitive.ObjectID(u.Id)
	if u.Admin != nil && *u.Admin == true {
		return AuthInfo{
			Id:   idOut,
			Opts: UserPerms(utils.Set[string]{MaxAuthKey: {}}),
		}
	}
	// parse normal perms
	perms := utils.Set[string]{}
	if u.UpdateEntries != nil && *u.UpdateEntries == true {
		perms.Add(ChangeKey)
	}
	return AuthInfo{
		Id:   primitive.ObjectID(u.Id),
		Opts: nil,
	}
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
	bs, err := generateSalt(64) // TODO: 64 ok?
	return string(bs), err
}

func hashPassword(salt, pw string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(salt + pw))
	if err != nil {
		return "", errors.New("unable to hash password. Should never happen")
	}
	return string(h.Sum(nil)), nil
}

func (u User) validatePassword(pw string) error { // TODO: DELETE?
	if u.Salt == nil {
		return errors.New("salt is required")
	}
	if u.HashedPass == nil {
		return errors.New("hashed password is required")
	}
	toCheckAgainstHashed, err := hashPassword(*u.Salt, pw)
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

func idFieldFor(item any) string { // TODO: use THIS ELSEWHERE IF NEEDED
	itemType := reflect.TypeOf(item)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		tag := field.Tag.Get("bson")
		if tag == "_id" {
			return field.Tag.Get("json")
		}
	}
	return ""
}

func initializeUsers(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	_, err := db.Collection(userCollName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		// TODO: ensure indices can be re-created (or do nothing when they already exist)
		newSimpleIndex("name", "name", false, false, false),
		newSimpleIndex("googleId", "googleId", false, true, false),
		// TODO: any more?
		//lastUpdatedIndexModel, // TODO: ??????
	})
	return err
	// TODO: anything else in here? Do we want the manual admin account to get added here?
}

type createNormalUserRequest struct {
	Name          string
	Pass          string // TODO: validate correct complexity
	Admin         *bool  // TODO: ok?
	UpdateEntries *bool
}

// TODO: FIX SO THE ADMIN DOESNT KNOW THE USER'S PASSWORD!
func createNormalUserHandler(w http.ResponseWriter, r *http.Request) { // TODO: USE BEHIND AUTH HANDLER!
	// TODO: ENSURE THIS GOES OVER HTTPS!
	ctx := r.Context()
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	userInfo := GetAuthInfo(ctx)
	if !userInfo.Opts.Contains(MaxAuthKey) {
		http.Error(w, "you do not have permission to create accounts", http.StatusForbidden)
		return
	}
	req := createNormalUserRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "unable to parse request body", http.StatusBadRequest)
		return
	}
	// Create user
	salt, err := generateUserSalt()
	if err != nil {
		http.Error(w, "unable to generate salt", http.StatusInternalServerError)
		return
	}
	hashedPass, err := hashPassword(salt, req.Pass)
	if err != nil {
		http.Error(w, "unable to generate hash", http.StatusInternalServerError)
		return
	}
	newUser := User{
		Id:         newAlternateCollectionId(),
		Name:       req.Name,
		HashedPass: &hashedPass,
		Salt:       &salt,
	}
	if req.Admin != nil && *req.Admin == true {
		newUser.Admin = req.Admin
	}
	if req.UpdateEntries != nil && *req.UpdateEntries == true {
		newUser.UpdateEntries = req.UpdateEntries
	}
	_, err = ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(userCollName).
		InsertOne(ctx, newUser)
	if err != nil {
		http.Error(w, "unable to create user", http.StatusInternalServerError)
		return
	}
}
