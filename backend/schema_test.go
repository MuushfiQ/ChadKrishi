package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDatabaseFailureResponses(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		text   string
	}{
		{&mysql.MySQLError{Number: 1146, Message: "private table detail"}, 503, "setup is incomplete"},
		{&mysql.MySQLError{Number: 1054, Message: "private column detail"}, 503, "setup is incomplete"},
		{&mysql.MySQLError{Number: 1452, Message: "private constraint detail"}, 500, "No change was confirmed"},
		{sql.ErrNoRows, 404, "item not found"},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/posts/1/buy", nil)
		dbError(c, tc.err)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.text) {
			t.Fatalf("response %d %s", w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "private") {
			t.Fatal("database detail leaked")
		}
	}
}
func TestSchemaRequiresDatabase(t *testing.T) {
	if checkSchema(nil) == nil {
		t.Fatal("unconfigured database accepted")
	}
}
