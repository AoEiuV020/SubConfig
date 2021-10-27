<?php
# https://host/sub?url=<secret>
# https://host/sub?target=quan&url=<secret>-basic
$sub = "https://subc.020.name/sub?";
#$sub = "https://sub.prpr.xyz/sub?";
#$sub = "https://sub.maoxiongnet.com/sub?";
#$sub = "https://pub-api-1.bianyuan.xyz/sub?";
#$sub = "https://sub.id9.cc/sub?";
$params = $_SERVER['QUERY_STRING'];
$config = "https://github.com/AoEiuV020/SubConfig/raw/main/subconverter.ini";
$target = 'clash';
$example_params = "emoji=true&list=false&udp=true&tfo=false&scv=false&fdn=false&sort=false&new_name=true";
$key = trim(file_get_contents('sub_secret'));
parse_str($params, $get_array);
$url = $get_array['url'];
# key不对的直接301转给其他后端，
if (!(strlen($url) >= strlen($key) && substr($url, 0, strlen($key)) == $key)) {
    header("Status: 301 Moved Permanently");
    header("Location: " . $sub . $params);
    return;
}
# key对了就注入订阅和杂项参数，
parse_str($example_params, $example_array);
if (array_key_exists('target', $get_array)) {
    $target = $get_array['target'];
}
$example_array['target'] = $target;
if (array_key_exists('ver', $get_array)) {
    $target = $target . '&ver=' . $get_array['ver'];
    $example_array['ver'] = $ver; 
}
# 判断是否存在缓存，
$suffix = substr($url, strlen($key), strlen($url));
$cache_file = 'config/' . $target . $suffix;
if (!(array_key_exists('cache', $get_array) && $get_array['cache'] !== 'true')
    && file_exists($cache_file)) {
    # 没有指定不使用缓存，并且缓存文件存在，直接返回缓存，
    echo file_get_contents($cache_file);
    return;
}
# 注入订阅构造请求其他后端的最终url,
$lines = array();
if ($file = fopen("subscribe", "r")) {
    while (!feof($file)) {
        $line = trim(fgets($file));
        if (strlen($line) > 0) {
            $lines[] = $line;
        }
    }
    fclose($file);
}
$urls = join("|", $lines);
$example_array['url'] = join("|", $lines);
if (strlen($suffix) > 0 && substr($config, strlen($config) - 4, strlen($config)) == '.ini') {
    $config = substr($config, 0, strlen($config) - 4) . $suffix . '.ini';
}
$example_array['config'] = $config;
$params = http_build_query($example_array);
$result = file_get_contents($sub . $params);
# 请求完成把配置保存到缓存文件，
file_put_contents($cache_file, $result);
echo $result;
