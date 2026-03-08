import type { PasswordStrengthResult } from './types';

/** Mirrors Go auth.CheckPasswordStrength() regex rules exactly. */
export function checkPasswordStrength(password: string): PasswordStrengthResult {
  return {
    has_uppercase: /[A-Z]/.test(password),
    has_lowercase: /[a-z]/.test(password),
    has_number: /[0-9]/.test(password),
    has_special: /[!@#$%^&*()\-+]/.test(password),
    has_eight_chars: password.length >= 8,
  };
}

/** Returns true when all 5 password criteria pass. */
export function isPasswordValid(result: PasswordStrengthResult): boolean {
  return (
    result.has_uppercase &&
    result.has_lowercase &&
    result.has_number &&
    result.has_special &&
    result.has_eight_chars
  );
}
