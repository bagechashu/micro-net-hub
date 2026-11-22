import JSEncrypt from 'jsencrypt';
import { getPublicKey } from '@/api/system/base';

/**
 * 用后端当前的 RSA 公钥加密口令（登录、改密、管理员设置口令时使用）。
 * 公钥在运行时从 /api/base/pubkey 获取，不再在构建时写死（VITE_PUBLIC_KEY），
 * 这样每个部署可以使用自己的密钥，轮换密钥后也不用重新构建前端。
 * 不做缓存：只在提交口令时取一次，避免密钥轮换后页面还拿着旧公钥。
 */
export async function encryptPassword(plain: string): Promise<string> {
  const res: any = await getPublicKey();
  const key = String(res?.data?.publicKey || '');
  if (!key) {
    throw new Error('获取加密公钥失败');
  }
  const encryptor = new JSEncrypt();
  encryptor.setPublicKey(key);
  const encrypted = encryptor.encrypt(plain);
  if (!encrypted) {
    throw new Error('口令加密失败');
  }
  return encrypted;
}
