CREATE TABLE IF NOT EXISTS gitRepo (
    id SERIAL PRIMARY KEY NOT NULL ,
    login VARCHAR(50) NOT NULL,
    commit_count INTEGER DEFAULT 0
);

INSERT INTO gitRepo (login , commit_count) VALUES ('Andrey' , 1 );