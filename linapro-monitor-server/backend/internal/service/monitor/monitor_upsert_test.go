// This file verifies linapro-monitor-server snapshot persistence against the
// supported PostgreSQL runtime database when explicitly enabled.

package monitor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "lina-core/pkg/dbdriver"
	"lina-plugin-linapro-monitor-server/backend/internal/dao"
)

// TestPostgreSQLDriverRegisteredForMonitorTests verifies the plugin test
// package imports LinaPro's supported database driver registry before any
// PostgreSQL integration test initializes GoFrame's ORM.
func TestPostgreSQLDriverRegisteredForMonitorTests(t *testing.T) {
	ctx := context.Background()
	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:1)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure PostgreSQL driver registration smoke failed: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	db, err := gdb.NewByGroup()
	if err != nil {
		t.Fatalf("expected PostgreSQL driver to be registered for monitor tests: %v", err)
	}
	if closeErr := db.Close(ctx); closeErr != nil {
		t.Fatalf("close PostgreSQL driver registration smoke database failed: %v", closeErr)
	}
}

// TestUpsertMonitorSnapshotWorksOnPostgreSQL verifies PostgreSQL runtime
// persistence uses explicit Save conflict columns for monitor snapshots.
func TestUpsertMonitorSnapshotWorksOnPostgreSQL(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLMonitorServerDB(t, ctx)

	const (
		nodeIP     = "10.0.0.10"
		firstData  = `{"sample":1}`
		secondData = `{"sample":2}`
	)
	nodeName := fmt.Sprintf("unit-node-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := dao.Server.Ctx(context.Background()).
			Where(colNodeName, nodeName).
			Where(colNodeIp, nodeIP).
			Delete(); err != nil {
			t.Errorf("cleanup PostgreSQL monitor snapshot failed: %v", err)
		}
	})

	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, firstData); err != nil {
		t.Fatalf("insert PostgreSQL monitor snapshot failed: %v", err)
	}
	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, secondData); err != nil {
		t.Fatalf("update PostgreSQL monitor snapshot failed: %v", err)
	}

	count, err := dao.Server.Ctx(ctx).
		Where(colNodeName, nodeName).
		Where(colNodeIp, nodeIP).
		Count()
	if err != nil {
		t.Fatalf("count PostgreSQL monitor snapshots failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one snapshot row after duplicate node upsert, got %d", count)
	}

	var row *serverMonitorRecord
	if err = dao.Server.Ctx(ctx).Where(colNodeName, nodeName).Scan(&row); err != nil {
		t.Fatalf("read PostgreSQL monitor snapshot failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected PostgreSQL monitor snapshot row to exist")
	}
	if row.Data != secondData {
		t.Fatalf("expected latest monitor data %s, got %s", secondData, row.Data)
	}
}

// TestGetDBInfoReturnsPostgreSQLVersion verifies linapro-monitor-server
// database diagnostics return the active PostgreSQL version label.
func TestGetDBInfoReturnsPostgreSQLVersion(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLMonitorServerDB(t, ctx)

	info := New().GetDBInfo(ctx)
	if info == nil {
		t.Fatal("expected PostgreSQL DB info to be returned")
	}
	if !strings.Contains(info.Version, "PostgreSQL") {
		t.Fatalf("expected PostgreSQL database version label, got %q", info.Version)
	}
	if strings.TrimSpace(info.Version) == "" {
		t.Fatalf("expected PostgreSQL database version number to be non-empty, got %q", info.Version)
	}
}

// setupPostgreSQLMonitorServerDB points the generated DAO at an explicit
// PostgreSQL database and creates the linapro-monitor-server table.
func setupPostgreSQLMonitorServerDB(t *testing.T, ctx context.Context) {
	t.Helper()

	link := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if link == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL monitor persistence tests")
	}

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}},
	}); err != nil {
		t.Fatalf("configure PostgreSQL monitor database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close PostgreSQL monitor database failed: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	statements := []string{
		`CREATE TABLE IF NOT EXISTS plugin_linapro_monitor_server (
			"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			"node_name" VARCHAR(128) NOT NULL DEFAULT '',
			"node_ip" VARCHAR(64) NOT NULL DEFAULT '',
			"data" TEXT NOT NULL,
			"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_linapro_monitor_server_node ON plugin_linapro_monitor_server ("node_name", "node_ip")`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute linapro-monitor-server schema SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
}
