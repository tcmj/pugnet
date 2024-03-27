
INSERT INTO pug_data (thelink,description) VALUES 
('https://duckdb.org/docs/api/java.html','DuckDB Database Java API'),
('https://computingforgeeks.com/install-and-configure-dbeaver-on-ubuntu-debian/','How to install DBeaver'),
('https://dbeaver.io','Dbeaver Database Explorer');



INSERT INTO pug_label (parent_id,label,description) VALUES
(null,'tcmj','Root User Node');

INSERT INTO pug_label (parent_id,label,description) VALUES
((select id from pug_label where label = 'tcmj'),'development','Everything development related'),
((select id from pug_label where label = 'tcmj'),'software','Software'),
((select id from pug_label where label = 'tcmj'),'vip','special');


INSERT INTO pug_data_label (label_id,data_id) VALUES 
((select id from pug_label where label = 'vip'),(select id from pug_data where thelink like '%duckdb%')),
((select id from pug_label where label = 'development'),(select id from pug_data where thelink like '%computing%')),
((select id from pug_label where label = 'development'),(select id from pug_data where thelink like '%dbeaver.io%')),
((select id from pug_label where label = 'software'),(select id from pug_data where thelink like '%dbeaver.io%'));

