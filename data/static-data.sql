/*******************************************************************************
* static-data.sql
*
* Database data that should always exist
*******************************************************************************/

-- List of all supported APIs
INSERT INTO APIs (id_str, name_txt)
VALUES (
    'openrouter',
    'OpenRouter.ai'
);

-- Default LLM definition
INSERT INTO LLMs (name_txt, uri, model, api)
VALUES (
    'OpenRouter.ai Llama-3.3-70b (free)',
    'https://openrouter.ai/api/v1',
    'meta-llama/llama-3.3-70b-instruct:free',
    (SELECT id FROM APIs WHERE id_str = 'openrouter')
);

-- Default agent definition
INSERT INTO Agents (name_txt, llm, sys_prompt)
VALUES (
    'Agent Red',
    (SELECT id FROM LLMs ORDER BY id ASC LIMIT 1),
    'You are a helpful AI assistant'
);

-- Mark this version loaded
INSERT INTO Settings (id, val)
VALUES
    ('init.static-data-load.base', 'true')
;
