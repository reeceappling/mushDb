package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
)

type AclField struct { // Nil means allCanWrite
	ACL ACL `bson:"acl,omitempty" json:"acl,omitempty"`
}

func (field AclField) Permissions() ACL {
	return field.ACL
}

type DefaultAclField struct {
	ACL ACL `bson:"defaultAcl,omitempty" json:"defaultAcl,omitempty"`
}

func (field DefaultAclField) DefaultPermissions() ACL {
	return field.ACL
}

func allCanReadAcl(owner *string) AclField {
	out := ACL{
		BlanketPerm: utils.Pointer(false),
	}
	if owner != nil {
		out.Users[*owner] = true
	}
	return AclField{ACL: out}
}
func allCanWriteAcl() AclField {
	return AclField{ACL: ACL{
		BlanketPerm: utils.Pointer(true),
	}}
}

// ACL being nil means anyone authenticated can do anything (read/write)
type ACL struct { // TODO: ALWAYS ENSURE REFERENCED AS A STRUCT AND NOT A POINTER!
	Users    map[string] /*email*/ bool `bson:"users,omitempty" json:"users,omitempty"`       // bool is canWrite
	Projects map[projectName]bool       `bson:"projects,omitempty" json:"projects,omitempty"` // bool is canWrite
	// TODO: validate blanketPerm works as intended. It was changed from bool to *bool
	BlanketPerm *bool `bson:"blanketPerm,omitempty" json:"blanketPerm,omitempty"` // empty is private, false is public can read by default. True means public can write by default.
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

// TODO: USE THIS EVERYWHERE NEEDED!!!
func (acl ACL) Clone() ACL {
	//if acl == nil {
	//	return nil
	//}
	return ACL{
		Users:       cloneMap(acl.Users),
		Projects:    cloneMap(acl.Projects),
		BlanketPerm: acl.BlanketPerm,
	}
	//return &out
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
	*out.BlanketPerm, ok = blanketPermIfc.(bool)
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
	// If blanketPerm is read, then remove all users/projects that can only read
	for user, canWrite := range acl.Users {
		if !canWrite {
			delete(acl.Users, user)
		}
	}
	for proj, canWrite := range acl.Projects {
		if !canWrite {
			delete(acl.Projects, proj)
		}
	}
	return acl
}

func (acl ACL) Equivalent(other ACL) bool {
	if acl.BlanketPerm != other.BlanketPerm {
		return false
	}
	for user, perm := range acl.Users {
		otherPerm, exists := other.Users[user]
		if !exists || otherPerm != perm {
			return false
		}
	}
	for proj, perm := range acl.Projects {
		otherPerm, exists := other.Projects[proj]
		if !exists || otherPerm != perm {
			return false
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
	if userPerm, exists := acl.Users[email]; exists {
		return newPerm(userPerm)
	}
	return (*ReadWritePerm)(acl.BlanketPerm)
}

// TODO: ensure this works
func (acl ACL) HighestPermFor(perms ResolvedUserPerms) *ReadWritePerm {
	if perms.isAdmin() || (acl.BlanketPerm != nil && *acl.BlanketPerm) {
		return newPerm(true)
	}
	// Handle blanket perm
	var maxPerm *ReadWritePerm = nil
	if acl.BlanketPerm != nil {
		maxPerm = newPerm(false)
	}
	if perms.isGuest() {
		return maxPerm
	}

	maxPerm = acl.userIdPermission(perms.Email)
	if maxPerm != nil && *maxPerm == true {
		return newPerm(true)
	}
	for proj, canWrite := range acl.Projects {
		if projPerm, exists := perms.projects[proj]; exists {
			userPermForProjectAndEntry := (projPerm != nil) && canWrite
			if userPermForProjectAndEntry {
				return newPerm(projPerm != nil) // TODO: ensure ok
			}
			if maxPerm == nil {
				maxPerm = newPerm(false) // TODO: ensure ok
			}
			maxPerm = maxPermsBetween(maxPerm, newPerm(projPerm != nil))
		}
	}
	return maxPerm
}

type ReadWritePerm bool // must always be used as *ReadWritePerm! nil is cant do anything, false is read, true is write
func (rw *ReadWritePerm) Copy() *ReadWritePerm {
	if rw == nil {
		return nil
	}
	return utils.Pointer(ReadWritePerm(*rw == true))
}
func newPerm(canWrite bool) *ReadWritePerm {
	out := ReadWritePerm(canWrite)
	return &out
}
func noPerm() *ReadWritePerm {
	return nil
}

func maxPermsBetween(ps ...*ReadWritePerm) *ReadWritePerm {
	if len(ps) == 0 {
		return nil
	}
	var out *ReadWritePerm = nil
	for _, perm := range ps {
		if perm != nil {
			if *perm {
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

type ResolvedUserPerms struct {
	Email    string                `bson:"email" json:"email"`
	admin    *bool                 // nil is guest (never write), false is normal email, true is admin
	projects map[projectName]*bool // nil is readonly, false is canWrite, true is admin of project
}

func (perms ResolvedUserPerms) HasPermissionToEdit(item Permissioned) bool {
	userPerm := item.Permissions().HighestPermFor(perms)
	return userPerm != nil && bool(*userPerm)
}

func (perms ResolvedUserPerms) isAdmin() bool {
	return perms.admin != nil && *perms.admin
}

func (perms ResolvedUserPerms) isGuest() bool {
	return perms.admin == nil
}
func (perms ResolvedUserPerms) PermsForProject(projName projectName) *ReadWritePerm {
	if perms.admin == nil {
		return nil // Guest, cant
	}
	if *perms.admin {
		return newPerm(true)
	}
	userProjPerm, exists := perms.projects[projName]
	if !exists {
		return nil
	}
	if userProjPerm == nil {
		return newPerm(false)
	}
	// False is can write on project items, but not project itself
	// True is project admin
	return newPerm(*userProjPerm == true)
}

func (perms ResolvedUserPerms) lowestPermBetweenEntries(entryPermsets ...Permissioned) *ReadWritePerm {
	out := true
	for _, item := range entryPermsets {
		minPermsBetween()
		thisPerm := item.Permissions().HighestPermFor(perms)
		if thisPerm == nil {
			return nil
		}
		out = out && bool(*thisPerm)
	}
	return newPerm(out)
}

// TODO: change to "admin/read/write" ?
type ProjectPerms map[string]string // TODO: USE // map of email to perm. nil is readOnly, false is write but not edit the project, true is full control over project

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

type UserPerms struct {
	Admin    *bool         `bson:"admin,omitempty" json:"admin,omitempty"` // nil == guest, false == regular email, true==Admin
	Projects []projectName `bson:"projects,omitempty" json:"projects,omitempty"`
}

// TODO: use?
func newAlwaysReadableAcl(ctx context.Context, thisUserPerms ResolvedUserPerms, usersThatCanEdit []string, projectsThatCanEdit []projectName) (AclField, error) {
	return PermsOnRequest{
		UserPerms: slices.MapToMap(usersThatCanEdit, func(i string) (string, bool) {
			return i, true
		}),
		ProjectPerms: slices.MapToMap(projectsThatCanEdit, func(i projectName) (projectName, bool) {
			return i, true
		}),
		BlanketPerm: utils.Pointer(false), // TODO: used to be *false...
	}.AclForUser(ctx, thisUserPerms)
}
func alwaysWriteableAcl() AclField { // TODO: use?
	return AclField{
		ACL: ACL{
			BlanketPerm: utils.Pointer(true),
		},
	}
}

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
	BlanketPerm: utils.Pointer(false), // Blanket read
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
