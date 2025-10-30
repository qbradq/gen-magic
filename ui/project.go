package ui

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/qbradq/gen-magic/data"
	"github.com/qbradq/gen-magic/llm"
	_ "modernc.org/sqlite"
)

// Project holds all of the data for a project.
type Project struct {
	db *sql.DB
}

// WithDB executes a method with the database.

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
	// Construct schema
	if _, err = p.db.Exec(data.SchemaSQL); err != nil {
		log.Printf("error running schema script: %v\n", err)
		return err
	}
	// Lay down base data if needed
	if !p.BoolSetting("init.static-data-load.base", false) {
		if _, err := p.db.Exec(data.StaticDataSQL); err != nil {
			log.Printf("error running data script: %v\n", err)
			return err
		}
		p.SetBoolSetting("init.static-data-load.base", true)
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

// BoolSetting returns the given setting as a boolean or the default value.
func (p *Project) BoolSetting(key string, dv bool) bool {
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
	v, err := strconv.ParseBool(s)
	if err != nil {
		return dv
	}
	return v
}

// SetBoolSetting sets the given setting from a boolean value.
func (p *Project) SetBoolSetting(key string, v bool) error {
	return p.SetStringSetting(key, strconv.FormatBool(v))
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
		ORDER BY ID ASC
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
	ID int64
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
		ORDER BY id ASC
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
func (p *Project) GetLLM(id int64) *llm.LanguageModel {
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
	ret := &llm.LanguageModel{}
	err := row.Scan(&ret.ID, &ret.Name, &ret.API, &ret.APIEndpoint, &ret.APIKey, &ret.Model)
	if err != nil {
		panic(err)
		log.Fatalf("error getting LLM (scan): %v\n", err)
	}
	return ret
}

// SetLLM stores an LLM definition in the project.
func (p *Project) SetLLM(def *llm.LanguageModel) error {
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

// NewLLM returns a newly allocated LLM with a database ID.
func (p *Project) NewLLM() *llm.LanguageModel {
	ret := &llm.LanguageModel{
		Name: "Un-named LLM",
		API: "openrouter",
		APIEndpoint: "https://openrouter.ai/api/v1",
		Model: "meta-llama/llama-3.3-70b-instruct:free",
	}
	res, err := p.db.Exec(`
		INSERT INTO LLMs (name_txt, api, uri, model)
		VALUES (
			?,
			(SELECT id FROM APIs WHERE id_str = ?),
			?,
			?
		);
	`, ret.Name, ret.API, ret.APIEndpoint, ret.Model)
	if err != nil {
		log.Fatalf("error inserting new LLM definition: %v\n", err)
	}
	ret.ID, err = res.LastInsertId()
	if err != nil {
		log.Fatalf("error inserting new LLM definition ID: %v\n", err)
	}
	return ret
}

// DeleteLLM deletes the given LLM.
func (p *Project) DeleteLLM(def *llm.LanguageModel) {
	_, err := p.db.Exec(`
		DELETE FROM LLMs
		WHERE id = ?
		;
	`, def.ID)
	if err != nil {
		log.Fatalf("error deleting LLM definition: %v\n", err)
	}
}

// AgentName identifies an agent.
type AgentName struct {
	ID int64
	Name string
}

// ListAgents lists all agent definitions.
func (p *Project) ListAgents() []AgentName {
	ret := []AgentName{}
	rows, err := p.db.Query(`
		SELECT id, name_txt
		FROM Agents
		;
	`)
	if err != nil {
		log.Fatalf("error listing agents (select): %v\n", err)
	}
	for rows.Next() {
		name := AgentName{}
		if err := rows.Scan(&name.ID, &name.Name); err != nil {
			log.Fatalf("error listing agents (scan): %v\n", err)
		}
		ret = append(ret, name)
	}
	return ret
}

// GetAgent returns an agent by ID.
func (p *Project) GetAgent(id int64) *llm.Agent {
	ret := &llm.Agent{
		ID: id,
		System: llm.Message{
			Role: "system",
		},
	}
	row := p.db.QueryRow(`
		SELECT name_txt, llm, sys_prompt
		FROM Agents
		WHERE id = ?
		;
	`, id)
	var llmID int64
	if err := row.Scan(&ret.Name, &llmID, &ret.System.Content); err != nil {
		log.Fatalf("error getting agent (select): %v\n", err)
	}
	ret.LLM = p.GetLLM(llmID)
	return ret
}

// SetAgent sets the agent's information.
func (p *Project) SetAgent(agent *llm.Agent) {
	_, err := p.db.Exec(`
		UPDATE Agents
		SET
			name_txt = ?,
			llm = ?,
			sys_prompt = ?
		WHERE
			id = ?
		;
	`, agent.Name, agent.LLM.ID, agent.System.Content, agent.ID)
	if err != nil {
		log.Fatalf("error setting agent (update): %v\n", err)
	}
}

// NewAgent returns a new Agent object.
func (p *Project) NewAgent() *llm.Agent {
	llms := p.ListLLMs()
	ret := &llm.Agent{
		Name: "Unnamed Agent",
		LLM: p.GetLLM(llms[0].ID),
		System: llm.Message{
			Role: "system",
			Content: "You are a helpful AI assistant.",
		},
	}
	res, err := p.db.Exec(`
		INSERT INTO Agents (name_txt, llm, sys_prompt)
		VALUES (?, ?, ?)
		;
	`, ret.Name, ret.LLM.ID, ret.System.Content)
	if err != nil {
		log.Fatalf("error creating new agent (insert): %v\n", err)
	}
	ret.ID, err = res.LastInsertId()
	if err != nil {
		log.Fatalf("error creating new agent (last_id): %v\n", err)
	}
	return ret
}

// DeleteAgent deletes the given agent.
func (p *Project) DeleteAgent(agent *llm.Agent) {
	_, err := p.db.Exec(`
		DELETE FROM Agents
		WHERE id = ?
		;
	`, agent.ID)
	if err != nil {
		log.Fatalf("error deleting agent (delete): %v\n", err)
	}
}
