<?php

require_once __DIR__ . '/vendor/autoload.php';

echo "🔍 Test Metadata Debug\n";
echo "═══════════════════════════════════════\n\n";

// Patch temporal para debug - interceptar el payload antes de enviarlo
$originalPost = new ReflectionMethod('Versadumps\Versadumps\VersaDumps', 'post');
$originalPost->setAccessible(true);

// Test 1: Solo color
echo "📍 Test 1: Solo color (warning)\n";
vd(['test' => 'solo color'])->warning();
echo "\n";

// Test 2: Solo trace
echo "📍 Test 2: Solo trace\n";
vd(['test' => 'solo trace'])->trace(3);
echo "\n";

// Test 3: Color + Trace
echo "📍 Test 3: Color + Trace\n";
vd(['test' => 'color y trace'])->trace(3)->warning();
echo "\n";

// Test 4: Trace + Color (orden inverso)
echo "📍 Test 4: Trace + Color (orden inverso)\n";
vd(['test' => 'trace y color'])->warning()->trace(3);
echo "\n";

// Test 5: Con label
echo "📍 Test 5: Con label, trace y color\n";
vd(['test' => 'completo'])->label('🎯 Test Completo')->trace(5)->error();
echo "\n";

echo "═══════════════════════════════════════\n";
echo "✅ Tests enviados. Revisa el servidor.\n";
