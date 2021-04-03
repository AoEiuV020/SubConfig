<?php
$debug = true;
if ($_POST['token'] != trim(file_get_contents('upload_token'))) {
    if ($debug) {
        echo 'token wrong: ' . $_POST['token'];
    }
    http_response_code(400);
    return;
}
if (!array_key_exists('file', $_FILES)) {
    if ($debug) {
        echo 'file key wrong: ';
        print_r(array_keys($FILES));
    }
    http_response_code(400);
    return;
}
$file = $_FILES['file'];
if ($file['name'] != 'config.tar.gz.aes') {
    if ($debug) {
        echo 'file name wrong: ' . $file['name'];
    }
    http_response_code(400);
    return;
}
$output = false;
$encrypt_method = "AES-256-CBC";
$key = base64_decode(file_get_contents('upload_secret'));
$iv = base64_decode('EJwC9OfO/fkuTvPax7YHeQ==');
$raw = file_get_contents($file['tmp_name']);
$output = openssl_decrypt($raw, $encrypt_method, $key, OPENSSL_RAW_DATA, $iv);
if (!$output) {
    if ($debug) {
        echo 'decrypt failed: ' . strlen($raw);
    }
    http_response_code(400);
    return;
}
file_put_contents('config.tar.gz', $output);
try {
    $p = new PharData('config.tar.gz');
    unlink('config.tar');
    $p->decompress();
} catch (Exception $e) {
    if ($debug) {
        echo 'decompress failed: ' . $e->getMessage();
    }
    http_response_code(400);
    return;
} finally {
    unlink('config.tar.gz');
}
try {
    $phar = new PharData('config.tar');
    $phar->extractTo('config/', null, true);
} catch (Exception $e) {
    if ($debug) {
        echo 'unpack failed: ' . $e->getMessage();
    }
    http_response_code(400);
    return;
} finally {
    unlink('config.tar');
}
echo 'deploy success';
