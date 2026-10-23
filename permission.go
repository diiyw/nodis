package nodis

type Action uint16

const (
	ActionDiscover Action = 1 << iota
	ActionRead
	ActionCreate
	ActionUpdate
	ActionDelete
	ActionSubscribe
	ActionSync
	ActionManagePolicy
	ActionExecute
	ActionAll = ActionDiscover | ActionRead | ActionCreate | ActionUpdate | ActionDelete | ActionSubscribe | ActionSync | ActionManagePolicy | ActionExecute
)

type Scope uint8

const (
	ScopeDatabase Scope = iota
	ScopeDocument
	ScopeObject
	ScopeField
)

type SubjectKind uint8

const (
	SubjectAny SubjectKind = iota
	SubjectUser
	SubjectRole
	SubjectGroup
	SubjectService
	SubjectAuthenticated
	SubjectAnonymous
)

type Principal struct {
	ID      string
	Roles   []string
	Groups  []string
	Service bool
}

type Resource struct {
	Scope    Scope
	Document DocumentID
	Object   ObjectID
	Field    FieldID
}

type Rule struct {
	SubjectKind SubjectKind
	Subject     string
	Resource    Resource
	Actions     Action
	Allow       bool
}

func (p Principal) hasRole(role string) bool {
	for i := range p.Roles {
		if p.Roles[i] == role {
			return true
		}
	}
	return false
}

func (db *DB) allowedLocked(principal Principal, resource Resource, action Action) bool {
	if principal.ID != "" && (principal.ID == db.admin || principal.hasRole("admin")) {
		return true
	}
	best := -1
	allow := false
	denied := false
	for i := range db.rules {
		r := &db.rules[i]
		if r.Actions&action == 0 || !subjectMatches(principal, r) {
			continue
		}
		specificity, ok := resourceMatches(resource, r.Resource)
		if !ok || specificity < best {
			continue
		}
		if specificity > best {
			best, allow, denied = specificity, false, false
		}
		if r.Allow {
			allow = true
		} else {
			denied = true
		}
	}
	return best >= 0 && allow && !denied
}

func subjectMatches(p Principal, rule *Rule) bool {
	switch rule.SubjectKind {
	case SubjectAny:
		return true
	case SubjectUser:
		return p.ID == rule.Subject
	case SubjectRole:
		return p.hasRole(rule.Subject)
	case SubjectGroup:
		for i := range p.Groups {
			if p.Groups[i] == rule.Subject {
				return true
			}
		}
		return false
	case SubjectService:
		return p.Service && p.ID == rule.Subject
	case SubjectAuthenticated:
		return p.ID != ""
	case SubjectAnonymous:
		return p.ID == ""
	default:
		return false
	}
}

func resourceMatches(want, policy Resource) (int, bool) {
	if policy.Scope > want.Scope {
		return 0, false
	}
	switch policy.Scope {
	case ScopeDatabase:
		return 0, true
	case ScopeDocument:
		return 10, policy.Document == want.Document
	case ScopeObject:
		if policy.Document != want.Document || policy.Object != 0 && policy.Object != want.Object {
			return 0, false
		}
		if policy.Object == want.Object {
			return 22, true
		}
		return 20, true
	case ScopeField:
		if policy.Document != want.Document || policy.Object != 0 && policy.Object != want.Object || policy.Field != 0 && policy.Field != want.Field {
			return 0, false
		}
		specificity := 30
		if policy.Object == want.Object {
			specificity += 2
		}
		if policy.Field == want.Field {
			specificity++
		}
		return specificity, true
	default:
		return 0, false
	}
}

func (db *DB) Allowed(principal Principal, resource Resource, action Action) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return !db.closed && db.allowedLocked(principal, resource, action)
}

func (db *DB) Grant(actor Principal, rule Rule) error {
	rule.Allow = true
	return db.setRule(actor, rule)
}

func (db *DB) Deny(actor Principal, rule Rule) error {
	rule.Allow = false
	return db.setRule(actor, rule)
}

func (db *DB) setRule(actor Principal, rule Rule) error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrClosed
	}
	if !db.allowedLocked(actor, rule.Resource, ActionManagePolicy) {
		db.mu.Unlock()
		return ErrPermission
	}
	next := db.sequence + 1
	payload := appendPolicyRecord(nil, rule)
	if err := db.log.Append(next, payload, db.durable); err != nil {
		db.mu.Unlock()
		return err
	}
	db.rules = append(db.rules, rule)
	db.sequence = next
	db.permissionRevision++
	event := Event{Sequence: next, PermissionRevision: db.permissionRevision, Type: EventPolicyChanged}
	db.publishLocked(event)
	db.mu.Unlock()
	return nil
}

// Revoke removes all rules exactly matching subject, resource, actions, and effect.
func (db *DB) Revoke(actor Principal, rule Rule) error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return ErrClosed
	}
	if !db.allowedLocked(actor, rule.Resource, ActionManagePolicy) {
		db.mu.Unlock()
		return ErrPermission
	}
	found := false
	for i := range db.rules {
		if sameRule(db.rules[i], rule) {
			found = true
			break
		}
	}
	if !found {
		db.mu.Unlock()
		return ErrNotFound
	}
	next := db.sequence + 1
	payload := appendPolicyRevokeRecord(nil, rule)
	if err := db.log.Append(next, payload, db.durable); err != nil {
		db.mu.Unlock()
		return err
	}
	db.removeRuleLocked(rule)
	db.sequence, db.permissionRevision = next, db.permissionRevision+1
	event := Event{Sequence: next, PermissionRevision: db.permissionRevision, Type: EventPolicyChanged}
	db.publishLocked(event)
	db.mu.Unlock()
	return nil
}

func sameRule(a, b Rule) bool {
	return a.SubjectKind == b.SubjectKind && a.Subject == b.Subject && a.Resource == b.Resource && a.Actions == b.Actions && a.Allow == b.Allow
}

func (db *DB) removeRuleLocked(rule Rule) {
	out := db.rules[:0]
	for i := range db.rules {
		if !sameRule(db.rules[i], rule) {
			out = append(out, db.rules[i])
		}
	}
	db.rules = out
}
