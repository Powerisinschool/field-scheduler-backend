-- name: CreateVenue :one
INSERT INTO venues (
    name,
    location
) VALUES (
    $1, $2
) RETURNING id;

-- name: ListVenues :many
SELECT * FROM venues
ORDER BY name ASC;

-- name: GetVenueByName :one
SELECT id FROM venues
WHERE name ILIKE $1 LIMIT 1;
