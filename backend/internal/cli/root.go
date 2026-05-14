package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/api"
	"github.com/IsKenKenYa/Commory/backend/internal/config"
	"github.com/IsKenKenYa/Commory/backend/internal/importers"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/query"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

const cliUserID = "cli-user"

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	switch args[0] {
	case "serve":
		return runServe(ctx, stdout)
	case "import":
		if len(args) != 2 {
			return fmt.Errorf("usage: commory import <path>")
		}
		return runImport(ctx, args[1], stdout)
	case "validate":
		if len(args) != 2 {
			return fmt.Errorf("usage: commory validate <path>")
		}
		return runValidate(args[1], stdout)
	case "query":
		return runQuery(ctx, args[1:], stdout)
	case "timeline":
		return runTimeline(ctx, args[1:], stdout)
	case "contacts":
		return runContacts(ctx, args[1:], stdout)
	case "export":
		return runExport(ctx, args[1:], stdout)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "commory <serve|import|validate|query|timeline|contacts|export>")
}

func runServe(ctx context.Context, stdout io.Writer) error {
	cfg, store, validator, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	server := api.NewServer(cfg, store, validator)
	httpServer := &http.Server{Addr: cfg.ListenAddr, Handler: server.Handler()}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = api.Shutdown(shutdownCtx, httpServer)
	}()
	_, _ = fmt.Fprintf(stdout, "Commory listening on %s\n", cfg.ListenAddr)
	return httpServer.ListenAndServe()
}

func runImport(ctx context.Context, path string, stdout io.Writer) error {
	_, store, validator, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	importer := importers.JSONImporter{}
	export, raw, err := importer.Import(path)
	if err != nil {
		return err
	}
	if err := validator.ValidateBytes(raw); err != nil {
		return err
	}
	importID, err := store.Import(ctx, cliUserID, path, export, raw)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "imported %s as %s\n", path, importID)
	return nil
}

func runValidate(path string, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	validator, err := msglayer.NewValidator(cfg.SchemaRoot)
	if err != nil {
		return err
	}
	if err := validator.ValidateFile(path); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, "valid")
	return nil
}

func runQuery(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("query", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var params msglayer.SearchParams
	params.UserID = cliUserID
	params.Limit = 100
	flags.StringVar(&params.ContactID, "contact", "", "")
	flags.StringVar(&params.Type, "type", "", "")
	flags.StringVar(&params.From, "from", "", "")
	flags.StringVar(&params.To, "to", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if rest := flags.Args(); len(rest) > 0 {
		params.Keyword = rest[0]
	}
	_, store, _, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	items, err := query.New(store).Search(ctx, params)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(items)
}

func runTimeline(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("timeline", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var params msglayer.SearchParams
	params.UserID = cliUserID
	params.Limit = 100
	flags.StringVar(&params.From, "from", "", "")
	flags.StringVar(&params.To, "to", "", "")
	flags.StringVar(&params.Type, "type", "", "")
	flags.StringVar(&params.Participant, "participant", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	_, store, _, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	items, err := query.New(store).Timeline(ctx, params)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(items)
}

func runContacts(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 || args[0] != "list" {
		return fmt.Errorf("usage: commory contacts list")
	}
	_, store, _, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	items, err := query.New(store).Identities(ctx, cliUserID)
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(items)
}

func runExport(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var format string
	var importID string
	flags.StringVar(&format, "format", "json", "")
	flags.StringVar(&importID, "import-id", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if format != "json" {
		return fmt.Errorf("unsupported export format: %s", format)
	}
	_, store, _, err := boot(ctx)
	if err != nil {
		return err
	}
	defer store.Close()
	raw, err := store.ExportImport(ctx, cliUserID, importID)
	if err != nil {
		return err
	}
	_, err = stdout.Write(raw)
	return err
}

func boot(ctx context.Context) (config.Config, storage.Provider, *msglayer.Validator, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, nil, nil, err
	}
	validator, err := msglayer.NewValidator(cfg.SchemaRoot)
	if err != nil {
		return config.Config{}, nil, nil, err
	}
	store, err := openStore(cfg)
	if err != nil {
		return config.Config{}, nil, nil, err
	}
	if err := store.Init(ctx); err != nil {
		return config.Config{}, nil, nil, err
	}
	return cfg, store, validator, nil
}

func openStore(cfg config.Config) (storage.Provider, error) {
	switch cfg.Driver {
	case "sqlite":
		return storage.NewSQLiteProvider(cfg.DSN)
	case "postgres":
		dsn := cfg.DSN
		if cfg.DatabaseURL != "" {
			dsn = cfg.DatabaseURL
		}
		if dsn == "" {
			dsn = "./data/commory-postgres-store.json"
		}
		return storage.NewPostgresProvider(dsn)
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.Driver)
	}
}
