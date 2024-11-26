create table if not exists public.todo_items
(
    id         serial
        primary key,
    task       varchar(255) not null,
    status     boolean,
    created_at timestamp default CURRENT_TIMESTAMP
);

alter table public.todo_items
    owner to postgres;

