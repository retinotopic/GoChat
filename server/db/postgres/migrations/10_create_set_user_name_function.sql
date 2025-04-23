-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION public.set_user_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN
    IF NEW.user_name IS NULL THEN
        NEW.user_name := 'user' || NEW.user_id::text;
    END IF;
    RETURN NEW;
END;$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS public.set_user_name();
