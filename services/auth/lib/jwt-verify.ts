import { jwtVerify, createRemoteJWKSet, type JWTPayload } from "jose";

const AUTH_SERVICE_URL =
  process.env.AUTH_SERVICE_URL || "http://localhost:3001";

// Create a remote JWKS that fetches the public keys from the auth server
const JWKS = createRemoteJWKSet(new URL(`${AUTH_SERVICE_URL}/jwks`));

export interface AuthPayload extends JWTPayload {
  id: string;
  email: string;
  name?: string;
  role?: string;
}

/**
 * Verify a JWT token using the auth server's JWKS endpoint
 * @param token - The JWT token to verify
 * @returns The decoded payload if valid
 * @throws Error if token is invalid or expired
 */
export async function verifyToken(token: string): Promise<AuthPayload> {
  try {
    const { payload } = await jwtVerify(token, JWKS, {
      issuer: AUTH_SERVICE_URL,
      audience: AUTH_SERVICE_URL,
    });

    return payload as AuthPayload;
  } catch (error) {
    console.error("Token verification failed:", error);
    throw new Error("Invalid or expired token");
  }
}

/**
 * Extract the Bearer token from the Authorization header
 * @param authHeader - The Authorization header value
 * @returns The token string without the "Bearer " prefix
 */
export function extractBearerToken(authHeader: string | null): string | null {
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    return null;
  }
  return authHeader.substring(7);
}

/**
 * Middleware-style function to verify token from request headers
 * @param headers - Request headers object
 * @returns The verified user payload
 * @throws Error if token is missing or invalid
 */
export async function verifyRequestToken(
  headers: Record<string, string | undefined>
): Promise<AuthPayload> {
  const authHeader = headers.authorization || headers.Authorization;

  if (!authHeader) {
    throw new Error("Authorization header missing");
  }

  const token = extractBearerToken(authHeader);

  if (!token) {
    throw new Error("Invalid authorization format. Expected: Bearer <token>");
  }

  return await verifyToken(token);
}
