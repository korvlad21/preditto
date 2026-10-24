BEGIN;

CREATE TABLE participant_stage_rankings (
    participant_id BIGINT NOT NULL REFERENCES tournament_participants (id),
    stage_id BIGINT NOT NULL REFERENCES tournament_stages (id),
    stage_position INT NOT NULL CHECK (stage_position > 0),
    cumulative_position INT NOT NULL CHECK (cumulative_position > 0),
    stage_points INT NOT NULL CHECK (stage_points >= 0),
    cumulative_points INT NOT NULL CHECK (cumulative_points >= 0),
    stage_exact_scores INT NOT NULL CHECK (stage_exact_scores >= 0),
    cumulative_exact_scores INT NOT NULL CHECK (cumulative_exact_scores >= 0),
    stage_outcomes INT NOT NULL CHECK (stage_outcomes >= 0),
    cumulative_outcomes INT NOT NULL CHECK (cumulative_outcomes >= 0),
    calculated_at TIMESTAMP(6) NOT NULL,
    CONSTRAINT pk_participant_stage_rankings PRIMARY KEY (participant_id, stage_id)
);

CREATE INDEX idx_rankings_stage_position ON participant_stage_rankings (stage_id, stage_position);
CREATE INDEX idx_rankings_cumulative_position ON participant_stage_rankings (stage_id, cumulative_position);

COMMIT;
