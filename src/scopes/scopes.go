package scopes

import pb "github.com/kheina/openconman/src/gen/pbs/auth/auth"

var scopes = map[pb.SCOPE]string{
	pb.SCOPE_UNKNOWN_SCOPE: "unknown",
	pb.SCOPE_ANY_SCOPE:     "*",
	pb.SCOPE_ALL_SCOPES:    "**",
	pb.SCOPE_SYSTEMD:       "systemd",
	pb.SCOPE_ALIAS:         "alias",
	pb.SCOPE_CONTAINERS:    "containers",
}

// func (s pb.SCOPE) String() string {
// 	if str, ok := scopes[s]; ok {
// 		return str
// 	}
// 	return scopes[UnknownScope]
// }

var actions = map[pb.ACTION]string{
	pb.ACTION_UNKNOWN_ACTION: "unknown",
	pb.ACTION_ANY_ACTION:     "*",
	pb.ACTION_CREATE:         "create",
	pb.ACTION_READ:           "read",
	pb.ACTION_UPDATE:         "update",
	pb.ACTION_DELETE:         "delete",
	pb.ACTION_LIST:           "list",
}

// func (a Action) String() string {
// 	if str, ok := actions[a]; ok {
// 		return str
// 	}
// 	return actions[UnknownAction]
// }

func NewScope(scope string) pb.SCOPE {
	if s, ok := scopeMap[scope]; ok {
		return s
	}
	return pb.SCOPE_UNKNOWN_SCOPE
}

func NewAction(action string) pb.ACTION {
	if a, ok := actionMap[action]; ok {
		return a
	}
	return pb.ACTION_UNKNOWN_ACTION
}

var actionMap = map[string]pb.ACTION{}
var scopeMap = map[string]pb.SCOPE{}

func init() {
	for k, v := range actions {
		actionMap[v] = k
	}
	for k, v := range scopes {
		scopeMap[v] = k
	}
}
