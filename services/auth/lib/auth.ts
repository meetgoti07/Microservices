import { betterAuth } from "better-auth";
import { Pool } from "pg";
import { jwt } from "better-auth/plugins";
import { admin } from "better-auth/plugins";
import { organization } from "better-auth/plugins";
import { createClient } from "redis";

const redis = createClient();
await redis.connect();

export const auth = betterAuth({
  database: new Pool({
    connectionString: "postgres://meetgoti@localhost:5432/mapauth",
  }),
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
  plugins: [jwt(), organization(), admin()],
});
