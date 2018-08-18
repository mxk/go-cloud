package iamx

// Entity identifies IAM entity type by its ID prefix.
type Entity string

// IAM entity ID prefixes
// (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-prefixes).
const (
	Invalid         = Entity("")
	AccessKey       = Entity("AKIA")
	Group           = Entity("AGPA")
	InstanceProfile = Entity("AIPA")
	ManagedPolicy   = Entity("ANPA")
	PolicyVersion   = Entity("ANVA")
	Role            = Entity("AROA")
	Root            = Entity("A3T")
	ServerCert      = Entity("ASCA")
	TempKey         = Entity("ASIA")
	User            = Entity("AIDA")
)

// Type identifies entity type by its ID prefix.
func Type(id string) (e Entity) {
	if len(id) >= 4 {
		switch Entity(id[:3]) {
		case AccessKey[:3]:
			e = AccessKey
		case Group[:3]:
			e = Group
		case InstanceProfile[:3]:
			e = InstanceProfile
		case ManagedPolicy[:3]:
			e = ManagedPolicy
		case PolicyVersion[:3]:
			e = PolicyVersion
		case Role[:3]:
			e = Role
		case Root:
			return Root
		case ServerCert[:3]:
			e = ServerCert
		case TempKey[:3]:
			e = TempKey
		case User[:3]:
			e = User
		}
		if id[3] != 'A' {
			e = Invalid
		}
	} else if Entity(id) == Root {
		e = Root
	}
	return
}

// Is returns true if id belongs to the specified entity type.
func Is(id string, typ Entity) bool {
	return 0 < len(typ) && len(typ) <= len(id) && id[:len(typ)] == string(typ)
}
