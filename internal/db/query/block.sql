-- name: ListBlocks :many
SELECT * FROM blocks;

-- name: CreateBlock :one
INSERT INTO blocks (block_name, geometry_type, coordinates, color, card_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
