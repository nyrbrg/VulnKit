<?php
$search = $_GET['q'] ?? '';
?>
<!DOCTYPE html>
<html>
<head><title>VulnKit - Reflected XSS</title></head>
<body>
  <h2>Search</h2>
  <form method="GET">
    <input name="q" value="<?= $search ?>">
    <button>Search</button>
  </form>
  <?php if ($search): ?>
    <p>Results for: <?= $search ?></p>
  <?php endif; ?>
</body>
</html>