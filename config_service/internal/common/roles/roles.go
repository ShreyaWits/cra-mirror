package roles

const (
	RoleAdmin  = "ADMIN"
	RoleViewer = "VIEWER"
)

var RoleHierarchy = map[string]int{
	"ADMIN":  1,
	"VIEWER": 2,
}
