package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/maps"
	"strings"
)

type AclField struct {
	ACL ACL `bson:"acl" json:"acl"`
}

func (field AclField) Permissions() ACL {
	return field.ACL
}

func allCanReadAcl(owner *string) AclField {
	out := ACL{
		BlanketPerm: RWPermRead(),
		Users:       map[string]bool{},
		Projects:    map[projectName]bool{},
	}
	if owner != nil {
		out.Users[*owner] = true
	}
	return AclField{ACL: out}
}
func allCanWriteAcl() AclField {
	return AclField{ACL: ACL{
		BlanketPerm: RWPermWrite(),
	}}
}

// ACL being nil means anyone authenticated can do anything (read/write)
type ACL struct { // ALWAYS REFERENCED AS A STRUCT AND NOT A POINTER!
	Users       map[string] /*email*/ bool `bson:"users,omitempty" json:"users,omitempty"`             // bool is canWrite
	Projects    map[projectName]bool       `bson:"projects,omitempty" json:"projects,omitempty"`       // bool is canWrite
	BlanketPerm *ReadWritePerm             `bson:"blanketPerm,omitempty" json:"blanketPerm,omitempty"` // empty is private, false is public can read by default. True means public can write by default.
}

//TODO: consider trying out!
//var updatePolicyCompiled, viewPolicyCompiled, createPolicyCompiled rego.PreparedEvalQuery
//
//func init() {
//	defaultDeny := `# Default decision if no rules match
//default decision := "deny"
//
//# Main decision logic: deny takes absolute precedence
//topArea := `package authz
//`
//decision := "deny" {
//    deny
//}
//else := "allow" {
//    allow
//}
//`
//	helpers := `has_acl {
//		has_key(input, "acl")
//	}
//	has_acl_projects {
//		has_acl
//		has_key(input.acl, "projects")
//	}
//	has_acl_users {
//		has_acl
//		has_key(input.acl, "users")
//	}
//	has_user_projects {
//		has_key(input.user, "projects")
//	}
//	# Helper to check if a key exists in an object
//	has_key(obj, key) {
//		_ = obj[key]
//	}`
//	updatePolicyString := topArea+defaultDeny+`
//# Deny if accountType is missing (guest)
//deny["Guests cannot edit"] {
//   not object.get(input.user, "accountType", false)
//   # Check specifically if the key is absent
//   not has_key(input.user, "accountType")
//}
//# Allow admins
//allow if {
//   input.user.accountType == true
//}
//# Allow if blanketPerm is true
//allow if {
//	has_key(input, "blanketPerm")
//	input.blanketPerm == true
//}
//# allow users where the user is in the acl specifically
//allow if {
//	has_acl_users
//	input.acl.users[input.user.email]
//}
//# allow users who are part of groups that are allowed to write by the acl
//allow if {
//	has_acl_projects
//	has_user_projects
//	userWriteableProjects := {proj | k, v in input.user.projects; v == true; proj := k}
//	projectsThatCanWriteOnItem := {proj | k, v in input.acl.projects; v == true; proj := k}
//	some common_value in userWriteableProjects & projectsThatCanWriteOnItem
//}
//`+helpers
//	viewPolicyString := topArea+defaultDeny+`
//# Allow admins
//allow if {
//   input.user.accountType == true
//}
//# Allow if blanketPerm is non-nil
//allow if {
//	has_key(input, "blanketPerm")
//}
//# allow users where the user is in the acl specifically
//allow if {
//	has_acl_users
//	{k | k, v in input.acl.users} in input.user.email
//}
//# allow users who are part of groups that are allowed to read or write by the acl
//allow if {
//	has_acl_projects
//	has_user_projects
//	userReadableProjects := {proj | k, v in input.user.projects; proj := k}
//	projectsThatCanReadItem := {proj | k, v in input.acl.projects; proj := k}
//	some common_value in userReadableProjects & projectsThatCanReadItem
//	# TODO: also handle cases where the item itself is not innoculated???????0----------------------------
//}
//`+helpers
//	createPolicyString := topArea+defaultDeny+`
//deny["Guests cannot create"] {
//   not object.get(input.user, "accountType", false)
//   # Check specifically if the key is absent
//   not has_key(input.user, "accountType")
//}
//# Allow admins and normal users
//allow if {
//	has_key(input.user, "accountType")
//}
//`+helpers
//	var err error = nil
//	updatePolicyCompiled, err = rego.New(
//		rego.Query("data.authz.decision"),
//		rego.Module("policy.rego", updatePolicyString),
//	).PrepareForEval(context.Background())
//	if err != nil {
//		panic("failed to create update policy compiled query: " + err.Error())
//	}
//	viewPolicyCompiled, err = rego.New(
//		rego.Query("data.authz.decision"),
//		rego.Module("policy.rego", viewPolicyString),
//	).PrepareForEval(context.Background())
//	if err != nil {
//		panic("failed to create view policy compiled query: " + err.Error())
//	}
//	createPolicyCompiled, err = rego.New(
//		rego.Query("data.authz.decision"),
//		rego.Module("policy.rego", createPolicyString),
//	).PrepareForEval(context.Background())
//	if err != nil {
//		panic("failed to create create policy compiled query: " + err.Error())
//	}
//}
//
//func updatePolicy(ctx context.Context, acl *ACL) error { // TODO: USE THIS????!!!!
//	return checkPolicy(ctx, acl, updatePolicyCompiled)
//}
//func viewPolicy(ctx context.Context, acl *ACL) error { // TODO: USE THIS????!!!!
//	return checkPolicy(ctx, acl, viewPolicyCompiled)
//}
//func createPolicy(ctx context.Context, acl *ACL) error { // TODO: USE THIS????!!!!
//	return checkPolicy(ctx, acl, createPolicyCompiled)
//}
//func checkPolicy(ctx context.Context, acl *ACL, compiledPolicy rego.PreparedEvalQuery) error {
//	user, err := GetAuthInfo(ctx)
//	if err != nil {
//		return errors.Join(errors.New("failed to get auth info"), err)
//	}
//	input := rego.EvalInput(map[string]interface{}{ // TODO: ensure inserting these is ok
//		"user": user,
//		"acl":  acl,
//	})
//	results, err := compiledPolicy.Eval(ctx, input)
//	if err != nil {
//		return errors.Join(errors.New("failed to eval"), err)
//	}
//
//	if len(results) == 0 || !results.Allowed() {
//		return errors.New("Access Denied. " + results[0].Expressions[0].String())
//	}
//	return nil
//}

func (acl ACL) AsPermsOnRequest() PermsOnRequest {
	return PermsOnRequest{
		UserPerms:    cloneMap(acl.Users),
		ProjectPerms: cloneMap(acl.Projects),
		BlanketPerm:  acl.BlanketPerm.Copy(),
	}
}

func cloneMap[T comparable, U any](m map[T]U) map[T]U {
	if m == nil {
		return nil
	}
	out := make(map[T]U, len(m))
	for key, val := range m {
		out[key] = val
	}
	return out
}

func (acl ACL) Clone() ACL {
	return ACL{
		Users:       cloneMap(acl.Users),
		Projects:    cloneMap(acl.Projects),
		BlanketPerm: acl.BlanketPerm,
	}
}

// func (acl ACL) MarshalJSON()(bs []byte, err error) is not custom
// UnmarshalJSON unmarshals bytes into an access control list
func (acl *ACL) UnmarshalJSON(bs []byte) (err error) {
	out := ACL{} // TODO: out := &ACL{}
	temp := map[string]any{}
	if err = json.Unmarshal(bs, &temp); err != nil {
		return err
	}
	if usersInterface, ok := temp["users"]; ok {
		out.Users, ok = usersInterface.(map[string]bool)
		if !ok {
			return errors.New("ACL Users is not a map[string]bool or nil")
		}
	}
	if projectsInterface, ok := temp["projects"]; ok {
		out.Projects, ok = projectsInterface.(map[projectName]bool)
		if !ok {
			return errors.New("ACL projects is not a map[string]bool or nil")
		}
	}
	if blanketInterface, exists := temp["blanketPerm"]; exists { // TODO: validate ok! blanket perm may be missing!
		var blanketPermIsValid = false
		*out.BlanketPerm, blanketPermIsValid = blanketInterface.(ReadWritePerm) // TODO: validate ok! blanket perm may be missing!
		if !blanketPermIsValid {
			return errors.New("ACL blanketPerm must be a present boolean field v2")
		}
	} else {
		return errors.New("ACL blanketPerm must be a present boolean field v1") // TODO: maybe that should just make blanketPerm nil?
	}
	//blanketPermIfc, ok := temp["blanketPerm"] // TODO: Reenable if new does not work
	//if !ok {
	//	return errors.New("ACL blanketPerm must be a present boolean field v1")
	//}
	//*out.BlanketPerm, ok = blanketPermIfc.(ReadWritePerm)
	//if !ok {
	//	return errors.New("ACL blanketPerm must be a present boolean field v2")
	//}

	*acl = out.simplified() // TODO: ok to simplify here?
	return nil
}

func (acl ACL) simplified() ACL {
	if acl.BlanketPerm == nil {
		// If admin or default permission is no permission, return self
		return acl
	}
	// TODO: PROBABLY DONT WANT TO REMOVE USERS WHEN BLANKET CHANGES? DECIDE!
	// If blanketPerm is read or write, then remove all users that can only read? // TODO: I DONT LIKE THIS, WHAT IF BLANKET CHANGES LATER?
	if acl.Users != nil { // TODO: necessary? try disabling this
		for user, canWrite := range acl.Users {
			if !canWrite {
				delete(acl.Users, user) // TODO: unsure if I like this...
			}
		}
	}
	if len(acl.Users) == 0 {
		acl.Users = nil // TODO: ensure ok!
	}
	// Do not remove projects because projects need to be associated with the item
	return acl
}
func mapsMatch[A comparable, B comparable](a, b map[A]B) bool { // TODO: test. should be fine but probably needs a test around it.
	if len(a) == 0 {
		return len(b) == 0
	}
	for k, va := range a {
		vb, exists := b[k]
		if !exists || va != vb {
			return false
		}
	}
	return true
}
func (acl ACL) Equivalent(other ACL) bool {
	if acl.BlanketPerm != other.BlanketPerm {
		return false
	}
	return mapsMatch(acl.Users, other.Users) && mapsMatch(acl.Projects, other.Projects)
}

func (acl ACL) AsField() AclField {
	return AclField{ACL: acl}
}

func (acl ACL) userIdPermission(email string) *ReadWritePerm {
	if acl.BlanketPerm != nil && *acl.BlanketPerm {
		return newPerm(true)
	}
	if acl.Users != nil && len(acl.Users) != 0 {
		if userPerm, exists := acl.Users[email]; exists {
			return newPerm(userPerm)
		}
	}
	return acl.BlanketPerm
}

func (acl ACL) HighestPermFor(userPerms ResolvedUserPerms) *ReadWritePerm {
	if userPerms.IsAdmin() || acl.BlanketPerm.CanWrite() {
		return RWPermWrite()
	}
	// Handle blanket perm
	var maxPerm = RWPermNothing()
	if acl.BlanketPerm.CanRead() {
		maxPerm = RWPermRead()
	}
	if userPerms.isGuest() {
		return maxPerm
	}

	maxPerm = acl.userIdPermission(userPerms.Email)
	if maxPerm.CanWrite() {
		return RWPermWrite()
	}
	if userPerms.Projects != nil {
		for proj, projCanWriteOnEntry := range acl.Projects {
			if projPerm, exists := userPerms.Projects[proj]; exists { // TODO: account for project edit!?
				userCanWriteOnProj := projPerm != nil
				userCanWriteOnProjAndProjCanWriteOnEntry := userCanWriteOnProj && projCanWriteOnEntry
				if userCanWriteOnProjAndProjCanWriteOnEntry {
					return RWPermWrite()
				}
				maxPerm = RWPermRead()
			}
		}
	}
	return maxPerm
}

type ReadWritePerm bool // must always be used as *ReadWritePerm! nil is cant do anything, false is read, true is write
func RWPermWrite() *ReadWritePerm {
	p := ReadWritePerm(true)
	return &p
}
func RWPermRead() *ReadWritePerm {
	p := ReadWritePerm(false)
	return &p
}
func RWPermNothing() *ReadWritePerm {
	return nil
}
func (rw *ReadWritePerm) CanWrite() bool {
	return rw != nil && *rw == true
}
func (rw *ReadWritePerm) CanRead() bool {
	return rw != nil
}

//	func RWPermFor(inp *bool) *ReadWritePerm {
//		return (*ReadWritePerm)(inp)
//	}
func (rw *ReadWritePerm) Copy() *ReadWritePerm {
	if rw == nil {
		return nil
	}
	out := *rw
	return &out
}
func newPerm(canWrite bool) *ReadWritePerm {
	out := ReadWritePerm(canWrite)
	return &out
}

//func noPerm() *ReadWritePerm {
//	return nil
//}

//func maxPermsBetween(ps ...*ReadWritePerm) *ReadWritePerm {
//	if len(ps) == 0 {
//		return nil
//	}
//	var out *ReadWritePerm = nil
//	for _, perm := range ps {
//		if perm != nil {
//			if *perm == true {
//				// Return early because this is the highest possible permission
//				return newPerm(true) // can write
//			}
//			if out == nil {
//				out = newPerm(false)
//			}
//		}
//	}
//	return out
//}
//func minPermsBetween(ps ...*ReadWritePerm) *ReadWritePerm {
//	if len(ps) == 0 {
//		return nil
//	}
//	temp := true
//	for _, perm := range ps {
//		if perm == nil {
//			return nil
//		}
//		if !bool(*perm) && temp {
//			temp = false
//		}
//	}
//	return newPerm(temp)
//}

type Perm[T any] struct {
	Id       T    `bson:"id" json:"id"`
	CanWrite bool `bson:"canWrite" json:"canWrite"`
}

// AccountType is always used as a pointer. // nil is guest (never write), true is admin, false is regular user
type AccountType bool // always used as a ptr.
func (ad *AccountType) IsAdmin() bool {
	return ad != nil && bool(*ad)
}
func (ad *AccountType) IsGuest() bool {
	return ad == nil
}
func (ad *AccountType) IsRegular() bool {
	return ad != nil && !ad.IsAdmin()
}
func (ad *AccountType) String() string {
	if ad.IsGuest() {
		return "Guest" // TODO: USE
	}
	if ad.IsAdmin() {
		return "Admin" // TODO: USE
	}
	return "User" // TODO: USE
}
func AcctTypeAdmin() *AccountType {
	accType := AccountType(true)
	return &accType
}
func AcctTypeNormal() *AccountType {
	accType := AccountType(false)
	return &accType
}
func AcctTypeGuest() *AccountType {
	return nil
}

type ResolvedUserPerms struct {
	Email       string                           `bson:"email" json:"email"`
	AccountType *AccountType                     `bson:"accountType,omitempty" json:"accountType,omitempty"` // nil is guest (never write), false is normal email, true is admin
	Projects    map[projectName]*UserProjectPerm `bson:"projects,omitempty" json:"projects,omitempty"`       // nil is readonly, false is canWrite, true is admin of project
	//CanEditProjects map[projectName]struct{}// TODO: for project editors, maybe have a separate map? I dont really like this...
}

func (perms ResolvedUserPerms) GetUser(ctx context.Context) (*User, error) {
	var usr = &User{}
	email := perms.Email
	err := DbFrom(ctx).Collection(UserCollName).FindOne(ctx, BsonFindFilter(IDfld, email)).Decode(usr)
	return usr, err
}
func (perms ResolvedUserPerms) HasPermissionToEdit(item Permissioned) bool {
	userPerm := item.Permissions().HighestPermFor(perms)
	return userPerm != nil && bool(*userPerm)
}

func (perms ResolvedUserPerms) IsAdmin() bool {
	return perms.AccountType.IsAdmin()
}

func (perms ResolvedUserPerms) isGuest() bool {
	return perms.AccountType.IsGuest()
}
func (perms ResolvedUserPerms) isRegular() bool {
	return perms.AccountType.IsRegular()
}
func (perms ResolvedUserPerms) PermsForProject(projName projectName) *ReadWritePerm {
	if perms.AccountType.IsGuest() {
		return RWPermNothing()
	}
	if perms.IsAdmin() {
		return RWPermWrite()
	}
	// For standard users, rely on the user's project perms
	var userProjPerm *UserProjectPerm = nil
	var exists bool
	if perms.Projects == nil {
		return RWPermNothing()
	}
	userProjPerm, exists = perms.Projects[projName]
	if !exists {
		return RWPermNothing()
	}
	// True is project admin
	// False is can write on project items, but not project itself
	// Nil is can only read // TODO: validate that nil==read is ok here!
	return userProjPerm.RWPerm()
}

func (user ResolvedUserPerms) lowestPermBetweenEntries(entryPermsets ...Permissioned) *ReadWritePerm {
	var out ReadWritePerm = true
	for _, item := range entryPermsets {
		thisPerm := item.Permissions().HighestPermFor(user)
		if thisPerm == nil {
			// At least one permission was nil. Return nil early
			return nil
		}
		out = out && *thisPerm
	}
	return &out
}

type ProjectPerms map[string]ProjectPerm // map of email to perm where nil is readOnly, false is write but not edit the project, true is full control over project
func (pp ProjectPerms) Equal(other ProjectPerms) bool {
	if len(pp) != len(other) {
		return false
	}
	for email, permA := range pp {
		if permB, exists := other[email]; exists {
			if permA == permB {
				continue
			}
		}
		return false
	}
	return true
}
func (pp ProjectPerms) ForUser(email string) *ProjectPerm {
	if perm, exists := pp[email]; exists {
		return &perm
	}
	return nil
}

type ProjectPerm string // "read", "write", or "admin". Used as a pointer, where nil == no perm on project
var (
	// ProjectAdmin defines a ProjectPerm for a user that can write on entries for the specified project (if the project can write to the entry), as well as modify the project itself
	ProjectAdmin ProjectPerm = "admin"
	// TODO: next line
	//// ProjectModify defines a ProjectPerm for a user that can write and read on entries for the specified project, as well as modify everything on the project except permissions
	// TODO: this! ProjectModify ProjectPerm = "modify"
	// ProjectWrite defines a ProjectPerm for a user that can write (and read) on entries for the specified project (if the project can write to the entry)
	ProjectWrite ProjectPerm = "write"
	// ProjectRead defines a ProjectPerm for a user that can read entries for the specified project
	ProjectRead ProjectPerm = "read"
	// ProjectNone *ProjectPerm = nil
)

func (pp ProjectPerm) UserProjectPerm() *UserProjectPerm {
	switch pp {
	case ProjectAdmin:
		return UserProjectAdmin()
	case ProjectWrite:
		return UserProjectWrite()
	case ProjectRead:
		return UserProjectRead()
	//case ProjectModify:
	//	return UserProjectModify() // TODO: FIX! May be physically impossible without changing structs given UserProjectX is a *bool
	default:
		panic("invalid user project perm string: " + string(pp))
	}
}

//func (pp ProjectPerm) ReadWritePerm() *ReadWritePerm { // TODO: error even possible here?
//	return pp.UserProjectPerm().RWPerm()
//}

func (pp *ProjectPerm) IsAdmin() bool {
	return pp != nil && *pp == ProjectAdmin
}
func (pp *ProjectPerm) CanRead() bool {
	return pp != nil
}
func (pp *ProjectPerm) CanWrite() bool {
	return pp != nil && *pp != ProjectRead
}

//func (pp *ProjectPerm) CanModifyProject() bool { // TODO: this!
//	return pp != nil && *pp == ProjectModify
//}

func (projPerm *ProjectPerm) UnmarshalJSON(bs []byte) (err error) {
	s := strings.Trim(string(bs), `"`)
	switch s {
	case "admin":
		*projPerm = ProjectAdmin
	//case "modify":
	//	*projPerm = ProjectModify // TODO: reenable once ready!
	case "write":
		*projPerm = ProjectWrite
	case "read":
		*projPerm = ProjectRead
	default:
		println(fmt.Sprintf("invalid project perm string: %s was not read, write, or admin ", s))
		return fmt.Errorf("unknown project perm: %s, must be 'read', 'write', or 'admin'", s)
	}
	return nil
}

type UserPerms struct {
	Admin    *AccountType  `bson:"admin,omitempty" json:"admin,omitempty"` // nil == guest, false == regular email, true==Admin
	Projects []projectName `bson:"projects,omitempty" json:"projects,omitempty"`
}

var testAclStrings = []string{
	"blanket write",
	"blanket read",
	"blanket private",
	"Test project can write",
	"Test project can read",
	"Test entry without any projects",
}
var testAcls = []ACL{
	{ // Blanket write: 4Wj8HxCMmcs
		BlanketPerm: RWPermWrite(),
		// admin can write
		// testUserEmailPAA can write
		// guest can read
		// guest cannot write
	},
	{ // Blanket read: 31R5AgpvJDD
		BlanketPerm: RWPermRead(),
		// admin can write
		// testUserEmailPAA can read CONFIRMED
		// testUserEmailPAA cannot write CONFIRMED
		// guest can read
		// guest cannot write
	},
	{ // Blanket private: 4gRoJm1rNtP
		BlanketPerm: RWPermNothing(),
		// admin can write
		// ensure testUserEmailPAA cannot read
		// ensure guest cannot read
	},
	{ // Test project can write: 3B7kBVeQuUj
		Users: map[string]bool{
			testUserEmailPAA: true,  // User can write on entry, user can write for project, project can write == write
			testUserEmailPAB: false, // User can read on entry, user can write for project, project can write  == write
			//testUserEmailPAC: nil, // user not on entry     , user can write for project, project can write  == write
			testUserEmailPWA: true,  // User can write on entry, user can write for project , project can write == write
			testUserEmailPWB: false, // User can read on entry, user can write for project , project can write  == write
			//testUserEmailPWC: nil, // user not on entry     , user can write for project , project can write  == write
			testUserEmailPRA: true,  // User can write on entry, user can read for project , project can write == write
			testUserEmailPRB: false, // User can read on entry, user can read for project , project can write  == read
			//testUserEmailPRC: nil, // user not on entry     , user can read for project , project can write  == read
			testUserEmailPNA: true,  // User can write on entry, user not on project       , project can write == write
			testUserEmailPNB: false, // User can read on entry, user not on project       , project can write  == read
			//testUserEmailPNC: nil, // user not on entry     , user not on project       , project can write  == NONE
		},
		Projects: map[projectName]bool{
			testProjectsMap[TestProjectNamePublic].Name: true,
		},
		// admin can write
		// testUserEmailPAA can write CONFIRMED
		// testUserEmailPAB can write CONFIRMED
		// testUserEmailPAC can write CONFIRMED
		// testUserEmailPWA can write CONFIRMED
		// testUserEmailPWB can write CONFIRMED
		// testUserEmailPWC can write CONFIRMED
		// testUserEmailPRA can write CONFIRMED
		// testUserEmailPRB can read but cant write CONFIRMED
		// testUserEmailPRC can read but cant write CONFIRMED
		// testUserEmailPNA can write CONFIRMED
		// testUserEmailPNB can read but cant write CONFIRMED
		// testUserEmailPNC cannot read CONFIRMED
		// guest cannot read CONFIRMED
	}, { // Test project can read: 3LpRCJTuWkF
		Users: map[string]bool{
			testUserEmailPAA: true,  // User can write on entry, user can write for project, project can read == write
			testUserEmailPAB: false, // User can read on entry, user can write for project, project can read  == read
			//testUserEmailPAC: nil, // user not on entry     , user can write for project, project can read  == read
			testUserEmailPWA: true,  // User can write on entry, user can write for project , project can read == write
			testUserEmailPWB: false, // User can read on entry, user can write for project , project can read  == read
			//testUserEmailPWC: nil, // user not on entry     , user can write for project , project can read  == read
			testUserEmailPRA: true,  // User can write on entry, user can read for project , project can write == write
			testUserEmailPRB: false, // User can read on entry, user can read for project , project can write  == read
			//testUserEmailPRC: nil, // user not on entry     , user can read for project , project can write  == read
			testUserEmailPNA: true,  // User can write on entry, user not on project       , project can read == write
			testUserEmailPNB: false, // User can read on entry, user not on project       , project can read  == read
			//testUserEmailPNC: nil, // user not on entry     , user not on project       , project can read  == NONE
		},
		Projects: map[projectName]bool{
			testProjectsMap[TestProjectNamePublic].Name: false,
		},
		// admin can write
		// testUserEmailPAA can write CONFIRMED
		// testUserEmailPAB can read but cant write CONFIRMED
		// testUserEmailPAC can read but cant write CONFIRMED
		// testUserEmailPWA can write CONFIRMED
		// testUserEmailPWB can read but cant write CONFIRMED
		// testUserEmailPWC can read but cant write CONFIRMED
		// testUserEmailPNA can write CONFIRMED
		// testUserEmailPNB can read but cant write CONFIRMED
		// testUserEmailPNC cannot read CONFIRMED
		// guest cannot read CONFIRMED
	}, { // Test entry without any projects: 3gDmDv6tjHH
		Users: map[string]bool{
			testUserEmailPAA: true,  // User can write on entry  == write
			testUserEmailPAB: false, // User can read on entry  == read
			//testUserEmailPNC: nil, // user not on entry       == None
		},
		Projects: map[projectName]bool{},
		// admin can write
		// testUserEmailPAA can write CONFIRMED
		// testUserEmailPAB can read but not write CONFIRMED
		// testUserEmailPNC cannot read CONFIRMED
		// guest cannot read CONFIRMED
	},
}

type PermsOnRequest struct {
	UserPerms    map[string]bool      `json:"users,omitempty"` // Bool is canEdit
	ProjectPerms map[projectName]bool `json:"projects,omitempty"`
	BlanketPerm  *ReadWritePerm       `json:"blanketPerm,omitempty"` // If true then these entries are publicly writeable, if false then publicly readable
}

func (requestPerms PermsOnRequest) GetPermsOnRequest() PermsOnRequest {
	return requestPerms
}

func (requestPerms PermsOnRequest) AsACL() ACL {
	return ACL{
		Users:       requestPerms.UserPerms,
		Projects:    requestPerms.ProjectPerms,
		BlanketPerm: requestPerms.BlanketPerm,
	}
}

// AclForUser turns request perms into an ACL where the current user is ensured to be able to write!
func (requestPerms PermsOnRequest) AclForUser(ctx context.Context, perms ResolvedUserPerms) (AclField, error) {
	client := GetMongoClient(ctx)

	// validate Projects // TODO: check the existing ones first please....
	projectsToCheck := maps.Keys(requestPerms.ProjectPerms)
	count, err := client.
		Database(dbName).
		Collection(ProjectsCollectionName).
		CountDocuments(ctx, bson.M{
			"_id": bson.M{"$in": projectsToCheck},
		})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AclField{}, errors.New("could not find any of the projects")
		}
		return AclField{}, err
	}
	if int(count) != len(projectsToCheck) {
		return AclField{}, errors.New("could not find at least one project")
	}
	// validate users // TODO: only check new ones not on the existing item...
	usersToCheck := maps.Keys(requestPerms.UserPerms)
	count, err = client.
		Database(dbName).
		Collection(UserCollName).
		CountDocuments(ctx, bson.M{
			IDfld: bson.M{"$in": usersToCheck},
		})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AclField{}, errors.New("could not find any of the users")
		}
		return AclField{}, err
	}
	if int(count) != len(usersToCheck) {
		return AclField{}, errors.New("could not find at least one user")
	}

	// Resolve acl
	acl := ACL{
		Users:       requestPerms.UserPerms,
		Projects:    requestPerms.ProjectPerms,
		BlanketPerm: requestPerms.BlanketPerm,
	}
	if acl.Users == nil {
		acl.Users = map[string]bool{} // TODO: do we even want this?
	}
	if acl.Projects == nil {
		acl.Projects = map[projectName]bool{} // TODO: do we even want this?
	}
	// If not blanket write, ensure the user who made the request can write
	if !requestPerms.BlanketPerm.CanWrite() {
		acl.Users[perms.Email] = true // TODO: maybe add this regardless of the blanket perm?
	}
	return AclField{ACL: acl}, nil
}
