-- name: InsertRuntimeProbe :exec
INSERT INTO platform_runtime_probes (
    run_id,
    probe_name
) VALUES (
    $1,
    $2
);

-- name: CountRuntimeProbesByRunID :one
SELECT COUNT(*)
FROM platform_runtime_probes
WHERE run_id = $1;

-- name: DeleteRuntimeProbesByRunID :exec
DELETE FROM platform_runtime_probes
WHERE run_id = $1;
