<?php

namespace KenaitoConf;

class Crypto
{
    private const CRYPT_PREFIX = '{Cipher}';

    public static function decryptConfig(string $secret, string $data): string
    {
        if (!str_starts_with($data, self::CRYPT_PREFIX)) {
            return $data;
        }
        $encrypted = substr($data, strlen(self::CRYPT_PREFIX));
        return self::decryptWithAES($secret, $encrypted);
    }

    public static function decryptWithAES(string $key, string $data): string
    {
        $md5Hash = md5($key);
        $aesKey = substr($md5Hash, 16);
        $cipherText = base64_decode($data);
        $iv = str_repeat("\0", 16);
        $plainText = openssl_decrypt($cipherText, 'AES-128-CBC', $aesKey, OPENSSL_RAW_DATA | OPENSSL_ZERO_PADDING, $iv);
        $padding = ord($plainText[strlen($plainText) - 1]);
        return substr($plainText, 0, -$padding);
    }
}
