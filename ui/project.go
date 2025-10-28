package ui

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/qbradq/gen-magic/llm"
	_ "modernc.org/sqlite"
)

// Project holds all of the data for a project.
type Project struct {
	db *sql.DB
}

// Connect connects to a data source, completely replacing any existing data.
func (p *Project) Load(driver, source string) error {
	var err error
	p.db, err = sql.Open(driver, source)
	if err != nil {
		return err
	}
	return p.dbInit()
}

// Close closes the data source.
func (p *Project) Close() error {
	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return err
		}
	}
	return nil
}

// dbInit initializes the database.
func (p *Project) dbInit() error {
	var err error
	// Settings
	if _, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS Settings (
			id TEXT PRIMARY KEY NOT NULL,
			val TEXT
		);
	`); err != nil {
		log.Printf("error creating table Settings: %v\n", err)
		return err
	}
	// LLM APIs
	if _, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS APIs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			id_str VARCHAR(32) UNIQUE NOT NULL,
			name_txt VARCHAR(64) NOT NULL
		);
	`); err != nil {
		log.Printf("error creating table APIs: %v\n", err)
		return err
	}
	row := p.db.QueryRow("SELECT COUNT(*) FROM APIs")
	var count int
	if err := row.Scan(&count); err != nil {
		log.Printf("error counting table APIs: %v\n", err)
		return err
	}
	if count == 0 {
		if _, err := p.db.Exec(`
			INSERT INTO APIs (id_str, name_txt)
			VALUES
				('openrouter', 'OpenRouter.ai')
			;
		`); err != nil {
			log.Printf("error populating table APIs: %v\n", err)
			return err
		}
	}
	// LLM definitions
	if _, err = p.db.Exec(`
		CREATE TABLE IF NOT EXISTS LLMs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name_txt VARCHAR(64),
			api INTEGER,
			uri VARCHAR(255),
			api_key VARCHAR(255),
			model VARCHAR(255),
			FOREIGN KEY (api) REFERENCES APIs(id)
		);
	`); err != nil {
		log.Printf("error creating table LLMs: %v\n", err)
		return err
	}
	row = p.db.QueryRow("SELECT COUNT(*) FROM LLMs")
	if err := row.Scan(&count); err != nil {
		log.Printf("error counting table LLMs: %v\n", err)
		return err
	}
	if count == 0 {
		if _, err := p.db.Exec(`
			INSERT INTO LLMs (name_txt, uri, model, api)
			VALUES
				(
					'OpenRouter.ai Llama 3 (Free)',
					'https://openrouter.ai/api/v1',
					'meta-llama/llama-3.3-70b-instruct:free',
					(SELECT id FROM APIs WHERE id_str = 'openrouter')
				)
			;
		`); err != nil {
			log.Printf("error populating table LLMs: %v\n", err)
			return err
		}
	}
	return nil
}

// StringSetting returns the given setting as a string or the default value.
func (p *Project) StringSetting(key, dv string) (ret string) {
	row := p.db.QueryRow(`
		SELECT
			IFNULL(val, '') AS val
		FROM Settings
		WHERE id = ?
		;
	`, key)
	if err := row.Scan(&ret); err != nil {
		return dv
	}
	return ret
}

// SetStringSetting sets the given setting from a string value.
func (p *Project) SetStringSetting(key, v string) error {
	_, err := p.db.Exec(`
		INSERT INTO Settings(id, val)
		VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET
			val = ?
		;
	`, key, v, v)
	return err
}

// IntSetting returns the given setting as an int or the default value.
func (p *Project) IntSetting(key string, dv int) int {
	var s string
	row := p.db.QueryRow(`
		SELECT
			IFNULL(val, '') AS val
		FROM Settings
		WHERE id = ?
		;
	`, key)
	if err := row.Scan(&s); err != nil {
		return dv
	}
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return dv
	}
	return int(v)
}

// SetIntSetting sets the given setting from an int value.
func (p *Project) SetIntSetting(key string, v int) error {
	return p.SetStringSetting(key, strconv.FormatInt(int64(v), 10))
}

// LLMApi wraps the information for one LLM API.
type LLMApi struct {
	ID string
	Name string
}

// ListAPIs lists all APIs available.
func (p *Project) ListAPIs() []LLMApi {
	ret := []LLMApi{}
	rows, err := p.db.Query(`
		SELECT
			id_str,
			name_txt
		FROM APIs
		;
	`)
	if err != nil {
		log.Printf("error listing APIs (query): %v\n", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		n := LLMApi{}
		if err := rows.Scan(&n.ID, &n.Name); err != nil {
			log.Printf("error listing APIs (scan): %v\n", err)
			return nil
		}
		ret = append(ret, n)
	}
	return ret
}

// LLMName names an LLM definition.
type LLMName struct {
	ID int
	Name string
}

// ListLLMs lists all LLM definitions.
func (p *Project) ListLLMs() []LLMName {
	ret := []LLMName{}
	rows, err := p.db.Query(`
		SELECT
			id,
			name_txt
		FROM LLMs
		;
	`)
	if err != nil {
		log.Printf("error listing LLMs (query): %v\n", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		n := LLMName{}
		if err := rows.Scan(&n.ID, &n.Name); err != nil {
			log.Printf("error listing LLMs (scan): %v\n", err)
			return nil
		}
		ret = append(ret, n)
	}
	return ret
}

// GetLLM returns an LLM definition from the project.
func (p *Project) GetLLM(id int) *llm.Definition {
	row := p.db.QueryRow(`
		SELECT
			LLMs.id,
			IFNULL(LLMs.name_txt, '') AS name_txt,
			IFNULL(APIs.id_str, '') AS id_str,
			IFNULL(LLMs.uri, '') AS uir,
			IFNULL(LLMs.api_key, '') AS api_key,
			IFNULL(LLMs.model, '') AS model
		FROM LLMs
		INNER JOIN APIs ON LLMs.api = APIs.id
		WHERE LLMs.id = ?
		;
	`, id)
	ret := &llm.Definition{}
	err := row.Scan(&ret.ID, &ret.Name, &ret.API, &ret.APIEndpoint, &ret.APIKey, &ret.Model)
	if err != nil {
		log.Printf("error getting LLM (scan): %v\n", err)
		return nil
	}
	return ret
}

// SetLLM stores an LLM definition in the project.
func (p *Project) SetLLM(def *llm.Definition) error {
	_, err := p.db.Exec(`
		UPDATE LLMs
		SET
			name_txt = ?,
			api = (SELECT id FROM APIs WHERE id_str = ?),
			uri = ?,
			api_key = ?,
			model = ?
		WHERE
			id = ?
		;
	`, def.Name, def.API, def.APIEndpoint, def.APIKey, def.Model, def.ID)
	return err
}
