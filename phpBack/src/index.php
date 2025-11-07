<?php

// Incluye el autoload de Composer
require __DIR__ . '/../vendor/autoload.php';

// Crear archivo de configuración si no existe
if (!file_exists(__DIR__ . '/../versadumps.yml')) {
    file_put_contents(__DIR__ . '/../versadumps.yml', "host: 127.0.0.1\nport: 8080\n");
    echo "✅ Archivo versadumps.yml creado\n";
}

echo "\n🚀 VersaDumps PHP 2.2.0 - Test Completo de Características\n";
echo "════════════════════════════════════════════════════════════\n\n";

try {
    // ============================================
    // SECCIÓN 1: MÉTODOS SEMÁNTICOS (COLORES)
    // ============================================
    echo "📦 SECCIÓN 1: Métodos Semánticos\n";
    echo "────────────────────────────────────────\n\n";

    echo "  1.1 Success (Verde) - Operaciones exitosas\n";
    vd(['status' => 'completed', 'message' => 'Payment processed successfully'])->success();
    sleep(1);

    echo "  1.2 Error (Rojo) - Errores críticos\n";
    vd(['error' => 'Database connection failed', 'code' => 500, 'details' => 'Connection timeout'])->error();
    sleep(1);

    echo "  1.3 Info (Azul) - Información general\n";
    vd(['event' => 'user_login', 'user_id' => 12345, 'timestamp' => date('Y-m-d H:i:s')])->info();
    sleep(1);

    echo "  1.4 Warning (Amarillo) - Advertencias\n";
    vd(['warning' => 'Low disk space', 'available' => '10%', 'threshold' => '20%'])->warning();
    sleep(1);

    echo "  1.5 Important (Naranja) - Datos importantes\n";
    vd(['priority' => 'high', 'task' => 'Security update required', 'deadline' => '2025-11-01'])->important();
    sleep(1);

    // ============================================
    // SECCIÓN 2: COLORES PERSONALIZADOS
    // ============================================
    echo "\n🎨 SECCIÓN 2: Colores Personalizados\n";
    echo "────────────────────────────────────────\n\n";

    $colores = ['purple', 'pink', 'cyan', 'gray', 'white', 'red', 'green', 'blue', 'yellow', 'orange'];
    foreach ($colores as $index => $color) {
        echo "  2." . ($index + 1) . " Color: $color\n";
        vd(['color' => $color, 'message' => "Este mensaje tiene color $color"])->color($color);
        sleep(1);
    }

    // ============================================
    // SECCIÓN 3: LABELS PERSONALIZADOS
    // ============================================
    echo "\n🏷️  SECCIÓN 3: Labels Personalizados\n";
    echo "────────────────────────────────────────\n\n";

    echo "  3.1 Label con información de usuario\n";
    vd(['name' => 'Juan Pérez', 'email' => 'juan@example.com', 'role' => 'admin'])
        ->label('👤 Usuario Autenticado')
        ->info();
    sleep(1);

    echo "  3.2 Label con datos de pedido\n";
    vd(['order_id' => 'ORD-2025-001', 'total' => 1500.50, 'items' => 3])
        ->label('🛒 Pedido Procesado')
        ->success();
    sleep(1);

    echo "  3.3 Label con error de validación\n";
    vd(['field' => 'email', 'error' => 'Invalid format', 'value' => 'notanemail'])
        ->label('❌ Error de Validación')
        ->error();
    sleep(1);

    // ============================================
    // SECCIÓN 4: AUTO-DETECCIÓN DE VARIABLES
    // ============================================
    echo "\n🔍 SECCIÓN 4: Auto-detección de Variables\n";
    echo "────────────────────────────────────────\n\n";

    echo "  4.1 Variable simple\n";
    $usuario = ['nombre' => 'María García', 'edad' => 28, 'ciudad' => 'Barcelona'];
    vd($usuario)->info();
    sleep(1);

    echo "  4.2 Variable con estructura compleja\n";
    $configuracion = [
        'app' => ['name' => 'VersaDumps', 'version' => '2.2.0'],
        'database' => ['host' => 'localhost', 'port' => 3306],
        'cache' => ['driver' => 'redis', 'ttl' => 3600]
    ];
    vd($configuracion)->success();
    sleep(1);

    echo "  4.3 Propiedad de objeto\n";
    $api = new stdClass();
    $api->endpoint = 'https://api.example.com/v1';
    $api->credentials = ['key' => 'abc123', 'secret' => '***hidden***'];
    vd($api)->warning();
    sleep(1);

    // ============================================
    // SECCIÓN 5: STACK TRACES
    // ============================================
    echo "\n📚 SECCIÓN 5: Stack Traces\n";
    echo "────────────────────────────────────────\n\n";

    function nivelUno()
    {
        nivelDos();
    }

    function nivelDos()
    {
        nivelTres();
    }

    function nivelTres()
    {
        echo "  5.1 Trace con 3 niveles\n";
        vd(['debug' => 'Punto de debugging', 'level' => 3])
            ->label('🐛 Debug con Trace')
            ->trace(3)
            ->error();
    }

    nivelUno();
    sleep(1);

    echo "  5.2 Trace con 10 niveles\n";
    vd(['debug' => 'Trace completo', 'deep' => true])
        ->label('📍 Stack Trace Completo')
        ->trace(10)
        ->warning();
    sleep(1);

    // ============================================
    // SECCIÓN 6: CONTROL DE PROFUNDIDAD
    // ============================================
    echo "\n🌳 SECCIÓN 6: Control de Profundidad\n";
    echo "────────────────────────────────────────\n\n";

    $deepObject = [
        'level1' => [
            'level2' => [
                'level3' => [
                    'level4' => [
                        'level5' => ['deep' => 'value', 'more' => 'data']
                    ]
                ]
            ]
        ]
    ];

    echo "  6.1 Profundidad limitada a 2 niveles\n";
    vd($deepObject)->depth(2)->warning();
    sleep(1);

    echo "  6.2 Profundidad limitada a 4 niveles\n";
    vd($deepObject)->depth(4)->info();
    sleep(1);

    echo "  6.3 Sin límite de profundidad\n";
    vd($deepObject)->success();
    sleep(1);

    // ============================================
    // SECCIÓN 7: EJECUCIÓN CONDICIONAL
    // ============================================
    echo "\n🔀 SECCIÓN 7: Ejecución Condicional\n";
    echo "────────────────────────────────────────\n\n";

    $debug = true;
    $production = false;

    echo "  7.1 Condicional ->if() (debug=true)\n";
    vd(['message' => 'Este mensaje se muestra porque debug=true'])
        ->if($debug)
        ->info();
    sleep(1);

    echo "  7.2 Condicional ->if() (debug=false)\n";
    vd(['message' => 'Este mensaje NO se muestra'])
        ->if(false)
        ->warning();
    sleep(1);

    echo "  7.3 Condicional ->unless() (production=false)\n";
    vd(['message' => 'Este mensaje se muestra porque NO estamos en producción'])
        ->unless($production)
        ->success();
    sleep(1);

    echo "  7.4 Condicional ->unless() (production=true)\n";
    vd(['message' => 'Este mensaje NO se muestra'])
        ->unless(true)
        ->error();
    sleep(1);

    // ============================================
    // SECCIÓN 8: ONCE (EVITAR DUPLICADOS)
    // ============================================
    echo "\n🔂 SECCIÓN 8: Once (Evitar duplicados en loops)\n";
    echo "────────────────────────────────────────\n\n";

    echo "  8.1 Loop sin ->once() (envía 5 veces)\n";
    for ($i = 0; $i < 5; $i++) {
        vd(['iteration' => $i, 'value' => 'Item ' . $i])->label('Sin Once')->info();
        usleep(200000); // 0.2 segundos
    }

    echo "  8.2 Loop con ->once() (envía solo 1 vez)\n";
    for ($i = 0; $i < 5; $i++) {
        vd(['iteration' => $i, 'value' => 'Item ' . $i])->label('Con Once')->once()->success();
        usleep(200000);
    }
    sleep(1);

    // ============================================
    // SECCIÓN 9: COMBINACIONES AVANZADAS
    // ============================================
    echo "\n⚡ SECCIÓN 9: Combinaciones Avanzadas\n";
    echo "────────────────────────────────────────\n\n";

    echo "  9.1 Builder completo: label + color + trace + depth\n";
    $pedidoComplejo = [
        'order' => [
            'id' => 'ORD-12345',
            'customer' => [
                'name' => 'Ana Rodríguez',
                'email' => 'ana@example.com',
                'address' => [
                    'street' => 'Calle Mayor 123',
                    'city' => 'Madrid',
                    'country' => 'España'
                ]
            ],
            'items' => [
                ['product' => 'Laptop', 'qty' => 1, 'price' => 999.99],
                ['product' => 'Mouse', 'qty' => 2, 'price' => 25.50],
                ['product' => 'Keyboard', 'qty' => 1, 'price' => 75.00]
            ],
            'payment' => [
                'method' => 'credit_card',
                'status' => 'approved',
                'transaction_id' => 'TXN-ABC123'
            ]
        ]
    ];

    vd($pedidoComplejo)
        ->label('💎 Pedido Completo con Todas las Opciones')
        ->color('purple')
        ->trace(3)
        ->depth(3)
        ->if($debug);
    sleep(1);

    echo "  9.2 Combinación: success + label + depth\n";
    vd(['status' => 'ok', 'data' => ['nested' => ['deep' => 'value']]])
        ->label('✅ Operación Exitosa')
        ->success()
        ->depth(2);
    sleep(1);

    echo "  9.3 Combinación: error + trace + condicional\n";
    vd(['error' => 'Critical failure', 'stack' => 'trace included'])
        ->label('🚨 Error Crítico')
        ->error()
        ->trace(5)
        ->if($debug);
    sleep(1);

    // ============================================
    // SECCIÓN 10: DIFERENTES TIPOS DE DATOS
    // ============================================
    echo "\n📊 SECCIÓN 10: Diferentes Tipos de Datos\n";
    echo "────────────────────────────────────────\n\n";

    echo "  10.1 String\n";
    vd(['type' => 'string', 'value' => 'Hello VersaDumps!'])->info();
    sleep(1);

    echo "  10.2 Integer\n";
    vd(['type' => 'integer', 'value' => 42])->success();
    sleep(1);

    echo "  10.3 Float\n";
    vd(['type' => 'float', 'value' => 3.14159])->info();
    sleep(1);

    echo "  10.4 Boolean\n";
    vd(['type' => 'boolean', 'true' => true, 'false' => false])->success();
    sleep(1);

    echo "  10.5 Null\n";
    vd(['type' => 'null', 'value' => null])->warning();
    sleep(1);

    echo "  10.6 Array indexado\n";
    vd(['type' => 'indexed array', 'values' => [1, 2, 3, 4, 5]])->info();
    sleep(1);

    echo "  10.7 Array asociativo\n";
    vd(['type' => 'associative array', 'data' => ['a' => 1, 'b' => 2, 'c' => 3]])->success();
    sleep(1);

    // ============================================
    // SECCIÓN 11: OBJETOS CON toArray()
    // ============================================
    echo "\n🎭 SECCIÓN 11: Objetos con toArray()\n";
    echo "────────────────────────────────────────\n\n";

    class User
    {
        private string $name;
        private string $email;
        private array $roles;
        private bool $active;

        public function __construct(string $name, string $email, array $roles = [], bool $active = true)
        {
            $this->name = $name;
            $this->email = $email;
            $this->roles = $roles;
            $this->active = $active;
        }

        public function toArray(): array
        {
            return [
                'name' => $this->name,
                'email' => $this->email,
                'roles' => $this->roles,
                'active' => $this->active,
                'created_at' => date('Y-m-d H:i:s')
            ];
        }
    }

    echo "  11.1 Usuario admin\n";
    $admin = new User('Carlos Admin', 'carlos@example.com', ['admin', 'editor', 'user']);
    vd($admin)->label('👨‍💼 Usuario Administrador')->success();
    sleep(1);

    echo "  11.2 Usuario regular\n";
    $regular = new User('Laura User', 'laura@example.com', ['user']);
    vd($regular)->label('👤 Usuario Regular')->info();
    sleep(1);

    echo "  11.3 Usuario inactivo\n";
    $inactive = new User('Pedro Inactive', 'pedro@example.com', ['user'], false);
    vd($inactive)->label('🚫 Usuario Inactivo')->warning();
    sleep(1);

    // ============================================
    // SECCIÓN 12: USO TRADICIONAL (BACKWARD COMPATIBLE)
    // ============================================
    echo "\n🔙 SECCIÓN 12: Uso Tradicional (Backward Compatible)\n";
    echo "────────────────────────────────────────\n\n";

    echo "  12.1 Estilo tradicional: vd('label', \$data)\n";
    $tradicional = ['metodo' => 'tradicional', 'compatible' => true];
    vd('datos tradicionales', $tradicional);
    sleep(1);

    echo "  12.2 Estilo tradicional con array\n";
    vd('mi array', [1, 2, 3, 4, 5]);
    sleep(1);

    echo "  12.3 Estilo tradicional con objeto\n";
    $obj = new stdClass();
    $obj->property = 'value';
    vd('mi objeto', $obj);
    sleep(1);

    // ============================================
    // RESUMEN FINAL
    // ============================================
    echo "\n════════════════════════════════════════════════════════════\n";
    echo "✅ ¡Test completo finalizado!\n";
    echo "════════════════════════════════════════════════════════════\n\n";

    echo "📊 Resumen de características probadas:\n";
    echo "  ✓ Métodos semánticos (success, error, info, warning, important)\n";
    echo "  ✓ 10 colores personalizados\n";
    echo "  ✓ Labels personalizados\n";
    echo "  ✓ Auto-detección de variables\n";
    echo "  ✓ Stack traces con diferentes niveles\n";
    echo "  ✓ Control de profundidad\n";
    echo "  ✓ Ejecución condicional (if, unless)\n";
    echo "  ✓ Once (evitar duplicados)\n";
    echo "  ✓ Combinaciones avanzadas\n";
    echo "  ✓ Diferentes tipos de datos\n";
    echo "  ✓ Objetos con toArray()\n";
    echo "  ✓ Backward compatibility\n\n";

    echo "🎯 Revisa VersaDumps Visualizer para ver todos los resultados\n\n";
} catch (Exception $exception) {
    echo "\n❌ Error: " . $exception->getMessage() . "\n";
    echo "Stack trace:\n" . $exception->getTraceAsString() . "\n";
}
