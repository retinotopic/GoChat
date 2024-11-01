-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION public.set_username() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN
    IF NEW.username IS NULL THEN
        NEW.username := 'user' || NEW.user_id::text;
    END IF;
    RETURN NEW;
END;$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS public.set_username();