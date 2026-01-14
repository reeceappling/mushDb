package rfid

import (
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/mongo"
)

// ACL -> users / Projects
// Users -> Projects
// Projects -> users
// UserPermsResolved -> Projects

type AclField struct {
	ACL *ACL `bson:"acl,omitempty" json:"acl,omitempty"`
}

func (field AclField) Permissions() *ACL {
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
	Projects    map[projectName]bool       `bson:"Projects,omitempty" json:"Projects,omitempty"`       // bool is canWrite // TODO: project name?
	BlanketPerm bool                       `bson:"blanketPerm,omitempty" json:"blanketPerm,omitempty"` // false is public cannot read by default. // TODO: ENSURE PROPERLY SET EVERYWHERE
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

func (acl *ACL) userIdPermission(email string) ReadWritePerm {
	if acl == nil {
		return utils.Pointer(true)
	}
	if userPerm, exists := (*acl).Users[email]; exists {
		return &userPerm
	}
	return nil
}

// TODO: ensure this works
func (acl *ACL) HighestPermFor(perms ResolvedUserPerms) ReadWritePerm {
	if acl == nil || (perms.admin != nil && (*perms.admin)) {
		return utils.Pointer(true)
	}
	// TODO: handle guest?
	// Handle blanket perm
	var maxPerm ReadWritePerm = nil
	if acl.BlanketPerm {
		maxPerm = utils.Pointer(false)
	}

	maxPerm = acl.userIdPermission(perms.Email)
	if maxPerm != nil && *maxPerm == true {
		return utils.Pointer(true)
	}
	for proj, canWrite := range (*acl).Projects {
		if projPerm, exists := perms.projects[proj]; exists {
			userPermForProjectAndEntry := (projPerm != nil) && canWrite
			if userPermForProjectAndEntry {
				return utils.Pointer(projPerm != nil)
			}
			if maxPerm == nil {
				maxPerm = utils.Pointer(false)
			}
			maxPerm = maxPermsBetween(maxPerm, utils.Pointer(projPerm != nil))
		}
	}
	return maxPerm
}

type ReadWritePerm *bool // nil is cant do anything, false is read, true is write

func maxPermsBetween(ps ...ReadWritePerm) ReadWritePerm {
	if len(ps) == 0 {
		return nil
	}
	var out ReadWritePerm = nil
	for _, perm := range ps {
		if perm != nil {
			if *perm {
				return utils.Pointer(true) // can write
			}
			if out == nil {
				out = utils.Pointer(false)
			}
		}
	}
	return out
}
func minPermsBetween(ps ...ReadWritePerm) ReadWritePerm {
	if len(ps) == 0 {
		return nil
	}
	temp := true
	for _, perm := range ps {
		if perm == nil {
			return nil
		}
		if !(*perm) && temp {
			temp = false
		}
	}
	return &temp
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
	return userPerm != nil && *userPerm
}

func (perms ResolvedUserPerms) isAdmin() bool {
	return perms.admin != nil && *perms.admin
}

func (perms ResolvedUserPerms) isGuest() bool {
	return perms.admin == nil
}

func (perms ResolvedUserPerms) lowestPermBetweenEntries(entryPermsets ...Permissioned) ReadWritePerm {
	out := true
	for _, item := range entryPermsets {
		minPermsBetween()
		thisPerm := item.Permissions().HighestPermFor(perms)
		if thisPerm == nil {
			return nil
		}
		out = out && *thisPerm
	}
	return &out
}

type ProjectPerms map[string]*bool // TODO: USE // map of email to perm. nil is readOnly, false is write but not edit the project, true is full control over project

type UserPerms struct { // TODO: USE!
	Admin    *bool         `bson:"admin,omitempty" json:"admin,omitempty"` // nil == guest, false == regular email, true==Admin
	Projects []projectName `bson:"projects,omitempty" json:"projects,omitempty"`
}

func newAlwaysReadableAcl(ctx mongo.SessionContext, thisUserPerms ResolvedUserPerms, usersThatCanEdit []string, projectsThatCanEdit []projectName) (AclField, error) {
	return PermsOnRequest{
		UserPerms: slices.MapToMap(usersThatCanEdit, func(i string) (string, bool) {
			return i, true
		}),
		ProjectPerms: slices.MapToMap(projectsThatCanEdit, func(i projectName) (projectName, bool) {
			return i, true
		}),
		BlanketPerm: utils.Pointer(false),
	}.AclFor(ctx, thisUserPerms)
}
