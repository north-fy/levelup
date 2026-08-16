CREATE TABLE branches (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(512) NOT NULL DEFAULT '',
    color       VARCHAR(20) NOT NULL DEFAULT '',
    icon        VARCHAR(50) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_branches_user_id ON branches (user_id);

CREATE TABLE quests (
    id             BIGSERIAL PRIMARY KEY,
    branch_id      BIGINT NOT NULL REFERENCES branches (id) ON DELETE CASCADE,
    user_id        BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title          VARCHAR(255) NOT NULL,
    description    VARCHAR(1024) NOT NULL DEFAULT '',
    type           VARCHAR(20) NOT NULL,
    reward_xp      INTEGER NOT NULL DEFAULT 0,
    reward_gold    INTEGER NOT NULL DEFAULT 0,
    duration_hours INTEGER NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'todo',
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quests_branch_id ON quests (branch_id);
CREATE INDEX idx_quests_user_id ON quests (user_id);
CREATE INDEX idx_quests_status ON quests (status);