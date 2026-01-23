#!/bin/bash

# Script para configurar y ejecutar tests en versaDumps

echo "======================================"
echo "  VersaDumps - Setup de Tests"
echo "======================================"
echo ""

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Función para imprimir con color
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Verificar que estamos en el directorio correcto
if [ ! -f "go.mod" ]; then
    print_error "Este script debe ejecutarse desde el directorio 'app'"
    exit 1
fi

print_success "Directorio correcto detectado"

# 1. Verificar Go
echo ""
echo "Verificando Go..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go encontrado: $GO_VERSION"
else
    print_error "Go no está instalado"
    exit 1
fi

# 2. Verificar Node.js
echo ""
echo "Verificando Node.js..."
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    print_success "Node.js encontrado: $NODE_VERSION"
else
    print_error "Node.js no está instalado"
    exit 1
fi

# 3. Ejecutar tests de Go
echo ""
echo "======================================"
echo "  Ejecutando Tests Go"
echo "======================================"
echo ""

go test ./... -v | tee test-results-go.log
GO_EXIT_CODE=$?

if [ $GO_EXIT_CODE -eq 0 ]; then
    print_success "Tests Go: PASARON"
else
    print_warning "Tests Go: ALGUNOS FALLARON (ver test-results-go.log)"
fi

# 4. Instalar dependencias de Vue si es necesario
echo ""
echo "======================================"
echo "  Configurando Tests Vue"
echo "======================================"
echo ""

cd frontend

if [ ! -d "node_modules" ]; then
    print_warning "node_modules no encontrado. Instalando dependencias..."
    npm install
    if [ $? -eq 0 ]; then
        print_success "Dependencias instaladas correctamente"
    else
        print_error "Error instalando dependencias"
        exit 1
    fi
else
    print_success "node_modules ya existe"
    
    # Verificar si necesitamos instalar las nuevas dependencias de testing
    if ! npm list vitest &> /dev/null; then
        print_warning "Dependencias de testing no encontradas. Instalando..."
        npm install
        if [ $? -eq 0 ]; then
            print_success "Dependencias de testing instaladas"
        else
            print_error "Error instalando dependencias de testing"
            exit 1
        fi
    else
        print_success "Dependencias de testing ya instaladas"
    fi
fi

# 5. Ejecutar tests de Vue
echo ""
echo "======================================"
echo "  Ejecutando Tests Vue"
echo "======================================"
echo ""

npm run test run | tee ../test-results-vue.log
VUE_EXIT_CODE=$?

if [ $VUE_EXIT_CODE -eq 0 ]; then
    print_success "Tests Vue: PASARON"
else
    print_warning "Tests Vue: ALGUNOS FALLARON (ver test-results-vue.log)"
fi

# Volver al directorio app
cd ..

# 6. Resumen final
echo ""
echo "======================================"
echo "  Resumen de Tests"
echo "======================================"
echo ""

if [ $GO_EXIT_CODE -eq 0 ]; then
    print_success "Tests Go: PASARON"
else
    print_warning "Tests Go: FALLARON"
fi

if [ $VUE_EXIT_CODE -eq 0 ]; then
    print_success "Tests Vue: PASARON"
else
    print_warning "Tests Vue: FALLARON"
fi

echo ""
echo "Logs guardados en:"
echo "  - test-results-go.log"
echo "  - test-results-vue.log"
echo ""

# 7. Generar reporte de coverage
echo "======================================"
echo "  Generando Reportes de Coverage"
echo "======================================"
echo ""

echo "Generando coverage Go..."
go test ./... -coverprofile=coverage.out -cover
if [ $? -eq 0 ]; then
    go tool cover -html=coverage.out -o coverage-go.html
    print_success "Coverage Go generado: coverage-go.html"
fi

echo ""
echo "Generando coverage Vue..."
cd frontend
npm run test:coverage &> /dev/null
if [ $? -eq 0 ]; then
    print_success "Coverage Vue generado en: frontend/coverage/"
fi
cd ..

echo ""
print_success "Setup y tests completados!"
echo ""
echo "Para ver los reportes HTML de coverage:"
echo "  Go:  abrir coverage-go.html en navegador"
echo "  Vue: abrir frontend/coverage/index.html en navegador"
echo ""
