package dos

import "regexp"

var rxRoot = regexp.MustCompile("^([a-zA-Z]:)?[/\\\\]")
var rxDrive = regexp.MustCompile("^[a-zA-Z]:")

func joinPath2(a, b string) string {
	if len(a) <= 0 || rxRoot.MatchString(b) || rxDrive.MatchString(b) {
		return b
	}
	switch a[len(a)-1] {
	case '\\', '/', ':':
		return a + b
	default:
		return a + "\\" + b
	}
}

// Equals filepath.Join but this works right when path has drive-letter.
func Join(paths ...string) string {
	result := paths[len(paths)-1]
	for i := len(paths) - 2; i >= 0; i-- {
		result = joinPath2(paths[i], result)
	}
	return result
}
