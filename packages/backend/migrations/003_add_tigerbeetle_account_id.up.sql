-- Agregar campo tigerbeetle_account_id a la tabla users (si no existe)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'tigerbeetle_account_id') THEN
        ALTER TABLE users ADD COLUMN tigerbeetle_account_id BIGINT UNIQUE;
    END IF;
END $$;

-- Crear índice para tigerbeetle_account_id (si no existe)
CREATE INDEX IF NOT EXISTS idx_users_tigerbeetle_account_id ON users(tigerbeetle_account_id) WHERE deleted_at IS NULL;