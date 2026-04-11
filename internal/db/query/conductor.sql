-- name: CreateConductor :one
INSERT INTO conductors (
    full_name
) VALUES (
    $1
) RETURNING id;

-- name: ListConductors :many
SELECT * FROM conductors
ORDER BY full_name ASC;

-- name: GetConductorByName :one
SELECT id FROM conductors
WHERE full_name ILIKE $1 LIMIT 1;
