import crypto from 'node:crypto';

const CRYPT_PREFIX = '{Cipher}';

export function decryptConfig(secret, data) {
  if (!data.startsWith(CRYPT_PREFIX)) {
    return data;
  }
  const encrypted = data.slice(CRYPT_PREFIX.length);
  return decryptWithAES(secret, encrypted);
}

export function decryptWithAES(key, data) {
  const md5Hash = crypto.createHash('md5').update(key).digest('hex');
  const aesKey = md5Hash.slice(16);
  const cipherText = Buffer.from(data, 'base64');
  const iv = Buffer.alloc(16, 0);
  const decipher = crypto.createDecipheriv('aes-128-cbc', aesKey, iv);
  decipher.setAutoPadding(false);
  let plainText = Buffer.concat([decipher.update(cipherText), decipher.final()]);
  const padding = plainText[plainText.length - 1];
  plainText = plainText.slice(0, plainText.length - padding);
  return plainText.toString('utf8');
}
