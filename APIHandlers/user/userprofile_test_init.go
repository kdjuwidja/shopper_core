package apihandlersuser

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func setUpTestEnv(t *testing.T) (*UserProfileHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	userProfileHandler := InitializeUserProfileHandler(*testDBConn, apiHandlers.ResponseFactory{})
	return userProfileHandler, testDBConn
}
