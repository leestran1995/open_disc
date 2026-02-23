import type { JWTClaims } from './types';

/** Decode a JWT payload without verification (backend validates on every request). */
export function decodeJWT(token: string): JWTClaims | null {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload)) as JWTClaims;
  } catch {
    return null;
  }
}
