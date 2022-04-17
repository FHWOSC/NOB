create table timeslots
(
    id        int auto_increment,
    subjectId int     null,
    start     time    null,
    end       time    null,
    day       char(2) null,
    constraint timeslots_subjects_id_fk
        foreign key (subjectId) references subjects (id)
);

