package virtualbox

import (
	"os"
	"path/filepath"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

func vboxManageExecutable() string {
	if configuration.Windows {
		programFiles, exists := os.LookupEnv("ProgramFiles")
		if !exists {
			// Default to this folder if ProgramFiles env var doesn't exist
			programFiles = "C:\\Program Files"
		}
		return filepath.Join(programFiles, "Oracle", "VirtualBox", "VBoxManage.exe")
	}
	return "vboxmanage"
}
