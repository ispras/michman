package ansible

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/utils"
	"log"
)

type LauncherServer struct {
	Logger            *log.Logger
	Db                database.Database
	VaultCommunicator utils.SecretStorage
	Config            utils.Config
}
