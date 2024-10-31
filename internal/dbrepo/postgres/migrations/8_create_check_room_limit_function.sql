-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION public.check_room_limit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$DECLARE
    current_user_count INTEGER;
    current_room_count INTEGER;
BEGIN	
	IF TG_OP = 'INSERT' THEN

        SELECT count_rooms INTO current_user_count FROM users WHERE user_id = NEW.user_id FOR UPDATE;
        IF current_user_count >= 250 THEN
            RAISE EXCEPTION 'User already has 250 rooms';
        END IF;
        
        UPDATE users SET count_rooms = current_user_count + 1 WHERE user_id = NEW.user_id;

        SELECT count_users INTO current_room_count FROM rooms WHERE room_id = NEW.room_id FOR UPDATE;
        IF current_room_count >= 10 THEN
            RAISE EXCEPTION 'Room already has 10 users';
        END IF;
        
        UPDATE rooms SET count_users = current_room_count + 1 WHERE room_id = NEW.room_id;

        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN

        SELECT count_rooms INTO current_user_count FROM users WHERE user_id = OLD.user_id FOR UPDATE;
        IF current_user_count <= 0 THEN
            RAISE EXCEPTION 'User has no rooms to delete';
        END IF;
        
        UPDATE users SET count_rooms = current_user_count - 1 WHERE user_id = OLD.user_id;

        SELECT count_users INTO current_room_count FROM rooms WHERE room_id = OLD.room_id FOR UPDATE;
        IF current_room_count <= 0 THEN
            RAISE EXCEPTION 'Room has no users to delete';
        END IF;
        
        UPDATE rooms SET count_users = current_room_count - 1 WHERE room_id = OLD.room_id;

        RETURN OLD;
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS public.check_room_limit();