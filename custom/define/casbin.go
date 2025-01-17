package define

const (
	PrefixSystem   = "system"
	PrefixUser     = "user"
	PrefixRole     = "role"
	PrefixResource = "resource"
)

var ValidAction2Level = map[string]int{
	"view":     1,
	"download": 2,
	"edit":     3,
	"manage":   4,
}
