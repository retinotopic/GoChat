CREATE FUNCTION public.check_room_limit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    current_user_count INTEGER;
    current_room_count INTEGER;
BEGIN
    SELECT count_rooms INTO current_user_count FROM users WHERE user_id = NEW.user_id FOR UPDATE;
    IF current_user_count >= 250 THEN
        RAISE EXCEPTION 'already 250 rooms for user';
    END IF;
        
        UPDATE users SET count_rooms = current_user_count + 1 WHERE user_id = NEW.user_id;

    SELECT count_users INTO current_room_count FROM rooms WHERE room_id = NEW.room_id FOR UPDATE;
    IF current_room_count >= 10 THEN
        RAISE EXCEPTION 'already 10 users in this room';
    END IF;
        
    UPDATE rooms SET count_users = current_room_count + 1 WHERE room_id = NEW.room_id;

    RETURN NEW;
END;
$$;