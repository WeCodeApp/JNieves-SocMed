-- +goose Up
CREATE DATABASE IF NOT EXISTS school;

USE school;

-- Enable the UUID function for MySQL
SET @@global.sql_mode = 'STRICT_TRANS_TABLES';

-- Table: teachers
CREATE TABLE teachers (
                          id CHAR(36) PRIMARY KEY, -- Store UUIDs as CHAR(36)
                          first_name VARCHAR(255) NOT NULL,
                          last_name VARCHAR(255) NOT NULL,
                          subject VARCHAR(255) NOT NULL,
                          class VARCHAR(255) NOT NULL,
                          email VARCHAR(255) NOT NULL UNIQUE
);

-- Table: students
CREATE TABLE students (
                          id CHAR(36) PRIMARY KEY, -- Store UUIDs as CHAR(36)
                          first_name VARCHAR(255) NOT NULL,
                          last_name VARCHAR(255) NOT NULL,
                          class VARCHAR(255) NOT NULL,
                          email VARCHAR(255) NOT NULL UNIQUE
);

-- Table: executives
CREATE TABLE executives (
                            id CHAR(36) PRIMARY KEY, -- Store UUIDs as CHAR(36)
                            first_name VARCHAR(255) NOT NULL,
                            last_name VARCHAR(255) NOT NULL,
                            email VARCHAR(255) NOT NULL UNIQUE,
                            username VARCHAR(255) NOT NULL UNIQUE,
                            password CHAR(60) NOT NULL, -- To store bcrypt-hashed passwords
                            last_password_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            user_creation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                            password_reset_token VARCHAR(255),
                            reset_token_expiry TIMESTAMP,
                            role VARCHAR(255) NOT NULL,
                            user_inactive BOOLEAN DEFAULT FALSE
);
