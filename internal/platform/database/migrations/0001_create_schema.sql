CREATE TABLE categories (
    id          text PRIMARY KEY,
    title       text NOT NULL,
    description text NOT NULL,
    sort_order  int  NOT NULL
);

COMMENT ON TABLE categories IS 'Secciones de la lista de regalos';

CREATE TABLE gifts (
    id          text PRIMARY KEY,
    category_id text NOT NULL REFERENCES categories (id),
    name        text NOT NULL,
    variant     text,
    description text NOT NULL,
    required    int  NOT NULL CHECK (required > 0),
    selected    int  NOT NULL DEFAULT 0,
    sort_order  int  NOT NULL,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    -- Garantía final contra regalos repetidos, aunque falle la aplicación.
    CONSTRAINT gifts_selected_range CHECK (selected BETWEEN 0 AND required)
);

COMMENT ON TABLE gifts IS 'Regalos de la lista; selected = unidades ya reservadas';

CREATE INDEX gifts_category_id_idx ON gifts (category_id, sort_order);

CREATE TABLE reservations (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    gift_id      text NOT NULL REFERENCES gifts (id),
    quantity     int  NOT NULL CHECK (quantity > 0),
    created_at   timestamptz NOT NULL DEFAULT now(),
    cancelled_at timestamptz
);

COMMENT ON TABLE reservations IS 'Reservas anónimas de regalos; cancelled_at marca reservas anuladas';

CREATE INDEX reservations_gift_id_idx ON reservations (gift_id);
CREATE INDEX reservations_created_at_idx ON reservations (created_at DESC);
