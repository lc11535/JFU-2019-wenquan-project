DROP DATABASE IF EXISTS project;
DROP TABLE IF EXISTS user_info;
DROP TABLE IF EXISTS file_store;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS school_info;
DROP TABLE IF EXISTS notifications;
CREATE DATABASE project;
CREATE TABLE user_info (
	alias VARCHAR(20) NOT NULL UNIQUE,
	pwd VARCHAR(20) NOT NULL,
	school VARCHAR(20) NOT NULL,
	academy VARCHAR(20) NOT NULL,
	major VARCHAR(10) NOT NULL,
	grade VARCHAR(10) NOT NULL,
	auth SMALLINT NOT NULL,
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT
)ENGINE=innodb;
CREATE TABLE file_store(
	filename VARCHAR(200) NOT NULL,
	pointer VARCHAR(400) NOT NULL,
	academy VARCHAR(20) NOT NULL,
	major VARCHAR(10) NOT NULL,
	course VARCHAR(30) NOT NULL,
	description TEXT NOT NULL,
	auth SMALLINT NOT NULL,
	uploader VARCHAR(20) NOT NULL,
	upload_date VARCHAR(12) NOT NULL,
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	download_time INT NOT NULL DEFAULT 0，
	size INT NOT NULL,
	FOREIGN KEY(uploader) REFERENCES user_info(alias)
)ENGINE=innodb;
CREATE TABLE comments(
	file_id INT NOT NULL,
	comment_time VARCHAR(12) NOT NULL,
	commentator VARCHAR(20) NOT NULL,
	content VARCHAR(200),
	FOREIGN KEY(commentator) REFERENCES user_info(alias),
	FOREIGN KEY(file_id) REFERENCES file_store(id)
)ENGINE=innodb;
CREATE TABLE school_info(
	school VARCHAR(20) NOT NULL,
	academy VARCHAR(20) NOT NULL,
	major VARCHAR(10) NOT NULL,
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT
)ENGINE=innodb;
CREATE TABLE notifications(
	school VARCHAR(20) NOT NULL,
	upload_date VARCHAR(12) NOT NULL,
	title VARCHAR(100) NOT NULL,
	content TEXT
);