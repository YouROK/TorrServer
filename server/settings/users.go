package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ListUsers returns account names from accs.db.
func ListUsers() []string {
	buf, err := os.ReadFile(filepath.Join(Path, "accs.db"))
	if err != nil {
		return nil
	}

	accs := make(map[string]string)
	if err := json.Unmarshal(buf, &accs); err != nil {
		return nil
	}

	users := make([]string, 0, len(accs))
	for user := range accs {
		users = append(users, user)
	}
	sort.Strings(users)
	return users
}
