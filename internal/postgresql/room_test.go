package postgresql

import (
	"context"
	"fmt"
	"log"
	opendisc "open_discord"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupPostgresContainer(t *testing.T) (*pgxpool.Pool, error) {
	ctx := context.Background()
	dbName := "open_disc"
	dbUser := "appuser"
	dbPass := "password"
	testData, _ := filepath.Abs("../../testdata/init-db.sql")

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(testData),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.BasicWaitStrategies(),
	)

	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
				fmt.Printf("failed to terminate container: %s", err)
			}
		})

	if err != nil {
		t.Fatal(err)
		return nil, err
	}

	dbURL, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	t.Cleanup(func() { pool.Close() })

	return pool, nil
}

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()

	pool, err := SetupPostgresContainer(t)
	if err != nil {
		t.Fatal(err)
	}

	roomService := RoomService{DB: pool}
	req := opendisc.CreateRoomRequest{Name: "test"}

	room, _ := roomService.Create(ctx, req)
	assert.Equal(t, req.Name, room.Name)

	getResult, _ := roomService.GetByID(ctx, room.ID)
	assert.Equal(t, req.Name, getResult.Name)

	found := false

	getAllResult, _ := roomService.GetAllRooms(ctx)

	for _, elem := range getAllResult {
		if elem.Name == req.Name {
			found = true
		}
	}

	assert.True(t, found)
}

func TestReorderRooms(t *testing.T) {
	ctx := context.Background()
	pool, err := SetupPostgresContainer(t)
	if err != nil {
		t.Fatal(err)
	}
	roomService := RoomService{DB: pool}

	first, _ := roomService.Create(ctx, opendisc.CreateRoomRequest{Name: "First Room"})
	second, _ := roomService.Create(ctx, opendisc.CreateRoomRequest{Name: "Second Room"})
	third, _ := roomService.Create(ctx, opendisc.CreateRoomRequest{Name: "Third Room"})

	assert.Equal(t, 1, first.SortOrder)
	assert.Equal(t, 2, second.SortOrder)
	assert.Equal(t, 3, third.SortOrder)

	roomService.ReorderRooms(ctx, opendisc.SwapRoomOrderRequest{
		RoomIDs: []uuid.UUID{second.ID, third.ID, first.ID},
	})

	first, _ = roomService.GetByID(ctx, first.ID)
	second, _ = roomService.GetByID(ctx, second.ID)
	third, _ = roomService.GetByID(ctx, third.ID)

	assert.Equal(t, 3, first.SortOrder)
	assert.Equal(t, 1, second.SortOrder)
	assert.Equal(t, 2, third.SortOrder)
}
