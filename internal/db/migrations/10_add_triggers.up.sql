CREATE TRIGGER check_room_limit_trigger
BEFORE INSERT ON public.room_users_info
FOR EACH ROW EXECUTE FUNCTION public.check_room_limit();

CREATE TRIGGER update_room_last_activity_trigger
AFTER INSERT ON public.messages
FOR EACH ROW EXECUTE FUNCTION public.update_room_last_activity();