CREATE TABLE public.blocked_users (
    blocked_users_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    blocked_by_user_id integer NOT NULL,
    blocked_user_id integer NOT NULL,
    CONSTRAINT blocked_users_pkey PRIMARY KEY (blocked_users_id),
    CONSTRAINT user_id_fk1 FOREIGN KEY (blocked_by_user_id) REFERENCES public.users(user_id),
    CONSTRAINT user_id_fk2 FOREIGN KEY (blocked_user_id) REFERENCES public.users(user_id)
);

CREATE UNIQUE INDEX blocked_users_index ON public.blocked_users USING btree (blocked_by_user_id, blocked_user_id) WITH (deduplicate_items='true');