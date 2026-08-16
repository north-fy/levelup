CREATE TABLE roadmaps (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description VARCHAR(1024) NOT NULL DEFAULT '',
    source_type VARCHAR(20) NOT NULL DEFAULT 'own',
    source_id   BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_roadmaps_user_id ON roadmaps (user_id);

CREATE TABLE roadmap_nodes (
    id             BIGSERIAL PRIMARY KEY,
    roadmap_id     BIGINT NOT NULL REFERENCES roadmaps (id) ON DELETE CASCADE,
    title          VARCHAR(255) NOT NULL,
    description    VARCHAR(1024) NOT NULL DEFAULT '',
    position       INTEGER NOT NULL DEFAULT 0,
    type           VARCHAR(20) NOT NULL,
    reward_xp      INTEGER NOT NULL DEFAULT 0,
    reward_gold    INTEGER NOT NULL DEFAULT 0,
    duration_hours INTEGER NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'todo',
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_roadmap_nodes_roadmap_id ON roadmap_nodes (roadmap_id);

CREATE TABLE roadmap_edges (
    id           BIGSERIAL PRIMARY KEY,
    roadmap_id   BIGINT NOT NULL REFERENCES roadmaps (id) ON DELETE CASCADE,
    from_node_id BIGINT NOT NULL REFERENCES roadmap_nodes (id) ON DELETE CASCADE,
    to_node_id   BIGINT NOT NULL REFERENCES roadmap_nodes (id) ON DELETE CASCADE
);

CREATE INDEX idx_roadmap_edges_roadmap_id ON roadmap_edges (roadmap_id);
CREATE INDEX idx_roadmap_edges_from ON roadmap_edges (from_node_id);
CREATE INDEX idx_roadmap_edges_to ON roadmap_edges (to_node_id);

CREATE TABLE workshop_roadmaps (
    id                BIGSERIAL PRIMARY KEY,
    author_id         BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    source_roadmap_id BIGINT NOT NULL REFERENCES roadmaps (id) ON DELETE CASCADE,
    title             VARCHAR(255) NOT NULL,
    description       VARCHAR(1024) NOT NULL DEFAULT '',
    is_published      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_workshop_roadmaps_author_id ON workshop_roadmaps (author_id);
CREATE INDEX idx_workshop_roadmaps_published ON workshop_roadmaps (is_published);