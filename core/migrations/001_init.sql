create table runs (
    id               text primary key,
    created_at       text not null,
    updated_at       text not null,
    status           text not null,
    phase            text not null,
    advisor_mode     text not null,
    advisor_outcome  text not null,
    review_decision  text not null,
    source           text not null,
    case_id          text not null,
    benchmark_id     text not null,
    input_text       text not null,
    image            blob,
    image_mime       text not null,
    canaries_json    text not null,
    anonymized_text  text not null,
    vault_json       text not null,
    findings_json    text not null,
    card_json        text not null,
    draft            text not null,
    advice_json      text not null,
    final_anonymized text not null,
    final_text       text not null,
    judge_score      integer,
    judge_reason     text not null,
    gate2_passes     integer not null,
    escalated        integer not null,
    canaries_leaked  integer not null,
    error            text not null
);

create index runs_benchmark on runs (benchmark_id);

create table steps (
    id                text primary key,
    run_id            text not null references runs (id),
    seq               integer not null,
    kind              text not null,
    side              text not null,
    status            text not null,
    started_at        text not null,
    finished_at       text not null,
    model             text not null,
    prompt_tokens     integer not null,
    completion_tokens integer not null,
    input_json        text not null,
    output_json       text not null,
    error             text not null,
    unique (run_id, seq)
);

create table benchmarks (
    id         text primary key,
    created_at text not null,
    status     text not null,
    total      integer not null,
    finished   integer not null
);
