-- name: CreateRun :exec
insert into runs (
    id, created_at, updated_at, status, phase, advisor_mode, advisor_outcome, review_decision,
    source, case_id, benchmark_id, input_text, image, image_mime, canaries_json, anonymized_text,
    vault_json, findings_json, card_json, draft, advice_json, final_anonymized, final_text,
    judge_score, judge_reason, gate2_passes, escalated, canaries_leaked, error
) values (
    ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?
);

-- name: UpdateRun :exec
update runs set
    updated_at = ?, status = ?, phase = ?, advisor_outcome = ?, review_decision = ?,
    input_text = ?, anonymized_text = ?, vault_json = ?, findings_json = ?, card_json = ?,
    draft = ?, advice_json = ?, final_anonymized = ?, final_text = ?, judge_score = ?,
    judge_reason = ?, gate2_passes = ?, escalated = ?, canaries_leaked = ?, error = ?
where id = ?;

-- name: GetRun :one
select * from runs where id = ?;

-- name: ListRuns :many
select id, created_at, updated_at, status, phase, advisor_mode, advisor_outcome, review_decision,
    source, case_id, benchmark_id, input_text, length(image) as image_size, image_mime,
    canaries_json, anonymized_text, vault_json, findings_json, card_json, draft, advice_json,
    final_anonymized, final_text, judge_score, judge_reason, gate2_passes, escalated,
    canaries_leaked, error
from runs
where id < ?
order by id desc
limit ?;

-- name: ListRunsByBenchmark :many
select id, created_at, updated_at, status, phase, advisor_mode, advisor_outcome, review_decision,
    source, case_id, benchmark_id, input_text, length(image) as image_size, image_mime,
    canaries_json, anonymized_text, vault_json, findings_json, card_json, draft, advice_json,
    final_anonymized, final_text, judge_score, judge_reason, gate2_passes, escalated,
    canaries_leaked, error
from runs
where benchmark_id = ?
order by id asc;
