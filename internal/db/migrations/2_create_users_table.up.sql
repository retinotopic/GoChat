CREATE TABLE public.users (
    user_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    subject character varying(50) NOT NULL,
    username character varying(255) NOT NULL,
    allow_group_invites boolean DEFAULT true NOT NULL,
    allow_direct_messages boolean DEFAULT true NOT NULL,
    count_rooms integer DEFAULT 0 NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (user_id),
    CONSTRAINT username_unique UNIQUE (username)
);

CREATE INDEX trgm_idx_users_username ON public.users USING gin (username public.gin_trgm_ops);