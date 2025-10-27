-- Eliminar trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Eliminar función
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Eliminar índices
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tigerbeetle_account_id;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Eliminar tabla
DROP TABLE IF EXISTS users;