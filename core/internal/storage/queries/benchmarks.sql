-- name: CreateBenchmark :exec
insert into benchmarks (id, created_at, status, total, finished) values (?, ?, ?, ?, ?);

-- name: UpdateBenchmark :exec
update benchmarks set status = ?, finished = ? where id = ?;

-- name: GetBenchmark :one
select * from benchmarks where id = ?;

-- name: ListBenchmarks :many
select * from benchmarks order by id desc;
