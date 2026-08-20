ALTER INDEX special_results_competition_registrant_idx
    RENAME TO live_results_competition_registrant_idx;
ALTER INDEX special_dnf_attempts_round_registrant_idx
    RENAME TO live_dnf_attempts_round_registrant_idx;
ALTER INDEX special_dnf_attempts_competition_idx
    RENAME TO live_dnf_attempts_competition_idx;

ALTER TABLE special_results      RENAME TO live_results;
ALTER TABLE special_dnf_attempts RENAME TO live_dnf_attempts;
ALTER TABLE special_competitions RENAME TO live_tracked_competitions;
