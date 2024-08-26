-- +goose Up
CREATE TABLE public.duo_rooms (
    duo_room_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    user_id1 integer NOT NULL,
    user_id2 integer NOT NULL,
    room_id integer NOT NULL,
    CONSTRAINT duo_rooms_pkey PRIMARY KEY (duo_room_id),
    CONSTRAINT room_id FOREIGN KEY (room_id) REFERENCES public.rooms(room_id),
    CONSTRAINT user_id1_fk FOREIGN KEY (user_id1) REFERENCES public.users(user_id),
    CONSTRAINT user_id2_fk FOREIGN KEY (user_id2) REFERENCES public.users(user_id)
);

CREATE UNIQUE INDEX user1_user2_unique_index ON public.duo_rooms USING btree (user_id1, user_id2) WITH (deduplicate_items='true');

-- +goose Down
DROP TABLE IF EXISTS public.duo_rooms;