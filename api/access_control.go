package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type AclField struct {
	ACL ACL `bson:"acl" json:"acl"` // TODO: removed omitempty
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
		BlanketPerm:  acl.BlanketPerm, // TODO; clone?
	}
}

func cloneMap[T comparable, U any](m map[T]U) map[T]U { // TODO: use wherever needed
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

// TODO: is this ok if we don't want to remove projects that are below the threshold?
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
		acl.Users = nil
	}
	//for proj, canWrite := range acl.Projects { // TODO: keep all projects!
	//	if !canWrite {
	//		delete(acl.Projects, proj)
	//	}
	//}
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

// TODO: ensure this works
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
	for proj, projCanWriteOnEntry := range acl.Projects {
		if projPerm, exists := userPerms.projects[proj]; exists { // TODO: is projectPerm here ok
			userCanWriteOnProj := projPerm != nil
			userPermForProjectAndEntry := userCanWriteOnProj && projCanWriteOnEntry
			if userPermForProjectAndEntry {
				return RWPermWrite() //TODO: ?newPerm(projPerm != nil) // TODO: ensure ok
			}
			maxPerm = RWPermRead()
			//if maxPerm == nil {
			//	maxPerm = RWPermRead()
			//}
			// TODO: probably don't need: maxPerm = maxPermsBetween(maxPerm, newPerm(projPerm != nil)) // TODO: newPerm(projPerm != nil) or RWPermRead()
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

func maxPermsBetween(ps ...*ReadWritePerm) *ReadWritePerm {
	if len(ps) == 0 {
		return nil
	}
	var out *ReadWritePerm = nil
	for _, perm := range ps {
		if perm != nil {
			if *perm == true {
				// Return early because this is the highest possible permission
				return newPerm(true) // can write
			}
			if out == nil {
				out = newPerm(false)
			}
		}
	}
	return out
}
func minPermsBetween(ps ...*ReadWritePerm) *ReadWritePerm {
	if len(ps) == 0 {
		return nil
	}
	temp := true
	for _, perm := range ps {
		if perm == nil {
			return nil
		}
		if !bool(*perm) && temp {
			temp = false
		}
	}
	return newPerm(temp)
}

type Perm[T any] struct {
	Id       T    `bson:"id" json:"id"`
	CanWrite bool `bson:"canWrite" json:"canWrite"`
}

// AccountType is always used as a pointer. // nil is guest (never write), true is admin, false is regular user
type AccountType bool // always used as a ptr.
func (ad *AccountType) IsAdmin() bool {
	return bool(*ad)
}
func (ad *AccountType) IsGuest() bool {
	return ad == nil
}
func (ad *AccountType) IsRegular() bool {
	return !ad.IsAdmin() && !ad.IsGuest() // TODO: consider switching order
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
	accountType *AccountType                     // nil is guest (never write), false is normal email, true is admin
	projects    map[projectName]*UserProjectPerm // nil is readonly, false is canWrite, true is admin of project
}

func (perms ResolvedUserPerms) HasPermissionToEdit(item Permissioned) bool {
	userPerm := item.Permissions().HighestPermFor(perms)
	return userPerm != nil && bool(*userPerm)
}

func (perms ResolvedUserPerms) IsAdmin() bool {
	return perms.accountType.IsAdmin()
}

func (perms ResolvedUserPerms) isGuest() bool {
	return perms.accountType.IsGuest()
}
func (perms ResolvedUserPerms) isRegular() bool {
	return perms.accountType.IsRegular()
}
func (perms ResolvedUserPerms) PermsForProject(projName projectName) *ReadWritePerm {
	if perms.accountType.IsGuest() {
		return RWPermNothing()
	}
	if perms.IsAdmin() {
		return RWPermWrite()
	}
	// For standard users, rely on the user's project perms
	userProjPerm, exists := perms.projects[projName]
	if !exists {
		return RWPermNothing()
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
var (
	// ProjectAdmin defines a ProjectPerm for a user that can write on entries for the specified project (if the project can write to the entry), as well as modify the project itself
	ProjectAdmin ProjectPerm = "admin"
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

func (projPerm *ProjectPerm) UnmarshalJSON(bs []byte) (err error) {
	s := strings.Trim(string(bs), `"`)
	switch s {
	case "admin":
		*projPerm = ProjectAdmin
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
	"blanket read",
	"Test user can write",
	"Test user can read",
	"Test project can write (and so can test user)",
	"Test project can read (and so can test user)",
	"Test project can read but user can write",
	"Project without test user can write, so user cannot",
}
var testAcls = []ACL{{
	BlanketPerm: RWPermRead(), // Blanket read
}, {
	Users: map[string]bool{testUserEmail: true}, // Test user can write
}, {
	Users: map[string]bool{testUserEmail: false}, // Test user can read
}, {
	Projects: map[projectName]bool{testProjects[0].Name: true}, // Test project can write (and so can test user)
}, {
	Projects: map[projectName]bool{testProjects[0].Name: false}, // Test project can read (and so can test user)
}, {
	Users:    map[string]bool{testUserEmail: true}, // Test project can read but user can write
	Projects: map[projectName]bool{testProjects[0].Name: false},
}, {
	Projects: map[projectName]bool{testProjects[3].Name: true}, // Project without test user can write, so user cannot
}}
