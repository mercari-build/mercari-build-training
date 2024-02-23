create table items (
    id integer primary key,
    name String not null,
    category_id integer  not null,
    image_name String
);

create table categories (
    id integer primary key,
    category String not null
);