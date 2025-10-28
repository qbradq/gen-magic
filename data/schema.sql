/*******************************************************************************
* schema.sql
*
* Database schema for Gen Magic
*******************************************************************************/

-- User settings
CREATE TABLE IF NOT EXISTS Settings (
    id TEXT PRIMARY KEY NOT NULL,
    val TEXT
);

-- API info
CREATE TABLE IF NOT EXISTS APIs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    id_str VARCHAR(32) UNIQUE NOT NULL,
    name_txt VARCHAR(64) NOT NULL
);

-- LLM definitions
CREATE TABLE IF NOT EXISTS LLMs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name_txt VARCHAR(64),
    api INTEGER,
    uri VARCHAR(255),
    api_key VARCHAR(255),
    model VARCHAR(255),
    FOREIGN KEY (api) REFERENCES APIs(id)
);

