package runner

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"

	"github.com/ohler55/ojg/jp"
)

// DB executes a database query or command (Postgres or Redis).
func DB(wfConfig workflow.Config, step workflow.Step, store *state.Store) error {
	var engine workflow.DBEngine
	var query string
	var command []string

	if step.DB != nil {
		engine = step.DB.Engine
		query = step.DB.Query
		command = step.DB.Command
	} else {
		engine = workflow.DBEngine(step.Engine)
		query = step.Query
		switch c := step.Command.(type) {
		case string:
			command = []string{c}
		case []string:
			command = c
		case []any:
			for _, v := range c {
				command = append(command, fmt.Sprintf("%v", v))
			}
		}
	}

	if engine == "" {
		engine = workflow.DBEnginePostgres // Default to postgres as requested
	}

	ctx := context.Background()

	switch engine {
	case workflow.DBEnginePostgres:
		return runPostgres(ctx, wfConfig.Postgres.DSN, query, step, store)
	case workflow.DBEngineRedis:
		return runRedis(ctx, wfConfig.Redis.Addr, command, step, store)
	default:
		return fmt.Errorf("unsupported db engine: %s", engine)
	}
}

func runPostgres(ctx context.Context, defaultDSN string, queryRaw string, step workflow.Step, store *state.Store) error {
	dsn := template.Render(defaultDSN, store.All())
	if dsn == "" {
		return fmt.Errorf("postgres DSN is not configured")
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer conn.Close(ctx)

	query := template.Render(queryRaw, store.All())
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Capture all rows as a slice of maps
	var results []map[string]any
	fields := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, field := range fields {
			row[field.Name] = values[i]
		}
		results = append(results, row)
	}

	if rows.Err() != nil {
		return fmt.Errorf("error during row scanning: %w", rows.Err())
	}

	// Extraction
	if len(results) > 0 {
		store.Set("db_results", results)
	}

	if len(step.Extract) > 0 {
		for varName, path := range step.Extract {
			expr, err := jp.ParseString(path)
			if err != nil {
				return fmt.Errorf("invalid jsonpath %s: %w", path, err)
			}

			// We treat results as the root data for JP
			res := expr.Get(results)
			if len(res) == 0 {
				store.Set(varName, nil)
			} else if len(res) == 1 {
				store.Set(varName, res[0])
			} else {
				store.Set(varName, res)
			}
		}
	}

	// Assertions
	ae := NewAssertionEngine(store.All())
	for _, rule := range step.Assert {
		if err := ae.Check(rule, results); err != nil {
			return fmt.Errorf("assertion failed: %w", err)
		}
	}

	return nil
}

func runRedis(ctx context.Context, defaultAddr string, command []string, step workflow.Step, store *state.Store) error {
	addr := template.Render(defaultAddr, store.All())
	if addr == "" {
		addr = "localhost:6379" // Default Redis
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	defer rdb.Close()

	if len(command) == 0 {
		return fmt.Errorf("redis requires a command (e.g. ['GET', 'key'])")
	}

	// Render command arguments
	args := make([]interface{}, len(command))
	for i, arg := range command {
		args[i] = template.Render(arg, store.All())
	}

	cmd := rdb.Do(ctx, args...)
	val, err := cmd.Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("redis command failed: %w", err)
	}

	// For Redis, we usually get a single value (string, int) or a slice.
	// If it's Nil, we store as null.
	if err == redis.Nil {
		val = nil
	}

	store.Set("db_results", val)

	if len(step.Extract) > 0 {
		for varName, path := range step.Extract {
			if path == "$" {
				store.Set(varName, val)
				continue
			}

			// If we got a complex structure (HGETALL etc), we might want to use JP
			// For simple values, JP won't do much unless people expect results to be JSON strings
			expr, err := jp.ParseString(path)
			if err != nil {
				return fmt.Errorf("invalid jsonpath %s: %w", path, err)
			}

			res := expr.Get(val)
			if len(res) == 0 {
				store.Set(varName, nil)
			} else if len(res) == 1 {
				store.Set(varName, res[0])
			} else {
				store.Set(varName, res)
			}
		}
	}

	// Assertions
	ae := NewAssertionEngine(store.All())
	for _, rule := range step.Assert {
		if err := ae.Check(rule, val); err != nil {
			return fmt.Errorf("assertion failed: %w", err)
		}
	}

	return nil
}
