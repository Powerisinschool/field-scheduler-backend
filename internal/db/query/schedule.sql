-- name: CreateScheduleEntry :one
INSERT INTO schedule_entries (
    schedule_date, start_time, task_description, conductor_id, venue_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListScheduleEntries :many
SELECT * FROM schedule_entries
WHERE schedule_date >= sqlc.arg('start_date') 
  AND schedule_date <= sqlc.arg('end_date')
ORDER BY schedule_date ASC, start_time ASC;

-- name: DeleteScheduleEntriesByDateRange :exec
DELETE FROM schedule_entries
WHERE schedule_date >= $1 AND schedule_date <= $2;

-- -- name: GetScheduleConductorByName :one
-- SELECT id FROM conductors
-- WHERE full_name ILIKE $1 LIMIT 1;

-- -- name: CreateScheduleConductor :one
-- INSERT INTO conductors (
--     full_name
-- ) VALUES (
--     $1
-- ) RETURNING id;

-- name: ListUniqueTaskDescriptions :many
SELECT DISTINCT task_description FROM schedule_entries
ORDER BY task_description ASC;
