-- +goose Up
CREATE TABLE public.room_users_info (
    room_user_info_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    room_id integer NOT NULL,
    user_id integer NOT NULL,
    CONSTRAINT room_users_info_pkey PRIMARY KEY (room_user_info_id),
    CONSTRAINT room_id_fk FOREIGN KEY (room_id) REFERENCES public.rooms(room_id),
    CONSTRAINT users_id_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);
CREATE UNIQUE INDEX user_room_unique_index ON public.room_users_info USING btree (user_id, room_id) WITH (deduplicate_items='true');
CREATE INDEX rui_room_index ON public.room_users_info USING btree (room_id) WITH (deduplicate_items='true');

-- +goose Down
DROP TABLE IF EXISTS public.room_users_info;