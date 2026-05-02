-- name: ListBlocks :many
SELECT * FROM blocks;

-- name: CreateBlock :one
INSERT INTO blocks (block_name, geometry_type, coordinates, color, card_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteBlockByID :exec
DELETE FROM blocks WHERE id = $1 OR block_name = $1;

-- name: DeleteBlocksByCardName :exec
DELETE FROM blocks USING cards WHERE blocks.card_id = cards.id AND cards.card_name = $1;
