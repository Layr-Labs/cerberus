package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainer struct {
	Container testcontainers.Container
	DB        *sql.DB
}

func CreateTestContainer(t *testing.T) (*TestContainer, error) {
	ctx := context.Background()

	// Container configuration
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %v", err)
	}

	// Connection string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host,
		port.Port(),
	)

	// Connect and create test schema
	var db *sql.DB
	// Retry logic for database connection
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Create test schema
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS public.keys_metadata (
            public_key_g1 VARCHAR(255) PRIMARY KEY,
            public_key_g2 VARCHAR(255) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %v", err)
	}

	return &TestContainer{
		Container: container,
		DB:        db,
	}, nil
}

type testDB struct {
	db   *sql.DB
	Repo *keyMetadataRepo
}

// Modified test setup function
func SetupTestDB(t *testing.T) *testDB {
	container, err := CreateTestContainer(t)
	require.NoError(t, err)

	// Register cleanup
	t.Cleanup(func() {
		if err := container.DB.Close(); err != nil {
			t.Errorf("Failed to close db connection: %v", err)
		}
		if err := container.Container.Terminate(context.Background()); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	})

	return &testDB{
		db:   container.DB,
		Repo: &keyMetadataRepo{db: container.DB},
	}
}
