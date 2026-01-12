---------------------------
-- USERS
---------------------------
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) UNIQUE NOT NULL,
    firstname       VARCHAR(255) NOT NULL,
    lastname        VARCHAR(255) NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

---------------------------
-- WORKFLOWS
---------------------------
CREATE TABLE IF NOT EXISTS workflows (
    id               SERIAL PRIMARY KEY,
    user_id          INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    trigger_type     VARCHAR(64) NOT NULL,
    trigger_config   JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_url       TEXT NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

UPDATE workflows
SET enabled = FALSE
WHERE trigger_type <> 'manual' AND enabled = TRUE;

---------------------------
-- WORKFLOW RUNS
---------------------------
CREATE TABLE IF NOT EXISTS workflow_runs (
    id           SERIAL PRIMARY KEY,
    workflow_id  INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    status       VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ,
    error        TEXT
);

---------------------------
-- JOBS
---------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('pending', 'processing', 'succeeded', 'failed');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS jobs (
    id           SERIAL PRIMARY KEY,
    workflow_id  INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    run_id       INTEGER NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    payload      JSONB,
    status       job_status NOT NULL DEFAULT 'pending',
    error        TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created_at ON jobs (status, created_at);

---------------------------
-- GOOGLE TOKENS
---------------------------
CREATE TABLE IF NOT EXISTS google_tokens (
    id             SERIAL PRIMARY KEY,
    user_id        INTEGER REFERENCES users(id) ON DELETE CASCADE,
    access_token   TEXT NOT NULL,
    refresh_token  TEXT NOT NULL,
    expiry         TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

---------------------------
-- GITHUB TOKENS
---------------------------
CREATE TABLE IF NOT EXISTS github_tokens (
    id             SERIAL PRIMARY KEY,
    user_id        INTEGER REFERENCES users(id) ON DELETE CASCADE,
    access_token   TEXT NOT NULL,
    token_type     TEXT NOT NULL,
    scope          TEXT NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

---------------------------
-- AREA CATALOG
---------------------------
CREATE TABLE IF NOT EXISTS area_services (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    more_info    TEXT,
    oauth_scopes JSONB,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS area_capabilities (
    service_id      TEXT NOT NULL REFERENCES area_services(id) ON DELETE CASCADE,
    id              TEXT NOT NULL,
    kind            TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    action_url      TEXT,
    default_payload JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (service_id, id)
);

CREATE INDEX IF NOT EXISTS idx_area_capabilities_service_kind ON area_capabilities (service_id, kind);

CREATE TABLE IF NOT EXISTS area_fields (
    id             SERIAL PRIMARY KEY,
    service_id     TEXT NOT NULL,
    capability_id  TEXT NOT NULL,
    key            TEXT NOT NULL,
    type           TEXT NOT NULL,
    required       BOOLEAN NOT NULL DEFAULT FALSE,
    description    TEXT,
    example        JSONB,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (service_id, capability_id) REFERENCES area_capabilities(service_id, id) ON DELETE CASCADE
);

---------------------------
-- AREA CATALOG SEED
---------------------------
INSERT INTO area_services (id, name, enabled, more_info, oauth_scopes)
VALUES
    ('core', 'Core', TRUE, NULL, NULL),
    ('discord', 'Discord', TRUE, NULL, NULL),
    ('google', 'Google', TRUE, NULL, '["https://www.googleapis.com/auth/gmail.send","https://www.googleapis.com/auth/calendar.events","https://www.googleapis.com/auth/userinfo.email"]'::jsonb),
    ('github', 'GitHub', TRUE, NULL, NULL),
    ('slack', 'Slack', TRUE, NULL, NULL),
    ('notion', 'Notion', TRUE, NULL, NULL),
    ('weather', 'Weather', TRUE, NULL, NULL),
    ('steam', 'Steam', TRUE, NULL, NULL),
    ('crypto', 'Crypto', TRUE, NULL, NULL),
    ('nasa', 'NASA', TRUE, NULL, NULL),
    ('air_quality', 'Air Quality', TRUE, NULL, NULL),
    ('trello', 'Trello', TRUE, NULL, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO area_capabilities (id, service_id, kind, name, description, action_url, default_payload)
VALUES
    ('manual', 'core', 'trigger', 'Manual trigger', 'Trigger launched manually from the UI.', NULL, NULL),
    ('interval', 'core', 'trigger', 'Timer (interval)', 'Runs every N minutes.', NULL, NULL),
    ('gmail_inbound', 'google', 'trigger', 'When a Gmail is received', 'Triggers on new unread messages in Gmail inbox.', NULL, NULL),
    ('github_commit', 'github', 'trigger', 'When a GitHub commit is pushed', 'Triggers on new commits on a branch.', NULL, NULL),
    ('github_pull_request', 'github', 'trigger', 'When a GitHub pull request changes', 'Triggers on PR updates (opened/closed/merged).', NULL, NULL),
    ('github_issue', 'github', 'trigger', 'When a GitHub issue changes', 'Triggers on issue updates (opened/closed/reopened).', NULL, NULL),
    ('weather_temp', 'weather', 'trigger', 'When temperature crosses a threshold', 'Triggers when current temperature crosses above/below a threshold.', NULL, NULL),
    ('weather_report', 'weather', 'trigger', 'Weather report (interval)', 'Sends current weather for a city every X minutes.', NULL, NULL),
    ('reddit_new_post', 'core', 'trigger', 'Reddit new post', 'Triggers when a new post appears in a subreddit.', NULL, NULL),
    ('youtube_new_video', 'core', 'trigger', 'YouTube new video', 'Triggers when a channel publishes a new video.', NULL, NULL),
    ('steam_player_online', 'steam', 'trigger', 'Steam player online', 'Triggers when a Steam user becomes online.', NULL, NULL),
    ('steam_game_sale', 'steam', 'trigger', 'Steam game on sale', 'Triggers when a game goes on sale.', NULL, NULL),
    ('steam_price_change', 'steam', 'trigger', 'Steam price change', 'Triggers when a game price changes.', NULL, NULL),
    ('crypto_price_threshold', 'crypto', 'trigger', 'Crypto price threshold', 'Triggers when a crypto price crosses a threshold.', NULL, NULL),
    ('crypto_percent_change', 'crypto', 'trigger', 'Crypto percent change', 'Triggers when a crypto changes by a % over 1h or 24h.', NULL, NULL),
    ('nasa_apod', 'nasa', 'trigger', 'NASA APOD', 'Triggers when the Astronomy Picture of the Day updates.', NULL, NULL),
    ('nasa_mars_photo', 'nasa', 'trigger', 'NASA Mars rover photo', 'Triggers on latest Mars rover photos.', NULL, NULL),
    ('nasa_neo_close_approach', 'nasa', 'trigger', 'NASA NEO close approach', 'Triggers when a near-earth object passes within a distance.', NULL, NULL),
    ('air_quality_aqi_threshold', 'air_quality', 'trigger', 'Air Quality AQI threshold', 'Triggers when AQI crosses a threshold.', NULL, NULL),
    ('air_quality_pm25_threshold', 'air_quality', 'trigger', 'Air Quality PM2.5 threshold', 'Triggers when PM2.5 crosses a threshold.', NULL, NULL),

    ('trello_create_card', 'trello', 'reaction', 'Create card', 'Create a Trello card in a list.', '/actions/trello/card', NULL),
    ('trello_move_card', 'trello', 'reaction', 'Move card', 'Move a Trello card to another list.', '/actions/trello/card/move', NULL),
    ('trello_create_list', 'trello', 'reaction', 'Create list', 'Create a Trello list on a board.', '/actions/trello/list', NULL),

    ('discord_message', 'discord', 'reaction', 'Send message', 'Send a message to a channel using the bot.', '/actions/discord/message', '{"content":"Hello from Area"}'::jsonb),
    ('discord_embed', 'discord', 'reaction', 'Send embed', 'Send an embed to a channel.', '/actions/discord/embed', '{"title":"Area update","description":"Something happened"}'::jsonb),
    ('discord_edit_message', 'discord', 'reaction', 'Edit message', 'Edit a previously sent message.', '/actions/discord/message/edit', NULL),
    ('discord_delete_message', 'discord', 'reaction', 'Delete message', 'Delete a message by ID.', '/actions/discord/message/delete', NULL),
    ('discord_add_reaction', 'discord', 'reaction', 'Add reaction', 'Add a reaction emoji to a message.', '/actions/discord/message/react', NULL),

    ('google_gmail_send', 'google', 'reaction', 'Send Gmail', 'Send an email from the authenticated Google account.', '/actions/google/email', NULL),
    ('google_calendar_event', 'google', 'reaction', 'Create Calendar event', 'Create an event in the primary calendar.', '/actions/google/calendar', NULL),

    ('github_issue', 'github', 'reaction', 'Create issue', 'Create a new issue in a repository.', '/actions/github/issue', NULL),
    ('github_pull_request', 'github', 'reaction', 'Create pull request', 'Create a pull request from a branch.', '/actions/github/pr', NULL),

    ('slack_message', 'slack', 'reaction', 'Send message', 'Send a message to a Slack channel.', '/actions/slack/message', '{"text":"Hello from Area"}'::jsonb),
    ('slack_blocks', 'slack', 'reaction', 'Send blocks message', 'Send a message with Block Kit payload.', '/actions/slack/blocks', NULL),
    ('slack_update', 'slack', 'reaction', 'Update message', 'Update an existing message.', '/actions/slack/message/update', NULL),
    ('slack_delete', 'slack', 'reaction', 'Delete message', 'Delete a message by timestamp.', '/actions/slack/message/delete', NULL),
    ('slack_reaction', 'slack', 'reaction', 'Add reaction', 'Add an emoji reaction to a message.', '/actions/slack/message/react', NULL),

    ('notion_create_page', 'notion', 'reaction', 'Create page', 'Create a new Notion page.', '/actions/notion/page', NULL),
    ('notion_append_blocks', 'notion', 'reaction', 'Append blocks', 'Append blocks to a page or block.', '/actions/notion/blocks', NULL),
    ('notion_create_database_row', 'notion', 'reaction', 'Create database row', 'Create a new page in a database.', '/actions/notion/database', NULL),
    ('notion_update_page', 'notion', 'reaction', 'Update page', 'Update page properties.', '/actions/notion/page/update', NULL),

    ('http_webhook', 'core', 'reaction', 'HTTP POST', 'Send raw JSON to a custom HTTP endpoint.', NULL, NULL)
ON CONFLICT (service_id, id) DO NOTHING;

INSERT INTO area_fields (service_id, capability_id, key, type, required, description, example)
VALUES
    ('core', 'interval', 'interval_minutes', 'number', TRUE, 'Delay between runs in minutes', '5'::jsonb),

    ('github', 'github_commit', 'token_id', 'number', TRUE, 'Stored GitHub token id', NULL),
    ('github', 'github_commit', 'repo', 'string', TRUE, 'Repository in owner/name format', '"owner/repo"'::jsonb),
    ('github', 'github_commit', 'branch', 'string', TRUE, 'Branch to watch', '"main"'::jsonb),

    ('github', 'github_pull_request', 'token_id', 'number', TRUE, 'Stored GitHub token id', NULL),
    ('github', 'github_pull_request', 'repo', 'string', TRUE, 'Repository in owner/name format', '"owner/repo"'::jsonb),
    ('github', 'github_pull_request', 'actions', 'array<string>', FALSE, 'Actions to watch (opened,closed,merged)', '["opened","closed","merged"]'::jsonb),

    ('github', 'github_issue', 'token_id', 'number', TRUE, 'Stored GitHub token id', NULL),
    ('github', 'github_issue', 'repo', 'string', TRUE, 'Repository in owner/name format', '"owner/repo"'::jsonb),
    ('github', 'github_issue', 'actions', 'array<string>', FALSE, 'Actions to watch (opened,closed,reopened)', '["opened","closed"]'::jsonb),

    ('weather', 'weather_temp', 'city', 'string', TRUE, 'City name (e.g. Paris)', '"Paris"'::jsonb),
    ('weather', 'weather_temp', 'threshold', 'number', TRUE, 'Temperature threshold (°C)', '20'::jsonb),
    ('weather', 'weather_temp', 'direction', 'string', TRUE, 'above or below', '"above"'::jsonb),
    ('weather', 'weather_temp', 'interval_minutes', 'number', FALSE, 'Minimum polling interval in minutes', '5'::jsonb),

    ('weather', 'weather_report', 'city', 'string', TRUE, 'City name (e.g. Paris)', '"Paris"'::jsonb),
    ('weather', 'weather_report', 'interval_minutes', 'number', TRUE, 'Polling interval in minutes', '10'::jsonb),

    ('core', 'reddit_new_post', 'subreddit', 'string', TRUE, 'Subreddit name (without r/)', '"golang"'::jsonb),
    ('core', 'reddit_new_post', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '5'::jsonb),

    ('core', 'youtube_new_video', 'channel', 'string', TRUE, 'YouTube channel name, handle, or ID', '"@GoogleDevelopers"'::jsonb),
    ('core', 'youtube_new_video', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '5'::jsonb),

    ('steam', 'steam_player_online', 'steam_id', 'string', TRUE, 'SteamID64 of the user', '"76561198000000000"'::jsonb),
    ('steam', 'steam_player_online', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '5'::jsonb),

    ('steam', 'steam_game_sale', 'app_id', 'number', TRUE, 'Steam app ID', '570'::jsonb),
    ('steam', 'steam_game_sale', 'country', 'string', FALSE, 'Country code for pricing (e.g. us, fr)', '"us"'::jsonb),
    ('steam', 'steam_game_sale', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '10'::jsonb),

    ('steam', 'steam_price_change', 'app_id', 'number', TRUE, 'Steam app ID', '570'::jsonb),
    ('steam', 'steam_price_change', 'country', 'string', FALSE, 'Country code for pricing (e.g. us, fr)', '"us"'::jsonb),
    ('steam', 'steam_price_change', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '10'::jsonb),

    ('crypto', 'crypto_price_threshold', 'coin_id', 'string', TRUE, 'Coin id (CoinGecko, e.g. bitcoin)', '"bitcoin"'::jsonb),
    ('crypto', 'crypto_price_threshold', 'currency', 'string', FALSE, 'Currency (e.g. usd, eur)', '"usd"'::jsonb),
    ('crypto', 'crypto_price_threshold', 'threshold', 'number', TRUE, 'Price threshold', '50000'::jsonb),
    ('crypto', 'crypto_price_threshold', 'direction', 'string', TRUE, 'above or below', '"above"'::jsonb),
    ('crypto', 'crypto_price_threshold', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '5'::jsonb),

    ('crypto', 'crypto_percent_change', 'coin_id', 'string', TRUE, 'Coin id (CoinGecko, e.g. bitcoin)', '"bitcoin"'::jsonb),
    ('crypto', 'crypto_percent_change', 'currency', 'string', FALSE, 'Currency (e.g. usd, eur)', '"usd"'::jsonb),
    ('crypto', 'crypto_percent_change', 'percent', 'number', TRUE, 'Percent change threshold', '5'::jsonb),
    ('crypto', 'crypto_percent_change', 'period', 'string', TRUE, '1h or 24h', '"1h"'::jsonb),
    ('crypto', 'crypto_percent_change', 'direction', 'string', FALSE, 'above, below, or any', '"any"'::jsonb),
    ('crypto', 'crypto_percent_change', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '5'::jsonb),

    ('nasa', 'nasa_apod', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '60'::jsonb),

    ('nasa', 'nasa_mars_photo', 'rover', 'string', TRUE, 'Rover name (curiosity, perseverance, opportunity, spirit)', '"curiosity"'::jsonb),
    ('nasa', 'nasa_mars_photo', 'camera', 'string', FALSE, 'Camera name (e.g. FHAZ, RHAZ, NAVCAM)', '"FHAZ"'::jsonb),
    ('nasa', 'nasa_mars_photo', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '60'::jsonb),

    ('nasa', 'nasa_neo_close_approach', 'threshold_km', 'number', TRUE, 'Distance threshold in km', '500000'::jsonb),
    ('nasa', 'nasa_neo_close_approach', 'days_ahead', 'number', FALSE, 'Number of days to look ahead', '1'::jsonb),
    ('nasa', 'nasa_neo_close_approach', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '60'::jsonb),

    ('air_quality', 'air_quality_aqi_threshold', 'city', 'string', TRUE, 'City name (e.g. Paris)', '"Paris"'::jsonb),
    ('air_quality', 'air_quality_aqi_threshold', 'index', 'string', FALSE, 'AQI index (us_aqi or european_aqi)', '"us_aqi"'::jsonb),
    ('air_quality', 'air_quality_aqi_threshold', 'threshold', 'number', TRUE, 'AQI threshold', '100'::jsonb),
    ('air_quality', 'air_quality_aqi_threshold', 'direction', 'string', TRUE, 'above or below', '"above"'::jsonb),
    ('air_quality', 'air_quality_aqi_threshold', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '10'::jsonb),

    ('air_quality', 'air_quality_pm25_threshold', 'city', 'string', TRUE, 'City name (e.g. Paris)', '"Paris"'::jsonb),
    ('air_quality', 'air_quality_pm25_threshold', 'threshold', 'number', TRUE, 'PM2.5 threshold (µg/m³)', '15'::jsonb),
    ('air_quality', 'air_quality_pm25_threshold', 'direction', 'string', TRUE, 'above or below', '"above"'::jsonb),
    ('air_quality', 'air_quality_pm25_threshold', 'interval_minutes', 'number', FALSE, 'Polling interval in minutes', '10'::jsonb),

    ('trello', 'trello_create_card', 'api_key', 'string', TRUE, 'Trello API key', NULL),
    ('trello', 'trello_create_card', 'token', 'string', TRUE, 'Trello user token', NULL),
    ('trello', 'trello_create_card', 'list_id', 'string', TRUE, 'Target list ID', NULL),
    ('trello', 'trello_create_card', 'name', 'string', TRUE, 'Card name', NULL),
    ('trello', 'trello_create_card', 'desc', 'string', FALSE, 'Card description', NULL),
    ('trello', 'trello_create_card', 'pos', 'string', FALSE, 'Card position (top, bottom, or numeric)', '"bottom"'::jsonb),

    ('trello', 'trello_move_card', 'api_key', 'string', TRUE, 'Trello API key', NULL),
    ('trello', 'trello_move_card', 'token', 'string', TRUE, 'Trello user token', NULL),
    ('trello', 'trello_move_card', 'card_id', 'string', TRUE, 'Card ID to move', NULL),
    ('trello', 'trello_move_card', 'list_id', 'string', TRUE, 'Destination list ID', NULL),
    ('trello', 'trello_move_card', 'pos', 'string', FALSE, 'Card position (top, bottom, or numeric)', '"bottom"'::jsonb),

    ('trello', 'trello_create_list', 'api_key', 'string', TRUE, 'Trello API key', NULL),
    ('trello', 'trello_create_list', 'token', 'string', TRUE, 'Trello user token', NULL),
    ('trello', 'trello_create_list', 'board_id', 'string', TRUE, 'Board ID', NULL),
    ('trello', 'trello_create_list', 'name', 'string', TRUE, 'List name', NULL),
    ('trello', 'trello_create_list', 'pos', 'string', FALSE, 'List position (top, bottom, or numeric)', '"bottom"'::jsonb),

    ('discord', 'discord_message', 'channel_id', 'string', TRUE, 'Target channel ID', '"123456789012345678"'::jsonb),
    ('discord', 'discord_message', 'content', 'string', TRUE, 'Message content', '"Hello from Area"'::jsonb),
    ('discord', 'discord_message', 'bot_token', 'string', FALSE, 'Discord bot token', NULL),

    ('discord', 'discord_embed', 'channel_id', 'string', TRUE, 'Target channel ID', '"123456789012345678"'::jsonb),
    ('discord', 'discord_embed', 'title', 'string', TRUE, 'Embed title', NULL),
    ('discord', 'discord_embed', 'description', 'string', TRUE, 'Embed description', NULL),
    ('discord', 'discord_embed', 'url', 'string', FALSE, 'Embed URL', NULL),
    ('discord', 'discord_embed', 'color', 'string', FALSE, 'Hex color (e.g. #5865F2)', NULL),
    ('discord', 'discord_embed', 'content', 'string', FALSE, 'Optional message content', NULL),
    ('discord', 'discord_embed', 'bot_token', 'string', FALSE, 'Discord bot token', NULL),

    ('discord', 'discord_edit_message', 'channel_id', 'string', TRUE, 'Target channel ID', NULL),
    ('discord', 'discord_edit_message', 'message_id', 'string', TRUE, 'Message ID to edit', NULL),
    ('discord', 'discord_edit_message', 'content', 'string', TRUE, 'New message content', NULL),
    ('discord', 'discord_edit_message', 'bot_token', 'string', FALSE, 'Discord bot token', NULL),

    ('discord', 'discord_delete_message', 'channel_id', 'string', TRUE, 'Target channel ID', NULL),
    ('discord', 'discord_delete_message', 'message_id', 'string', TRUE, 'Message ID to delete', NULL),
    ('discord', 'discord_delete_message', 'bot_token', 'string', FALSE, 'Discord bot token', NULL),

    ('discord', 'discord_add_reaction', 'channel_id', 'string', TRUE, 'Target channel ID', NULL),
    ('discord', 'discord_add_reaction', 'message_id', 'string', TRUE, 'Message ID to react to', NULL),
    ('discord', 'discord_add_reaction', 'emoji', 'string', TRUE, 'Emoji (e.g. 😀 or name:id)', NULL),
    ('discord', 'discord_add_reaction', 'bot_token', 'string', FALSE, 'Discord bot token', NULL),

    ('google', 'google_gmail_send', 'token_id', 'number', TRUE, 'Stored Google token id', NULL),
    ('google', 'google_gmail_send', 'to', 'string', TRUE, 'Recipient email', '"dest@example.com"'::jsonb),
    ('google', 'google_gmail_send', 'subject', 'string', TRUE, 'Email subject', '"Hello"'::jsonb),
    ('google', 'google_gmail_send', 'body', 'string', TRUE, 'Email body', '"Hello from Area"'::jsonb),

    ('google', 'google_calendar_event', 'token_id', 'number', TRUE, 'Stored Google token id', NULL),
    ('google', 'google_calendar_event', 'summary', 'string', TRUE, 'Event title', '"Area event"'::jsonb),
    ('google', 'google_calendar_event', 'start', 'string', TRUE, 'Start datetime (RFC3339)', '"2025-12-09T14:00:00Z"'::jsonb),
    ('google', 'google_calendar_event', 'end', 'string', TRUE, 'End datetime (RFC3339)', '"2025-12-09T15:00:00Z"'::jsonb),
    ('google', 'google_calendar_event', 'attendees', 'array<string>', FALSE, 'Attendee emails', '["a@example.com"]'::jsonb),

    ('github', 'github_issue', 'token_id', 'number', TRUE, 'Stored GitHub token id', NULL),
    ('github', 'github_issue', 'repo', 'string', TRUE, 'Repository in owner/name format', '"owner/repo"'::jsonb),
    ('github', 'github_issue', 'title', 'string', TRUE, 'Issue title', NULL),
    ('github', 'github_issue', 'body', 'string', FALSE, 'Issue body', NULL),
    ('github', 'github_issue', 'labels', 'array<string>', FALSE, 'Labels to add', NULL),

    ('github', 'github_pull_request', 'token_id', 'number', TRUE, 'Stored GitHub token id', NULL),
    ('github', 'github_pull_request', 'repo', 'string', TRUE, 'Repository in owner/name format', '"owner/repo"'::jsonb),
    ('github', 'github_pull_request', 'title', 'string', TRUE, 'Pull request title', NULL),
    ('github', 'github_pull_request', 'head', 'string', TRUE, 'Source branch (or owner:branch)', '"feature-branch"'::jsonb),
    ('github', 'github_pull_request', 'base', 'string', TRUE, 'Base branch', '"main"'::jsonb),
    ('github', 'github_pull_request', 'body', 'string', FALSE, 'Pull request body', NULL),

    ('slack', 'slack_message', 'channel_id', 'string', TRUE, 'Target channel ID', '"C1234567890"'::jsonb),
    ('slack', 'slack_message', 'text', 'string', TRUE, 'Message text', '"Hello from Area"'::jsonb),
    ('slack', 'slack_message', 'bot_token', 'string', FALSE, 'Slack bot token', NULL),

    ('slack', 'slack_blocks', 'channel_id', 'string', TRUE, 'Target channel ID', '"C1234567890"'::jsonb),
    ('slack', 'slack_blocks', 'text', 'string', FALSE, 'Fallback text', NULL),
    ('slack', 'slack_blocks', 'blocks', 'array<object>', TRUE, 'Block Kit JSON array', NULL),
    ('slack', 'slack_blocks', 'bot_token', 'string', FALSE, 'Slack bot token', NULL),

    ('slack', 'slack_update', 'channel_id', 'string', TRUE, 'Target channel ID', '"C1234567890"'::jsonb),
    ('slack', 'slack_update', 'message_ts', 'string', TRUE, 'Message timestamp', NULL),
    ('slack', 'slack_update', 'text', 'string', TRUE, 'New message text', NULL),
    ('slack', 'slack_update', 'bot_token', 'string', FALSE, 'Slack bot token', NULL),

    ('slack', 'slack_delete', 'channel_id', 'string', TRUE, 'Target channel ID', '"C1234567890"'::jsonb),
    ('slack', 'slack_delete', 'message_ts', 'string', TRUE, 'Message timestamp', NULL),
    ('slack', 'slack_delete', 'bot_token', 'string', FALSE, 'Slack bot token', NULL),

    ('slack', 'slack_reaction', 'channel_id', 'string', TRUE, 'Target channel ID', '"C1234567890"'::jsonb),
    ('slack', 'slack_reaction', 'message_ts', 'string', TRUE, 'Message timestamp', NULL),
    ('slack', 'slack_reaction', 'emoji', 'string', TRUE, 'Emoji', NULL),
    ('slack', 'slack_reaction', 'bot_token', 'string', FALSE, 'Slack bot token', NULL),

    ('notion', 'notion_create_page', 'parent_page_id', 'string', TRUE, 'Parent page ID', NULL),
    ('notion', 'notion_create_page', 'title', 'string', TRUE, 'Page title', NULL),
    ('notion', 'notion_create_page', 'content', 'string', FALSE, 'Page content', NULL),
    ('notion', 'notion_create_page', 'blocks', 'array<object>', FALSE, 'Optional blocks (JSON array)', NULL),
    ('notion', 'notion_create_page', 'bot_token', 'string', FALSE, 'Notion token', NULL),

    ('notion', 'notion_append_blocks', 'block_id', 'string', TRUE, 'Block ID', NULL),
    ('notion', 'notion_append_blocks', 'blocks', 'array<object>', TRUE, 'Blocks JSON array', NULL),
    ('notion', 'notion_append_blocks', 'bot_token', 'string', FALSE, 'Notion token', NULL),

    ('notion', 'notion_create_database_row', 'database_id', 'string', TRUE, 'Database ID', NULL),
    ('notion', 'notion_create_database_row', 'properties', 'object', TRUE, 'Notion properties JSON object', NULL),
    ('notion', 'notion_create_database_row', 'children', 'array<object>', FALSE, 'Optional blocks (JSON array)', NULL),
    ('notion', 'notion_create_database_row', 'bot_token', 'string', FALSE, 'Notion token', NULL),

    ('notion', 'notion_update_page', 'page_id', 'string', TRUE, 'Page ID', NULL),
    ('notion', 'notion_update_page', 'properties', 'object', TRUE, 'Notion properties JSON object', NULL),
    ('notion', 'notion_update_page', 'bot_token', 'string', FALSE, 'Notion token', NULL),

    ('core', 'http_webhook', 'url', 'string', TRUE, 'Target URL', '"https://example.com/hook"'::jsonb),
    ('core', 'http_webhook', 'payload', 'object', FALSE, 'JSON payload to send', NULL);
