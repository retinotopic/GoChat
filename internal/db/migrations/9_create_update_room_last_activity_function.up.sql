-- +goose Up
CREATE FUNCTION public.update_room_last_activity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE rooms
    SET last_activity = NOW()
    WHERE room_id = NEW.room_id AND NOW() - last_activity > INTERVAL '24 hours';
    
    RETURN NEW;
END;
$$;

-- +goose Down
DROP FUNCTION IF EXISTS public.update_room_last_activity();