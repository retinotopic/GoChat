-- +goose Up
CREATE TABLE public.users (
    user_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    subject character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    allow_group_invites boolean DEFAULT true NOT NULL,
    allow_direct_messages boolean DEFAULT true NOT NULL,
    count_rooms integer DEFAULT 0 NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (user_id),
    CONSTRAINT name_unique UNIQUE (name)
);

CREATE INDEX trgm_idx_users_name ON public.users USING gin (name public.gin_trgm_ops);

-- +goose Down
DROP TABLE IF EXISTS public.users;