CREATE TABLE attendance_confirmations (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL CHECK (char_length(name) BETWEEN 2 AND 80),
    attendees  int  NOT NULL CHECK (attendees BETWEEN 1 AND 10),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE attendance_confirmations IS 'Confirmaciones de asistencia al baby shower; attendees incluye a quien confirma';

CREATE INDEX attendance_confirmations_created_at_idx ON attendance_confirmations (created_at DESC);
