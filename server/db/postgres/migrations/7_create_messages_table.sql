-- +goose Up
CREATE TABLE public.messages (
    message_id bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
    message_payload text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    user_id integer NOT NULL,
    room_id integer,
    CONSTRAINT messages_pkey PRIMARY KEY (message_id)
    -- CONSTRAINT users_rooms_info_fk FOREIGN KEY (user_id, room_id) REFERENCES public.room_users_info(user_id, room_id)
);
CREATE INDEX message_room_index ON public.messages USING btree (room_id) WITH (deduplicate_items='true');

-- +goose Down
DROP TABLE IF EXISTS public.messages;
