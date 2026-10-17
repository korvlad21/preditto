BEGIN;

ALTER TABLE user_info
    DROP CONSTRAINT user_info_selected_tournament_id_fkey;

DROP TABLE tournaments;

COMMIT;
