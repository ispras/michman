package ansible

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
)

type LauncherServer struct {
	Logger            *logrus.Logger
	Db                database.Database
	VaultCommunicator utils.SecretStorage
	Config            utils.Config
	OsCreds           utils.OsCredentials
}
