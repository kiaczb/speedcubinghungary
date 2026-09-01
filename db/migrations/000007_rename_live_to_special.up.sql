-- A funkció neve "Budapest Special" lett, ezért a táblák is special_* előtagot
-- kapnak. A "live" szó a kódban ezentúl KIZÁRÓLAG a külső WCA Live szolgáltatást
-- jelenti — ezért az oszlopnevek (live_result_id, live_round_id,
-- live_competition_id) szándékosan változatlanok: azok valóban WCA Live
-- azonosítók.
--
-- ALTER ... RENAME-et használunk (nem újra-létrehozást), így a már felvett
-- emberi megfigyelések (elapsed_cs) és az audit nyom sértetlenül megmarad.

ALTER TABLE live_tracked_competitions RENAME TO special_competitions;
ALTER TABLE live_dnf_attempts         RENAME TO special_dnf_attempts;
ALTER TABLE live_results              RENAME TO special_results;

ALTER INDEX live_dnf_attempts_competition_idx
    RENAME TO special_dnf_attempts_competition_idx;
ALTER INDEX live_dnf_attempts_round_registrant_idx
    RENAME TO special_dnf_attempts_round_registrant_idx;
ALTER INDEX live_results_competition_registrant_idx
    RENAME TO special_results_competition_registrant_idx;
