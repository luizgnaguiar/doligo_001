-- 000002_create_identity_tables.down.sql
-- This script reverses the changes made in the corresponding 'up' migration.
-- It drops all tables related to the identity domain in the reverse order
-- of their creation to respect foreign key constraints.

-- Drop join tables first
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;

-- Drop the primary entity tables
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
