create table subjects
(
    id      int auto_increment,
    splanId int          null,
    name    varchar(128) null,
    constraint subjects_pk
        primary key (id),
    foreign key (splanId) references splan (id)
);

