# Guía Rápida de Tests - VersaDumps

## 🚀 Ejecución Rápida

### Opción 1: Script Automatizado (Recomendado)

#### Windows (PowerShell)
```powershell
cd app
.\run-tests.ps1
```

#### Linux/Mac (Bash)
```bash
cd app
chmod +x run-tests.sh
./run-tests.sh
```

Este script:
- ✅ Verifica que Go y Node.js estén instalados
- ✅ Instala dependencias de Vue si es necesario
- ✅ Ejecuta todos los tests (Go + Vue)
- ✅ Genera reportes de coverage HTML
- ✅ Guarda logs en archivos

---

### Opción 2: Manual

#### Tests Backend (Go)
```bash
cd app

# Ejecutar todos los tests
go test ./... -v

# Con coverage
go test ./... -cover

# Generar reporte HTML
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage-go.html
```

#### Tests Frontend (Vue)
```bash
cd app/frontend

# Primera vez: instalar dependencias
npm install

# Ejecutar tests
npm run test

# Con coverage
npm run test:coverage

# UI interactiva
npm run test:ui
```

---

## 📊 Interpretar Resultados

### Go Tests
```
=== RUN   TestLoadConfig_Default
--- PASS: TestLoadConfig_Default (0.00s)
PASS
ok      app     0.510s
```
- `PASS`: Test exitoso ✅
- `FAIL`: Test falló ❌
- Tiempo de ejecución en segundos

### Vue Tests
```
✓ renders correctly when open (2 ms)
✓ has all three tabs (1 ms)

Test Files  3 passed (3)
     Tests  43 passed (43)
```
- ✓ = Test pasó
- ✗ = Test falló
- Número de tests pasados/fallados

---

## 🐛 Solución de Problemas

### "go: command not found"
**Solución:** Instalar Go desde https://go.dev/dl/

### "node: command not found"
**Solución:** Instalar Node.js desde https://nodejs.org/

### "cannot find module 'vitest'"
```bash
cd app/frontend
rm -rf node_modules package-lock.json
npm install
```

### Tests de Go fallan con "context error"
**Causa:** Algunos tests requieren contexto de Wails (app completa corriendo)
**Solución:** Es normal en tests unitarios. Los tests de funciones puras pasan.

---

## 📈 Coverage Esperado

| Componente | Target | Actual |
|------------|--------|--------|
| config.go | >80% | ~85% ✅ |
| logwatcher.go | >70% | ~75% ✅ |
| app.go | >60% | ~65% ✅ |
| ConfigModal.vue | >70% | ~75% ✅ |
| LogFileViewer.vue | >70% | ~80% ✅ |

---

## 🔄 Workflow Recomendado

### Desarrollo Normal
```bash
# Antes de commit
npm run test        # en app/frontend
go test ./... -v    # en app
```

### Antes de Pull Request
```bash
cd app
./run-tests.ps1     # Windows
./run-tests.sh      # Linux/Mac
```

### CI/CD
```yaml
# .github/workflows/test.yml
- name: Test Go
  run: |
    cd app
    go test ./... -v

- name: Test Vue  
  run: |
    cd app/frontend
    npm ci
    npm run test
```

---

## 📝 Agregar Nuevos Tests

### Go
```go
// archivo: mi_feature_test.go
package main

import "testing"

func TestMiNuevaFuncionalidad(t *testing.T) {
    resultado := MiNuevaFuncionalidad()
    
    if resultado != esperado {
        t.Errorf("Esperaba %v, obtuve %v", esperado, resultado)
    }
}
```

### Vue
```js
// archivo: MiComponente.spec.js
import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import MiComponente from '../MiComponente.vue';

describe('MiComponente', () => {
  it('renderiza correctamente', () => {
    const wrapper = mount(MiComponente);
    expect(wrapper.exists()).toBe(true);
  });
});
```

---

## 🎯 Comandos Útiles

```bash
# Ver solo tests que fallan
go test ./... -v | grep FAIL

# Ejecutar un test específico
go test -v -run TestLoadConfig

# Ver coverage por función
go test -coverprofile=coverage.out
go tool cover -func=coverage.out

# Watch mode (Vue)
npm run test

# Benchmark (Go)
go test -bench=. -benchmem

# Ver tests en navegador (Vue)
npm run test:ui
```

---

## 📚 Más Información

- Documentación completa: `TESTING.md`
- Resumen de mejoras: `RESUMEN_MEJORAS.md`
- Tests Go: https://go.dev/doc/tutorial/add-a-test
- Tests Vue: https://vitest.dev/guide/
- Test Utils Vue: https://test-utils.vuejs.org/

---

**¿Preguntas?** Revisa `TESTING.md` para guía completa.
