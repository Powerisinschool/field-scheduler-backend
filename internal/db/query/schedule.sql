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

-- name: GetOrCreateCard :one
INSERT INTO cards (card_name)
VALUES ($1)
ON CONFLICT (card_name) DO UPDATE SET card_name = EXCLUDED.card_name
RETURNING *;

-- -- name: CreateBlock :one
-- INSERT INTO blocks (
--     block_name, geometry_type, coordinates, card_id
-- ) VALUES (
--     $1, $2, $3, $4
-- ) RETURNING *;

-- name: ListUniqueTaskDescriptions :many
SELECT DISTINCT task_description FROM schedule_entries
ORDER BY task_description ASC;
