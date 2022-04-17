create database stundenplan DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
use stundenplan;

create table Employee
(
    id   varchar(14)  null,
    name varchar(128) null
);

create table Module
(
    id      int auto_increment,
    splanId int         not null,
    name    varchar(128) null,
    constraint module_pk
        primary key (id)
);

create table Lecture
(
    id       int auto_increment,
    day char(2) not null,
    start    time null,
    end      time null,
    moduleId int  not null,
    constraint Lecture_pk
        primary key (id),
    constraint Lecture_Module_id_fk
        foreign key (moduleId) references Module (id)
);

create table Lecture_by_Employee
(
    lectureId  int         null,
    employeeId varchar(14) null
);

create table Lecture_in_Room
(
    roomId    varchar(14) null,
    lectureId int         null
);

create table Room
(
    id   varchar(14) null,
    name varchar(126) null
);

