-- Agregar campo deleted_at a la tabla users para soft delete (si no existe)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'deleted_at') THEN
        ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Crear índice para deleted_at (si no existe)
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Actualizar índice de email para excluir registros eliminados
DROP INDEX IF EXISTS idx_users_email;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;