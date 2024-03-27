CREATE SEQUENCE IF NOT EXISTS pugdata_seq START 1;

CREATE SEQUENCE IF NOT EXISTS puglabel_seq START 100 INCREMENT BY 10;

CREATE TABLE IF NOT EXISTS pug_data(id int8 primary key DEFAULT nextval('pugdata_seq'), thelink VARCHAR NOT NULL, description VARCHAR, created TIMESTAMP NOT NULL DEFAULT current_timestamp, updated TIMESTAMP, deleted TIMESTAMP);

CREATE TABLE IF NOT EXISTS pug_label(id int8 primary key DEFAULT nextval('puglabel_seq'), parent_id int8, label VARCHAR NOT NULL, description VARCHAR, created TIMESTAMP NOT NULL DEFAULT current_timestamp, deleted TIMESTAMP);

CREATE TABLE IF NOT EXISTS pug_data_label(label_id int8, data_id int8, deleted TIMESTAMP, PRIMARY KEY (label_id, data_id), FOREIGN KEY (label_id) REFERENCES pug_label (id), FOREIGN KEY (data_id) REFERENCES pug_data (id) );

