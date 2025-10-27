-- Revertir cambios de la migración 003

-- Eliminar índice
DROP INDEX IF EXISTS idx_users_tigerbeetle_account_id;

-- Eliminar campo tigerbeetle_account_id
ALTER TABLE users DROP COLUMN tigerbeetle_account_id;