package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		out.Users = map[string]bool{*owner: true}
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
	Users       map[string] /*email*/ bool `bson:"users,omitempty" json:"users,omitempty"`             // bool is canWrite // TODO: omitempty ok?
	Projects    map[projectName]bool       `bson:"projects,omitempty" json:"projects,omitempty"`       // bool is canWrite // TODO: omitempty ok?
	BlanketPerm *ReadWritePerm             `bson:"blanketPerm,omitempty" json:"blanketPerm,omitempty"` // empty is private, false is public can read by default. True means public can write by default.
}

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

func (acl *ACL) UnmarshalJSON(bs []byte) (err error) {
	out := &ACL{}
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
	blanketPermIfc, ok := temp["blanketPerm"]
	if !ok {
		return errors.New("ACL blanketPerm must be a present boolean field v1")
	}
	*out.BlanketPerm, ok = blanketPermIfc.(ReadWritePerm)
	if !ok {
		return errors.New("ACL blanketPerm must be a present boolean field v2")
	}
	*acl = out.simplified()
	return nil
}

// func (acl ACL) MarshalJSON()(bs []byte, err error) is not custom

func (acl ACL) simplified() ACL {
	if acl.BlanketPerm == nil {
		// If admin or default permission is no permission, return self
		return acl
	}
	// If blanketPerm is read, then remove all users that can only read
	if acl.Users != nil {
		for user, canWrite := range acl.Users {
			if !canWrite {
				delete(acl.Users, user)
			}
		}
	}
	if len(acl.Users) == 0 {
		acl.Users = nil // TODO: ensure ok!
	}
	return acl
}

func (acl ACL) Equivalent(other ACL) bool {
	if acl.BlanketPerm != other.BlanketPerm {
		return false
	}
	if acl.Users == nil {
		if other.Users != nil || len(other.Users) != 0 {
			return false
		}
	} else {
		for user, perm := range acl.Users {
			otherPerm, exists := other.Users[user]
			if !exists || otherPerm != perm {
				return false
			}
		}
	}
	if acl.Projects == nil {
		if other.Projects != nil || len(other.Projects) != 0 {
			return false
		}
	} else {
		for proj, perm := range acl.Projects {
			otherPerm, exists := other.Projects[proj]
			if !exists || otherPerm != perm {
				return false
			}
		}
	}

	return true
}

func (acl ACL) AsField() AclField {
	return AclField{ACL: acl}
}

func (acl ACL) userIdPermission(email string) *ReadWritePerm {
	if acl.BlanketPerm != nil && *acl.BlanketPerm {
		return newPerm(true)
	}
	if acl.Users != nil {
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
		println("user can read") // TODO: del!
		maxPerm = RWPermRead()
	}
	if userPerms.isGuest() {
		println("user was guest, returning perm: ", maxPerm) // TODO: del!
		return maxPerm
	}

	maxPerm = acl.userIdPermission(userPerms.Email)
	if maxPerm.CanWrite() {
		return RWPermWrite()
	}
	if userPerms.Projects != nil {
		for proj, projCanWriteOnEntry := range acl.Projects {
			if projPerm, exists := userPerms.Projects[proj]; exists {
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
	AccountType *AccountType                     `bson:"accountType" json:"accountType"`               // nil is guest (never write), false is normal email, true is admin
	Projects    map[projectName]*UserProjectPerm `bson:"projects,omitempty" json:"projects,omitempty"` // nil is readonly, false is canWrite, true is admin of project
}

func (perms ResolvedUserPerms) GetUser(ctx context.Context) (*User, error) {
	var usr = &User{}
	email := perms.Email
	err := DbFrom(ctx).Collection(UserCollName).FindOne(ctx, BsonFindFilter("_id", email)).Decode(usr)
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
	if perms.Projects != nil {
		userProjPerm, exists = perms.Projects[projName]
		if !exists {
			return RWPermNothing()
		}
	}
	// True is project admin
	// False is can write on project items, but not project itself
	// Nil is can only read
	return userProjPerm.RWPerm() // TODO: validate that nil==read is ok here!
}

func (user ResolvedUserPerms) lowestPermBetweenEntries(entryPermsets ...Permissioned) *ReadWritePerm {
	var out ReadWritePerm = true
	for _, item := range entryPermsets {
		perm := item.Permissions()

		thisPerm := perm.HighestPermFor(user)
		if thisPerm == nil {
			// At least one permission was nil. Return nil early
			return nil
		}
		// TODO: variation 1, test against variation 2
		out = out && *thisPerm
		// TODO: variation 2, test against variation 1
		//if *thisPerm || !out  {
		//	continue
		//}
		//out = false
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
var (                   // TODO: const?
	// ProjectAdmin defines a ProjectPerm for a user that can write on entries for the specified project (if the project can write to the entry), as well as modify the project itself
	ProjectAdmin ProjectPerm = "admin"
	// TODO: next line
	//// ProjectModify defines a ProjectPerm for a user that can write and read on entries for the specified project, as well as modify everything on the project except permissions
	//ProjectModify ProjectPerm = "modify"
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

//func (pp *ProjectPerm) CanModify() bool { // TODO: this!
//	return pp != nil && *pp != ProjectRead
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

// TODO: consider moving!
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
