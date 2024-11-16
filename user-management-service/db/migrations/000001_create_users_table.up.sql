CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- CreateTable
CREATE TABLE IF NOT EXISTS "users"
(
    "ID"            UUID         NOT NULL UNIQUE DEFAULT uuid_generate_v4(),
    "email"         VARCHAR(50)  NOT NULL PRIMARY KEY,
    "name"          VARCHAR(100) NOT NULL,
    "password"      VARCHAR      NOT NULL,
    "createdAtUTC"  TIMESTAMP(3) NOT NULL        DEFAULT NOW(),
    "updatedAtUTC"  TIMESTAMP(3) NOT NULL        DEFAULT NOW(),
    "deletedAtUTC"  TIMESTAMP(3) NOT NULL        DEFAULT NOW()
);
