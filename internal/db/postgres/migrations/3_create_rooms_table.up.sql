-- +goose Up
CREATE TABLE public.rooms (
    room_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    name character varying(50),
    is_group boolean NOT NULL,
    created_by_user_id integer NOT NULL,
    last_activity date DEFAULT now() NOT NULL,
    count_users integer DEFAULT 0 NOT NULL,
    CONSTRAINT rooms_pkey PRIMARY KEY (room_id),
    CONSTRAINT created_by_user_id_fk FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id)
);

CREATE INDEX room_index ON public.rooms USING btree (room_id) WITH (deduplicate_items='true');

-- +goose Down
DROP TABLE IF EXISTS public.rooms;