package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
)

type AclField struct {
	ACL *ACL `bson:"acl,omitempty" json:"acl,omitempty"`
}

func (field AclField) Permissions() *ACL {
	return field.ACL
}

type DefaultAclField struct {
	ACL *ACL `bson:"defaultAcl,omitempty" json:"defaultAcl,omitempty"`
}

func (field DefaultAclField) DefaultPermissions() *ACL {
	return field.ACL
}

func allCanReadAcl() AclField {
	return AclField{ACL: &ACL{
		BlanketPerm: true,
	}}
}
func allCanWriteAcl() AclField {
	return AclField{ACL: nil}
}

type ACL struct {
	Users       map[string] /*email*/ bool `bson:"users,omitempty" json:"users,omitempty"`             // bool is canWrite
	Projects    map[projectName]bool       `bson:"projects,omitempty" json:"projects,omitempty"`       // bool is canWrite // TODO: project name?
	BlanketPerm bool                       `bson:"blanketPerm,omitempty" json:"blanketPerm,omitempty"` // false is public cannot read by default. // TODO: ENSURE PROPERLY SET EVERYWHERE
}

func (acl *ACL) UnmarshalJSON(bs []byte) (err error) {
	out := &ACL{
		Users:       nil,
		Projects:    nil,
		BlanketPerm: false,
	}
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
		return errors.New("ACL blanketPerm must be a present boolean field")
	}
	out.BlanketPerm, ok = blanketPermIfc.(bool)
	if !ok {
		return errors.New("ACL blanketPerm must be a present boolean field")
	}
	*acl = *(out.simplified())
	return nil
}

// func (acl ACL) MarshalJSON()(bs []byte, err error) is not custom

func (acl *ACL) simplified() *ACL {
	if acl == nil || acl.BlanketPerm == false {
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

// TODO: USE THIS!!!!!!
func (acl *ACL) Equivalent(other *ACL) bool {
	if acl == nil {
		return other == nil
	}
	if other == nil {
		return true
	}
	if (*acl).BlanketPerm != (*other).BlanketPerm {
		return false
	}
	for user, perm := range (*acl).Users {
		otherPerm, exists := (*other).Users[user]
		if !exists || otherPerm != perm {
			return false
		}
	}
	for proj, perm := range (*acl).Projects {
		otherPerm, exists := (*other).Projects[proj]
		if !exists || otherPerm != perm {
			return false
		}
	}
	return true
}

func (acl *ACL) AsField() AclField {
	return AclField{ACL: acl}
}

func (acl *ACL) userIdPermission(email string) *ReadWritePerm {
	if acl == nil {
		return newPerm(true)
	}
	if userPerm, exists := (*acl).Users[email]; exists {
		return newPerm(userPerm)
	}
	return noPerm()
}

// TODO: ensure this works
func (acl *ACL) HighestPermFor(perms ResolvedUserPerms) *ReadWritePerm {
	if acl == nil || (perms.admin != nil && (*perms.admin)) {
		return newPerm(true)
	}
	// TODO: handle guest?
	// Handle blanket perm
	var maxPerm *ReadWritePerm = nil
	if acl.BlanketPerm {
		maxPerm = newPerm(false)
	}

	maxPerm = acl.userIdPermission(perms.Email)
	if maxPerm != nil && *maxPerm == true {
		return newPerm(true)
	}
	for proj, canWrite := range (*acl).Projects {
		if projPerm, exists := perms.projects[proj]; exists {
			userPermForProjectAndEntry := (projPerm != nil) && canWrite
			if userPermForProjectAndEntry {
				return newPerm(projPerm != nil)
			}
			if maxPerm == nil {
				maxPerm = newPerm(false)
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
type ProjectPerms map[string]*bool // TODO: USE // map of email to perm. nil is readOnly, false is write but not edit the project, true is full control over project

func (pp ProjectPerms) Equal(other ProjectPerms) bool {
	if len(pp) != len(other) {
		return false
	}
	for email, permA := range pp {
		permB, exists := other[email]
		if !exists {
			return false
		}
		if permA == nil {
			if permB != nil {
				return false
			}
		} else {
			if permB == nil || *permA != *permB {
				return false
			}
		}
	}
	return true
}

type UserPerms struct { // TODO: USE!
	Admin    *bool         `bson:"admin,omitempty" json:"admin,omitempty"` // nil == guest, false == regular email, true==Admin
	Projects []projectName `bson:"projects,omitempty" json:"projects,omitempty"`
}

func newAlwaysReadableAcl(ctx context.Context, thisUserPerms ResolvedUserPerms, usersThatCanEdit []string, projectsThatCanEdit []projectName) (AclField, error) {
	return PermsOnRequest{
		UserPerms: slices.MapToMap(usersThatCanEdit, func(i string) (string, bool) {
			return i, true
		}),
		ProjectPerms: slices.MapToMap(projectsThatCanEdit, func(i projectName) (projectName, bool) {
			return i, true
		}),
		BlanketPerm: newPerm(false),
	}.AclForUser(ctx, thisUserPerms)
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
var testAcls = []*ACL{{
	BlanketPerm: true, // Blanket read
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
