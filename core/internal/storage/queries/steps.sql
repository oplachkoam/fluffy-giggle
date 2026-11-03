-- name: CreateStep :exec
insert into steps (
    id, run_id, seq, kind, side, status, started_at, finished_at, model,
    prompt_tokens, completion_tokens, input_json, output_json, error
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListStepsByRun :many
select * from steps where run_id = ? order by seq asc;
