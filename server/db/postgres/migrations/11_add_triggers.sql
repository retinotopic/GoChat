-- +goose Up
CREATE OR REPLACE TRIGGER check_room_limit_trigger BEFORE INSERT OR DELETE ON public.room_users_info FOR EACH ROW EXECUTE FUNCTION public.check_room_limit();

CREATE TRIGGER update_room_last_activity_trigger AFTER INSERT ON public.messages FOR EACH ROW EXECUTE FUNCTION public.update_room_last_activity();

CREATE TRIGGER default_user_name BEFORE INSERT ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_user_name();

-- +goose Down
DROP TRIGGER IF EXISTS check_room_limit_trigger ON public.room_users_info;
DROP TRIGGER IF EXISTS update_room_last_activity_trigger ON public.messages;
DROP TRIGGER IF EXISTS default_username ON public.users;
