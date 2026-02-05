package rfid

import (
	"encoding/json"
	"golang.org/x/exp/maps"
	"net/http"
)

// var (
//
//	_ Permissioned = &Perms{}
//	_ Permissioned = &Bag{}
//	_ Permissioned = &Fruit{}
//	_ Permissioned = &FruitingChamber{}
//	_ Permissioned = &GrainJar{}
//	_ Permissioned = &LiquidCulture{}
//	_ Permissioned = &MSS{}
//	_ Permissioned = &Plate{}
//	_ Permissioned = &Slant{}
//	_ Permissioned = &StasisTube{}
//	_ Permissioned = &SporePrint{} // TODO: sporeSwab?
//	_ Permissioned = &Species{}    // TODO: sporeSwab? // TODO: plugs?
//	_ Permissioned = &Subspecies{} // TODO: sporeSwab?
//	_ Permissioned = &Transfer{}   // TODO: HANDLE
//
// )
type Permissioned interface {
	Permissions() *ACL
}

// // TODO: HOLD ON TO USER AND GROUP PERMS IN A CACHE FOR A LITTLE BIT?
//
// // TODO: MAKE PERMS TRANSFER ON TRANSFERS
//
//	type Perms struct {
//		Users    objectPermSubset[AlternateCollectionId] `bson:"userPerms,omitempty" json:"userPerms,omitempty"`       // TODO: index ids
//		Projects objectPermSubset[projectName]           `bson:"projectPerms,omitempty" json:"projectPerms,omitempty"` // TODO: index ids
//		Blanket  perms.Perm                              `bson:"blanketPerms,omitempty" json:"blanketPerms,omitempty"` // TODO: index?
//	}
//
// func (p *Perms) Projects() []projectName { // TODO: ever need to be used?
//
//		if p == nil {
//			return nil // TODO: ok?
//		}
//		return p.Projects.Ids
//	}
//
// func (p *Perms) asField() PermsField { // TODO: ever need to be used?
//
//		return PermsField{p}
//	}
//
// // TODO: ON ENTRY PERMS CHANGE (or creation) modify all email session perms
//
// type ServerSideUserPermsSubset objectPermSubset[AlternateCollectionId]
//
//	func (ssups ServerSideUserPermsSubset) Convert(ctx context.Context) (ClientSideUserPermsSubset, error) {
//		ids := make([]UserPermsPair, len(ssups.Ids))
//		indForId := map[AlternateCollectionId]int{}
//		for i, id := range ssups.Ids {
//			indForId[id] = i
//		}
//		//GET ALL USERS FOR IDS (from db)
//		results := make([]User, len(ssups.Ids))
//		cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(UserCollName).
//			Find(ctx, bson.M{"_id": bson.M{"$in": ssups.Ids}})
//		if err != nil {
//			return ClientSideUserPermsSubset{}, err
//		}
//		if err = cursor.All(ctx, &results); err != nil {
//			return ClientSideUserPermsSubset{}, err
//		}
//		if len(results) != len(ids) {
//			return ClientSideUserPermsSubset{}, errors.New("non-matching id lengths")
//		}
//		for _, email := range results {
//			i, exists := indForId[email.Email]
//			if !exists {
//				return ClientSideUserPermsSubset{}, errors.New("found id not in set")
//			}
//			ids[i] = UserPermsPair{
//				Email:  email.Email,
//				val: email.humanReadableId(),
//			}
//		}
//		return ClientSideUserPermsSubset(objectPermSubset[UserPermsPair]{
//			Ids:      ids,
//			CanWrite: ssups.CanWrite,
//		}), nil
//	}
//
// type ClientSideUserPermsSubset objectPermSubset[UserPermsPair] // TODO: USE THESE
//
//	func (csups ClientSideUserPermsSubset) Convert() ServerSideUserPermsSubset {
//		return ServerSideUserPermsSubset(objectPermSubset[AlternateCollectionId]{
//			Ids: sliceutils.Map(csups.Ids, func(upp UserPermsPair) AlternateCollectionId {
//				return upp.Email
//			}),
//			CanWrite: csups.CanWrite,
//		})
//	}
//
//	type UserPermsPair struct {
//		Email  AlternateCollectionId `json:"id" bson:"id"`
//		val string                `bson:"val" json:"val"` // email or username
//	}
//
// func (p *Perms) Permissions() *Perms         { return p }
// func (p *Perms) ProjectsList() []projectName { return p.ProjectPerms().Ids } // TODO: USE THIS
//
//	func (p *Perms) GetSessionPerm(ctx context.Context) perms.Perm {
//		authInfo, err := GetAuthInfo(ctx)
//		if err != nil {
//			return perms.None
//		}
//		return p.PermissionFor(authInfo)
//
// }
//
//	func (p *Perms) ValidateUserCanWrite(ctx context.Context) error {
//		if p.GetSessionPerm(ctx) != perms.Write {
//			return errors.New("current session not authorized to modify")
//		}
//		return nil
//	}
//
// func (p *Perms) OnUpdateIfDifferent(existingPerms *Perms, upd bson.D) bson.D { // TODO: MAKE THIS ALSO TAKE OLD PERMS AND CROSSCHECK THEM
//
//		if p.ExactMatch(existingPerms) {
//			return upd
//		}
//		return p.OnUpdate(upd)
//	}
//
//	func (p *Perms) OnUpdate(upd bson.D) bson.D {
//		out := upd
//		if p.BlanketPerm() == perms.Write {
//			out = append(upd, bson.E{"$unset", "perms"})
//			return out
//		}
//		out = append(upd, bson.E{"$set", bson.D{{"perms", p}}}) // TODO: is this ok?
//		return out
//	}
//
// func (p *Perms) Union(toAdd *Perms) *Perms { // takes greatest permissions
//
//		if p == nil {
//			if toAdd == nil {
//				return nil
//			}
//			return toAdd // TODO: ok?
//		}
//		return &Perms{
//			Users:    p.UserPerms().combine(toAdd.UserPerms()),
//			Projects: p.ProjectPerms().combine(toAdd.ProjectPerms()),
//			Blanket:  max(p.BlanketPerm(), toAdd.BlanketPerm()),
//		}
//	}
//
// //func (p *Perms) Overlap(other ...*Perms) *Perms { // TODO: takes least permissions
// //	return overlappingPerms(p, other)
// //}
//
//	func (p *Perms) ExactMatch(b *Perms) bool {
//		switch utils.CountNotNil(p, b) {
//		case 0:
//			return true
//		case 2:
//			if p.Blanket != b.Blanket {
//				return false
//			}
//			if p.Projects.Len()-b.Projects.Len() != 0 || p.Users.Len()-b.Users.Len() != 0 {
//				return false
//			}
//			bProjMap := b.Projects.asMap()
//			for proj, permA := range p.Projects.asMap() {
//				permB, exists := bProjMap[proj]
//				if !exists || permA != permB {
//					return false
//				}
//			}
//			bUsrMap := b.Users.asMap()
//			for usr, permA := range p.Users.asMap() {
//				permB, exists := bUsrMap[usr]
//				if !exists || permA != permB {
//					return false
//				}
//			}
//			return true
//		default:
//			if p == nil && b != nil && b.Blanket == perms.Write {
//				return true
//			}
//			if b == nil && p != nil && p.Blanket == perms.Write {
//				return true
//			}
//			return false
//		}
//	}
//
//	func (p *Perms) UserPerms() objectPermSubset[AlternateCollectionId] {
//		if p == nil {
//			return objectPermSubset[AlternateCollectionId]{}
//		}
//		return p.Users
//	}
//
//	func (p *Perms) ProjectPerms() objectPermSubset[projectName] {
//		if p == nil {
//			return objectPermSubset[projectName]{}
//		}
//		return p.Projects
//	}
//
//	func (p *Perms) BlanketPerm() perms.Perm {
//		if p == nil {
//			return perms.Write
//		}
//		return p.Blanket
//	}
//
// // TODO: likely don't want a "overlap" function, as it will detract from performance if overutilized?
//
//	type ProjectPerms struct {
//		Users   objectPermSubset[ProjectPermUserId] `bson:"users,omitempty" json:"users,omitempty"`     // TODO: index? // TODO: make sure to populate emails
//		Blanket perms.Perm                          `bson:"blanket,omitempty" json:"blanket,omitempty"` // If Read, then only writes should be in users, if write then do not have users // TODO: index?
//	}
//
// // TODO: ON PROJECT PERMS CHANGE, CHANGE USER SESSION PERMS
//
//	type ProjectPermUserId struct {
//		Email  AlternateCollectionId `bson:"id" json:"id"`
//		Val string                `bson:"val" json:"val"` // Email or username
//	}
//
//	type objectPermSubset[T comparable] struct {
//		Ids      []T    `bson:"ids,omitempty" json:"ids,omitempty"`
//		CanWrite []bool `bson:"canWrite,omitempty" json:"canWrite,omitempty"`
//	}
//
// func (ops objectPermSubset[T]) WithAuthor(author T) objectPermSubset[T] { // TODO: USE THIS ON ALL IMPORTS?
//
//		out := ops
//		for i, id := range out.Ids {
//			if id == author {
//				if !out.CanWrite[i] {
//					out.CanWrite[i] = true
//				}
//				return out
//			}
//		}
//		out.Ids = append(ops.Ids, author)
//		out.CanWrite = append(ops.CanWrite, true)
//		return out
//	}
//
// func (ops objectPermSubset[T]) Len() int { // TODO: is this used?
//
//		return len(ops.Ids)
//	}
//
//	func (ops objectPermSubset[T]) asMap() map[T]bool {
//		out := make(map[T]bool, len(ops.Ids))
//		for i, id := range ops.Ids {
//			out[id] = ops.CanWrite[i]
//		}
//		return out
//	}
//
//	func (ops objectPermSubset[T]) combine(toAdd objectPermSubset[T]) objectPermSubset[T] {
//		items := map[T]bool{}
//		for i, id := range ops.Ids {
//			items[id] = ops.CanWrite[i]
//		}
//		for i, id := range toAdd.Ids {
//			canWrite, exists := items[id]
//			if exists { // TODO: is this max efficiency?
//				if !canWrite {
//					items[id] = toAdd.CanWrite[i]
//				}
//			} else {
//				items[id] = toAdd.CanWrite[i]
//			}
//		}
//		out := objectPermSubset[T]{
//			Ids:      make([]T, len(items)),
//			CanWrite: make([]bool, len(items)),
//		}
//		for i, key := range maps.Keys(items) { // TODO: is this max efficiency?
//			out.Ids[i] = key
//			out.CanWrite[i] = items[key]
//		}
//		return out
//	}
//
//	func permForCanWriteBool(canWrite bool) perms.Perm {
//		if canWrite {
//			return perms.Write
//		}
//		return perms.Read
//	}
//
//	func canWriteBoolForPerm(p perms.Perm) bool {
//		return p == perms.Write
//	}
//
//	func mapAsPermSubset[T comparable](inp map[T]bool) objectPermSubset[T] {
//		out := objectPermSubset[T]{
//			Ids:      maps.Keys(inp),
//			CanWrite: make([]bool, len(inp)),
//		}
//		for i, id := range out.Ids {
//			out.CanWrite[i] = inp[id]
//		}
//		return out
//	}
//
// func minimalPermsBetween(psIn ...Permissioned) *Perms { // This is minimal perms only
//
//		if len(psIn) == 0 {
//			return nil
//		}
//		ps := sliceutils.Map(psIn, func(permd Permissioned) *Perms { return permd.Permissions() })
//		nonBlanketWrites := []int{}
//		lowestBlanketIndex := 0
//		lowestBlanketPerm := perms.Write
//		for i, p := range ps {
//			if p.BlanketPerm() == perms.Write {
//				continue
//			}
//			nonBlanketWrites = append(nonBlanketWrites, i)
//			if p.BlanketPerm() < lowestBlanketPerm {
//				lowestBlanketIndex = i
//				lowestBlanketPerm = p.BlanketPerm()
//			}
//		}
//		if len(nonBlanketWrites) == 0 {
//			return nil
//		}
//		lowestBlanketItem := ps[lowestBlanketIndex]
//		users := lowestBlanketItem.Users.asMap()
//		Projects := lowestBlanketItem.Projects.asMap()
//		for _, i := range nonBlanketWrites {
//			if i == lowestBlanketIndex {
//				continue
//			}
//			p := ps[i]
//			for j, uid := range ps[i].Users.Ids {
//				canWriteOnToCheck := p.Users.CanWrite[j]
//				if currentUCanWrite, exists := users[uid]; exists {
//					if currentUCanWrite && !canWriteOnToCheck {
//						if lowestBlanketPerm == perms.Read {
//							delete(users, uid)
//						} else {
//							users[uid] = false
//						}
//					}
//				} else {
//					delete(users, uid)
//				}
//			}
//			for j, id := range ps[i].Projects.Ids {
//				canWriteOnToCheck := p.Projects.CanWrite[j]
//				if currentPCanWrite, exists := Projects[id]; exists {
//					if currentPCanWrite && !canWriteOnToCheck {
//						if lowestBlanketPerm == perms.Read {
//							delete(Projects, id)
//						} else {
//							Projects[id] = false
//						}
//					}
//				} else {
//					delete(Projects, id)
//				}
//			}
//		}
//		return &Perms{
//			Users:    mapAsPermSubset(users),
//			Projects: mapAsPermSubset(Projects),
//			Blanket:  lowestBlanketPerm,
//		}
//	}
//
// // TODO: MAKE SURE WE ARE PROPERLY DOING THIS
// func (p *Perms) PermissionFor(userAuthInfo AuthInfo) perms.Perm { // TODO: USE THIS EVERYWHERE
//
//		if p == nil {
//			return perms.Write
//		}
//		// if email is root, return WritePerm
//		if userAuthInfo.Opts.Admin != nil && *userAuthInfo.Opts.Admin {
//			return perms.Write
//		}
//		// Check blanket status
//		if p.Blanket == perms.Write {
//			return perms.Write
//		}
//		out := p.Blanket
//		// Check if email is in item users
//		if userIndex := slices.Index(p.Users.Ids, AlternateCollectionId(userAuthInfo.Email)); userIndex != -1 {
//			if p.Users.CanWrite[userIndex] {
//				return perms.Write
//			}
//			if out == perms.None {
//				out = perms.Read
//			}
//		}
//
//		// Check Projects overlap
//		// Generate set of Projects to grab
//		minToKeep := perms.Read
//		if out == perms.Read {
//			minToKeep = perms.Write
//		}
//		for i, project := range p.Projects.Ids {
//			if permForCanWriteBool(p.Projects.CanWrite[i]) <= minToKeep {
//				continue
//			}
//			if userCanWriteOnProj, exists := userAuthInfo.Opts.Projects[project]; exists {
//				if p.Projects.CanWrite[i] && userCanWriteOnProj {
//					return perms.Write
//				}
//				out = perms.Read
//			}
//		}
//		return out
//	}
//
//	type ProjectWithPerm struct {
//		ProjectName projectName `json:"projectName"`
//		CanWrite    *bool       `json:"canWrite,omitempty"` // nil is perms.None, false is perms.Read, true is perms.Write
//	}
func SessionUserProjectsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := GetAuthInfo(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(maps.Keys(user.projects))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		println("sending projects list: ", string(bs))
		_, err = w.Write(bs)
		if err != nil {
			handleWriteErr(err, w)
		}
	}
}

//func (p *Perms) Valid() bool {
//	return p == nil || p.Blanket == perms.Write || p.Projects.Len()+p.Users.Len() > 0
//}
