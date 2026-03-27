-- AWD Arena Platform - Initial Schema
-- PostgreSQL 17

-- Users
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(64)  NOT NULL UNIQUE,
    password    VARCHAR(256) NOT NULL,
    email       VARCHAR(128),
    role        VARCHAR(20)  NOT NULL DEFAULT 'player',
    team_id     BIGINT REFERENCES teams(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Teams
CREATE TABLE teams (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    token       VARCHAR(64)  NOT NULL UNIQUE,
    description TEXT,
    avatar_url  VARCHAR(256),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users(team_id);

-- Games
CREATE TABLE games (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(128) NOT NULL,
    description     TEXT,
    mode            VARCHAR(32)  NOT NULL DEFAULT 'awd_score',
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    round_duration  INTERVAL     NOT NULL DEFAULT '5 minutes',
    break_duration  INTERVAL     NOT NULL DEFAULT '2 minutes',
    total_rounds    INT          NOT NULL DEFAULT 20,
    current_round   INT          NOT NULL DEFAULT 0,
    current_phase   VARCHAR(20)  NOT NULL DEFAULT 'preparation',
    flag_format     VARCHAR(64)  NOT NULL DEFAULT 'flag{%s}',
    attack_weight   DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    defense_weight  DECIMAL(3,2) NOT NULL DEFAULT 0.5,
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    created_by      BIGINT       REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Game-Team junction
CREATE TABLE game_teams (
    game_id     BIGINT REFERENCES games(id),
    team_id     BIGINT REFERENCES teams(id),
    score       DECIMAL(10,2) NOT NULL DEFAULT 0,
    rank        INT,
    PRIMARY KEY (game_id, team_id)
);

CREATE INDEX idx_games_status ON games(status);

-- Challenges
CREATE TABLE challenges (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT       REFERENCES games(id),
    name        VARCHAR(128) NOT NULL,
    description TEXT,
    image_name  VARCHAR(256) NOT NULL,
    image_tag   VARCHAR(64)  DEFAULT 'latest',
    difficulty  VARCHAR(20)  DEFAULT 'medium',
    base_score  INT          NOT NULL DEFAULT 100,
    exposed_ports JSONB,
    cpu_limit   DECIMAL(3,2) DEFAULT 0.5,
    mem_limit   INT          DEFAULT 256,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Team Containers
CREATE TABLE team_containers (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    team_id         BIGINT REFERENCES teams(id),
    challenge_id    BIGINT REFERENCES challenges(id),
    container_id    VARCHAR(128),
    ip_address      VARCHAR(45),
    port_mapping    JSONB,
    status          VARCHAR(20)  NOT NULL DEFAULT 'creating',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, team_id, challenge_id)
);

-- Flag Records
CREATE TABLE flag_records (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT REFERENCES games(id),
    round       INT          NOT NULL,
    team_id     BIGINT REFERENCES teams(id),
    flag_hash   VARCHAR(256) NOT NULL,
    flag_value  VARCHAR(256) NOT NULL,
    service     VARCHAR(128) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, team_id, service)
);

-- Flag Submissions
CREATE TABLE flag_submissions (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    round           INT          NOT NULL,
    attacker_team   BIGINT REFERENCES teams(id),
    target_team     BIGINT REFERENCES teams(id),
    flag_value      VARCHAR(256) NOT NULL,
    is_correct      BOOLEAN      NOT NULL,
    points_earned   DECIMAL(10,2) NOT NULL DEFAULT 0,
    submitted_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, attacker_team, target_team, flag_value)
);

CREATE INDEX idx_submissions_game_round ON flag_submissions(game_id, round);

-- Round Scores
CREATE TABLE round_scores (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    round           INT          NOT NULL,
    team_id         BIGINT REFERENCES teams(id),
    attack_score    DECIMAL(10,2) NOT NULL DEFAULT 0,
    defense_score   DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_score     DECIMAL(10,2) NOT NULL DEFAULT 0,
    rank            INT,
    calculated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, team_id)
);

-- Event Logs
CREATE TABLE event_logs (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT,
    event_type  VARCHAR(64)  NOT NULL,
    level       VARCHAR(20)  NOT NULL DEFAULT 'info',
    team_id     BIGINT,
    detail      JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_game_type ON event_logs(game_id, event_type);
