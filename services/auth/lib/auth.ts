import { betterAuth } from "better-auth";
import { Pool } from "pg";
import { jwt, role } from "better-auth/plugins";
import { admin } from "better-auth/plugins";
import { organization } from "better-auth/plugins";
import { createClient } from "redis";

const redis = createClient({
  url: process.env.REDIS_URL || "redis://localhost:6379",
});
await redis.connect();

export const auth = betterAuth({
  database: new Pool({
    connectionString:
      process.env.DATABASE_URL || "postgres://postgres@localhost:5432/meetgoti",
  }),
  logger: {
    level: process.env.NODE_ENV === "production" ? "error" : "debug",
    disabled: false,
    verboseLogging: true,
  },
  advanced: {
    generateId: false,
    cookieSameSite: "lax",
    useSecureCookies: process.env.NODE_ENV === "production",
  },
  secondaryStorage: {
    get: async (key) => {
      return await redis.get(key);
    },
    set: async (key, value, ttl) => {
      if (ttl) await redis.set(key, value, { EX: ttl });
      // or for ioredis:
      // if (ttl) await redis.set(key, value, 'EX', ttl)
      else await redis.set(key, value);
    },
    delete: async (key) => {
      await redis.del(key);
    },
  },
  emailAndPassword: {
    enabled: true,
  },
  user: {
    additionalFields: {
      role: {
        type: "string",
        required: true,
        defaultValue: "user",
        return: true,
      },
    },
  },
  session: {
    cookieCache: {
      enabled: true,
      maxAge: 5 * 60, // Cache duration in seconds
    },
  },
  rateLimit: {
    window: 10, // time window in seconds
    max: 100, // max requests in the window
  },
  plugins: [
    jwt({
      jwt: {
        issuer: process.env.BASE_URL || "http://localhost:3001",
        audience: process.env.BASE_URL || "http://localhost:3001",
        expirationTime: "1h", // Token expires in 1 hour
        // Customize JWT payload to include only necessary user data
        definePayload: ({ user }) => {
          const userWithRole = user as any; // Type assertion to access role
          return {
            id: userWithRole.id,
            email: userWithRole.email,
            name: userWithRole.name,
            role: userWithRole.role || "user",
          };
        },
      },
    }),
    organization(),
    admin(),
  ],
});
