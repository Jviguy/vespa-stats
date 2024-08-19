-- Write your migrate up statements here

create table players
(
    match_id uuid references matches(id),
    name varchar(255) not null,
    PRIMARY KEY(match_id, name),
    team varchar(255) not null,
    kills int not null,
    deaths int not null,
    assists int not null,
    entry_frags int not null,
    entry_fragged int not null,
    headshots int not null,
    objective int,
    rounds int not null,
    kost_rounds int not null,
    onevx int not null
);

---- create above / drop below ----

drop table players;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
