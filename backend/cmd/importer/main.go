package main

// Importer CLI: parses the v7 HTML prototype and upserts seed data into Postgres.
//
// EXTRACTION STRATEGY (TODO: full implementation):
//   1. Read the HTML file into memory.
//   2. The JS seed objects (PEOPLE, ACTIVITIES, REGS, STAKEHOLDERS, CLIENT_PROFILES,
//      ACADEMIC_PUBLICATIONS, PHOTOS) are declared inside <script> tags as
//      `const NAME = { ... };` or `const NAME = [ ... ];`.
//   3. Locate each assignment with a regex anchored on `const <NAME> = ` and balance
//      braces/brackets to extract the literal. JS object keys are unquoted, single
//      quotes may be used, and trailing commas are legal — none of which is valid
//      JSON. Either:
//        (a) pre-process with a small tokenizer that quotes keys and normalises
//            quoted strings, then unmarshal with encoding/json; or
//        (b) shell out to `node -e 'console.log(JSON.stringify(OBJ))'` against a
//            stripped copy of the <script> block (simplest + most robust).
//   4. For each collection, map fields to the corresponding repo upsert (by stable
//      key: person name+dept, client slug, activity title+date, etc.).
//   5. For PHOTOS, upload to MinIO and set people.photo_key to the returned object
//      key, rather than storing the base64 blob in Postgres.
//
// For now this is a runnable skeleton that reads the file and logs what it sees.

import (
	"context"
	"flag"
	"os"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/config"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/db"
)

var seedNames = []string{
	"PEOPLE",
	"ACTIVITIES",
	"REGS",
	"STAKEHOLDERS",
	"CLIENT_PROFILES",
	"ACADEMIC_PUBLICATIONS",
	"PHOTOS",
}

func main() {
	_ = godotenv.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	path := flag.String("file", "../prototype/DCGG_Intelligence_Platform_v7.html", "path to v7 HTML prototype")
	dry := flag.Bool("dry", true, "log what would be imported without writing")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	raw, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal().Err(err).Str("path", *path).Msg("read prototype html")
	}
	log.Info().Int("bytes", len(raw)).Str("path", *path).Msg("prototype loaded")

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer pool.Close()

	for _, name := range seedNames {
		// Locate `const NAME = ` and report the offset — full brace-balanced
		// extraction happens in the TODO above.
		re := regexp.MustCompile(`(?m)const\s+` + regexp.QuoteMeta(name) + `\s*=\s*`)
		loc := re.FindIndex(raw)
		if loc == nil {
			log.Warn().Str("collection", name).Msg("seed block not found")
			continue
		}
		log.Info().
			Str("collection", name).
			Int("offset", loc[0]).
			Bool("dry", *dry).
			Msg("would import seed collection")
	}

	if *dry {
		log.Info().Msg("dry-run complete; no rows written")
		return
	}

	log.Warn().Msg("write path not implemented yet — see TODO header")
}
