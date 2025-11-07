<?php

require 'vendor/autoload.php';

// Crear archivo de configuración si no existe
if (!file_exists('versadumps.yml')) {
    file_put_contents('versadumps.yml', "host: 127.0.0.1\nport: 8080\n");
}

echo "🧪 Testing versaDumps PHP 2.2.0 with VersaDumps Visualizer\n";
echo "══════════════════════════════════════════════════════════\n\n";

sleep(1);

// Test 1: Métodos semánticos
echo "1️⃣ Testing semantic methods (colors)...\n";
vd(['status' => 'Operation completed successfully'])->success();
sleep(1);

vd(['error' => 'Database connection failed', 'code' => 500])->error();
sleep(1);

vd(['message' => 'User logged in', 'user_id' => 123])->info();
sleep(1);

vd(['warning' => 'Low disk space', 'available' => '10%'])->warning();
sleep(1);

vd(['priority' => 'high', 'task' => 'Review security update'])->important();
sleep(1);

// Test 2: Colores personalizados
echo "\n2️⃣ Testing custom colors...\n";
vd(['custom' => 'Purple message'])->color('purple');
sleep(1);

vd(['custom' => 'Pink message'])->color('pink');
sleep(1);

vd(['custom' => 'Cyan message'])->color('cyan');
sleep(1);

// Test 3: Con labels personalizados
echo "\n3️⃣ Testing custom labels...\n";
vd(['user' => 'John', 'age' => 30])->label('User Data')->info();
sleep(1);

vd(['product' => 'Laptop', 'price' => 999])->label('Product Info')->success();
sleep(1);

// Test 4: Auto-detección de variables
echo "\n4️⃣ Testing variable auto-detection...\n";
$usuario = ['nombre' => 'Pedro', 'email' => 'pedro@example.com'];
vd($usuario)->info();
sleep(1);

$pedido = ['id' => 12345, 'total' => 150.50, 'status' => 'pending'];
vd($pedido)->warning();
sleep(1);

// Test 5: Con stack trace
echo "\n5️⃣ Testing with stack trace...\n";
function testFunction()
{
    vd(['debug' => 'Inside test function'])->trace(5)->error();
}
testFunction();
sleep(1);

// Test 6: Estructuras profundas
echo "\n6️⃣ Testing deep structures with depth control...\n";
$deepObject = [
    'level1' => [
        'level2' => [
            'level3' => [
                'level4' => ['deep' => 'value', 'more' => 'data']
            ]
        ]
    ]
];
vd($deepObject)->depth(2)->warning();
sleep(1);

vd($deepObject)->depth(4)->info();
sleep(1);

// Test 7: Datos complejos combinados
echo "\n7️⃣ Testing complex combined scenarios...\n";
$order = [
    'order_id' => 'ORD-2025-001',
    'customer' => [
        'name' => 'Jane Doe',
        'email' => 'jane@example.com',
        'address' => [
            'street' => '123 Main St',
            'city' => 'Springfield',
            'country' => 'USA'
        ]
    ],
    'items' => [
        ['product' => 'Laptop', 'qty' => 1, 'price' => 999],
        ['product' => 'Mouse', 'qty' => 2, 'price' => 25],
    ],
    'total' => 1049,
    'status' => 'processing'
];

vd($order)
    ->label('Order Processing')
    ->success()
    ->depth(3);
sleep(1);

// Test 8: Condicionales
echo "\n8️⃣ Testing conditional execution...\n";
$debug = true;
vd(['message' => 'This only shows if debug is true'])->if($debug)->info();
sleep(1);

$production = false;
vd(['message' => 'This only shows if NOT in production'])->unless($production)->warning();
sleep(1);

// Test 9: Once (simular loop)
echo "\n9️⃣ Testing 'once' in loop...\n";
for ($i = 0; $i < 5; $i++) {
    vd(['iteration' => $i, 'data' => 'Item ' . $i])->once()->info();
}
sleep(1);

// Test 10: Diferentes tipos de datos
echo "\n🔟 Testing different data types...\n";
vd(['string' => 'Hello World'])->info();
vd(['number' => 42])->info();
vd(['float' => 3.14159])->info();
vd(['boolean' => true])->info();
vd(['null' => null])->warning();
vd(['array' => [1, 2, 3, 4, 5]])->success();

echo "\n✅ All tests completed!\n";
echo "Check the VersaDumps Visualizer for results.\n";
