BEGIN;

CREATE TABLE tournament_prize_rules (
    tournament_id BIGINT NOT NULL REFERENCES tournaments (id),
    place SMALLINT NOT NULL,
    percentage DECIMAL(5,2) NOT NULL CHECK (percentage BETWEEN 0 AND 100),
    PRIMARY KEY (tournament_id, place)
);

COMMIT;
