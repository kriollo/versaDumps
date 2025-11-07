<?php

require __DIR__ . '/vendor/autoload.php';

// Crear archivo de configuración
if (!file_exists(__DIR__ . '/versadumps.yml')) {
    file_put_contents(__DIR__ . '/versadumps.yml', "host: 127.0.0.1\nport: 8080\n");
}

echo "🔍 Debug: Verificando payload con trace\n";
echo "════════════════════════════════════════\n\n";

// Test 1: Trace básico
echo "Test 1: ->trace(3)\n";
vd(['test' => 'trace básico'])->label('Test Trace 3')->trace(3)->error();
sleep(2);

// Test 2: Trace con color
echo "\nTest 2: ->trace(5)->color('purple')\n";
vd(['test' => 'trace con color'])->label('Test Trace Color')->trace(5)->color('purple');
sleep(2);

// Test 3: Trace con método semántico
echo "\nTest 3: ->trace(3)->error()\n";
vd(['test' => 'trace con error'])->label('Test Trace Error')->trace(3)->error();
sleep(2);

// Test 4: Trace en función anidada
function nivel1()
{
    nivel2();
}

function nivel2()
{
    nivel3();
}

function nivel3()
{
    echo "\nTest 4: Trace desde función anidada\n";
    vd(['test' => 'trace anidado', 'nivel' => 3])->label('Test Trace Anidado')->trace(10)->warning();
}

nivel1();
sleep(2);

// Test 5: Combinación completa
echo "\nTest 5: Combinación completa\n";
vd([
    'data' => 'complejo',
    'nested' => ['level' => 1]
])
    ->label('Test Completo')
    ->trace(5)
    ->color('cyan')
    ->depth(3);

echo "\n✅ Tests enviados. Revisa el payload en el servidor.\n";
