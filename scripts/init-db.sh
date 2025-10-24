#!/bin/bash

# =============================================================================
# Script de Inicialización de Bases de Datos - Banca en Línea
# =============================================================================
# Este script inicializa todas las bases de datos necesarias para el proyecto
# - TigerBeetle: Base de datos de contabilidad de doble entrada
# - PostgreSQL: Base de datos relacional (se agregará más tarde)
# =============================================================================

set -e  # Salir si cualquier comando falla

echo "🚀 Iniciando configuración de bases de datos para Banca en Línea..."

# =============================================================================
# CONFIGURACIÓN DE TIGERBEETLE
# =============================================================================

echo "📊 Configurando TigerBeetle..."

# Directorio de datos de TigerBeetle
TIGERBEETLE_DATA_DIR="/data"
TIGERBEETLE_DATA_FILE="$TIGERBEETLE_DATA_DIR/0_0.tigerbeetle"

# Crear directorio de datos si no existe
if [ ! -d "$TIGERBEETLE_DATA_DIR" ]; then
    echo "📁 Creando directorio de datos: $TIGERBEETLE_DATA_DIR"
    mkdir -p "$TIGERBEETLE_DATA_DIR"
fi

# Formatear archivo de datos si no existe
if [ ! -f "$TIGERBEETLE_DATA_FILE" ]; then
    echo "🔧 Formateando archivo de datos de TigerBeetle..."
    tigerbeetle format \
        --cluster=0 \
        --replica=0 \
        --replica-count=1 \
        --development \
        "$TIGERBEETLE_DATA_FILE"
    
    echo "✅ Archivo de datos de TigerBeetle creado exitosamente"
else
    echo "ℹ️  Archivo de datos de TigerBeetle ya existe, omitiendo formateo"
fi

# Verificar que el archivo fue creado correctamente
if [ -f "$TIGERBEETLE_DATA_FILE" ]; then
    echo "✅ TigerBeetle configurado correctamente"
    echo "📄 Archivo de datos: $TIGERBEETLE_DATA_FILE"
    echo "📏 Tamaño: $(du -h "$TIGERBEETLE_DATA_FILE" | cut -f1)"
else
    echo "❌ Error: No se pudo crear el archivo de datos de TigerBeetle"
    exit 1
fi

# =============================================================================
# CONFIGURACIÓN DE POSTGRESQL (Placeholder para implementación futura)
# =============================================================================

echo "🐘 PostgreSQL será configurado en una futura actualización..."

# TODO: Agregar configuración de PostgreSQL
# - Crear base de datos
# - Ejecutar migraciones
# - Insertar datos de prueba

# =============================================================================
# FINALIZACIÓN
# =============================================================================

echo ""
echo "🎉 ¡Configuración de bases de datos completada!"
echo ""
echo "📋 Resumen:"
echo "  ✅ TigerBeetle: Configurado y listo"
echo "  ⏳ PostgreSQL: Pendiente de configuración"
echo ""
echo "🚀 Iniciando TigerBeetle en modo desarrollo..."

# Configurar variables de entorno para compatibilidad con Docker
export TIGERBEETLE_IO_MODE=blocking
export TIGERBEETLE_DISABLE_IO_URING=1

# Iniciar TigerBeetle
exec tigerbeetle start \
    --addresses=0.0.0.0:3002 \
    "$TIGERBEETLE_DATA_FILE"