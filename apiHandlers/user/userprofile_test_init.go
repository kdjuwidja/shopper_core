package apihandlersuser

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"netherealmstudio.com/m/v2/apiHandlers"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

func setUpTestEnv(t *testing.T) (*UserProfileHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	userProfileHandler := InitializeUserProfileHandler(*testDBConn, apiHandlers.ResponseFactory{})
	return userProfileHandler, testDBConn
}
