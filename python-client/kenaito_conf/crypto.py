import base64
import hashlib

import pyaes

CRYPT_PREFIX = "{Cipher}"


def decrypt_config(secret: str, data: str) -> str:
    if not data.startswith(CRYPT_PREFIX):
        return data
    encrypted = data[len(CRYPT_PREFIX):]
    return decrypt_with_aes(secret, encrypted)


def decrypt_with_aes(key: str, data: str) -> str:
    md5_hash = hashlib.md5(key.encode()).hexdigest()
    aes_key = md5_hash[16:].encode()

    cipher_text = base64.b64decode(data)

    iv = b"\0" * 16
    aes_cbc = pyaes.AESModeOfOperationCBC(aes_key, iv=iv)
    plain_text = aes_cbc.decrypt(cipher_text)

    padding = plain_text[-1]
    plain_text = plain_text[:-padding]

    return plain_text.decode()
