CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL, -- To avoid privacy issues, we won't store email addresses, but we can use the username field for login purposes
    -- email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE conductors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- A conductor may not have a user account, but if they do, we want to cascade delete
    full_name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255), -- Optional field to store the location of the venue. If not provided, it defaults to the name of the venue.
    is_rotation_pool BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE, -- end_date can be null for single-day events
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE schedule_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    task_description VARCHAR(255) NOT NULL, -- This will store the description of the task (e.g., "Conduct field", "Maintenance", etc.). We don't use an enum here to allow for flexibility in task descriptions. We will return all possible task descriptions in the API response so the frontend can display them in a dropdown or similar UI element.
    conductor_id UUID REFERENCES conductors(id) ON DELETE SET NULL,
    venue_id UUID REFERENCES venues(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE cards (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    card_name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now(),
    zoom_level integer, -- Optional field to specify the zoom level for the map when this card is selected. This allows for better user experience by automatically adjusting the map view to fit the relevant blocks.
    center_coordinates jsonb, -- Optional field to specify the center coordinates for the map when this card is selected. This allows for better user experience by automatically centering the map on the relevant area when the card is selected.
    color text -- Optional field to specify a color for the card, which can be used in the frontend to visually distinguish different cards or to color-code them based on certain criteria (e.g., event type, priority, etc.)
);

CREATE TABLE blocks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    block_name text NOT NULL,
    geometry_type text NOT NULL,
    coordinates jsonb NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now(),
    card_id uuid REFERENCES cards(id) ON DELETE SET NULL,
    color text,
    CONSTRAINT blocks_geometry_type_check CHECK ((geometry_type = ANY (ARRAY['Polygon'::text, 'LineString'::text])))
);

CREATE TABLE schedule_entry_blocks (
    schedule_entry_id uuid REFERENCES schedule_entries(id) ON DELETE CASCADE,
    block_id uuid REFERENCES blocks(id) ON DELETE CASCADE,
    entry_subgroup_id uuid NOT NULL,
    PRIMARY KEY (schedule_entry_id, block_id)
);


CREATE INDEX idx_conductors_full_name_user_id ON conductors(full_name, user_id);
CREATE INDEX idx_venues_name ON venues(name);
CREATE INDEX idx_schedule_entries_date ON schedule_entries(schedule_date);
CREATE INDEX idx_schedule_entries_conductor_id ON schedule_entries(conductor_id);
CREATE INDEX idx_schedule_entries_venue_id ON schedule_entries(venue_id);
CREATE INDEX idx_blocks_block_name ON blocks(block_name);
CREATE INDEX idx_cards_card_name ON cards(card_name);
CREATE INDEX idx_schedule_entry_blocks_schedule_entry_id ON schedule_entry_blocks(schedule_entry_id);
CREATE INDEX idx_schedule_entry_blocks_schedule_entry_id_entry_subgroup_id ON schedule_entry_blocks(schedule_entry_id, entry_subgroup_id);
