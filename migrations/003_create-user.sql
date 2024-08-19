-- Write your migrate up statements here
create table users(
    email varchar(255) primary key,
    password text not null,
    username varchar(255) not null,
    role varchar(255) not null
);
---- create above / drop below ----
drop table users;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
