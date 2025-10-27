-- Revertir cambios de la migración 002

-- Eliminar índices creados
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tigerbeetle_account_id;

-- Recrear índices originales sin filtro WHERE
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tigerbeetle_account_id ON users(tigerbeetle_account_id);

-- Eliminar campo deleted_at
ALTER TABLE users DROP COLUMN deleted_at;