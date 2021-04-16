create database test;
use test;


create table user(
                     id bigint primary key auto_increment,
                     name nvarchar(20)
);

create table message(
                        id bigint primary key auto_increment,
                        user_id bigint,
                        msg_content TEXT
)