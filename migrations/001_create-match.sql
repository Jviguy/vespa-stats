-- Write your migrate up statements here

create table matches(
    id uuid primary key,
    teamA varchar(255) not null,
    teamB varchar(255) not null,
    league varchar(255) not null,
    winner int not null,
    scoreA int not null,
    scoreB int not null,
    date date not null,
    map varchar(255) not null
);

---- create above / drop below ----

drop table matches;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
