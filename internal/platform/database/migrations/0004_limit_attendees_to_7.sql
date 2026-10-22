-- El máximo de asistentes por confirmación pasa de 10 a 7.
-- NOT VALID: aplica a las filas nuevas o actualizadas sin revisar las que ya
-- existen, para que la migración no falle si alguna confirmación tiene más de 7.
ALTER TABLE attendance_confirmations
    DROP CONSTRAINT attendance_confirmations_attendees_check,
    ADD CONSTRAINT attendance_confirmations_attendees_check CHECK (attendees BETWEEN 1 AND 7) NOT VALID;
