import { Elysia } from "elysia";
import { cors } from "@elysiajs/cors";
import { Pool } from "pg";

import { auth } from "../lib/auth";

// Get database pool
const db = new Pool({
  connectionString:
    process.env.DATABASE_URL || "postgres://postgres@localhost:5432/meetgoti",
});

// Custom logger for Better Auth requests
const logger = new Elysia()
  .onRequest(({ request }) => {
    const timestamp = new Date().toISOString();
    const url = new URL(request.url);
    console.log(
      `\x1b[36m[${timestamp}]\x1b[0m \x1b[33m${request.method}\x1b[0m ${url.pathname}`
    );
  })
  .onError(({ error, request }) => {
    const timestamp = new Date().toISOString();
    const url = new URL(request.url);
    console.error(
      `\x1b[36m[${timestamp}]\x1b[0m \x1b[31m[ERROR]\x1b[0m ${request.method} ${url.pathname}`
    );
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`\x1b[31m${errorMessage}\x1b[0m`);
    if (error instanceof Error && error.stack) {
      console.error(`\x1b[90m${error.stack}\x1b[0m`);
    }
  })
  .onAfterHandle(({ request, set }) => {
    const timestamp = new Date().toISOString();
    const url = new URL(request.url);
    const status = typeof set.status === "number" ? set.status : 200;
    const statusColor = status >= 400 ? "\x1b[31m" : "\x1b[32m";
    console.log(
      `\x1b[36m[${timestamp}]\x1b[0m ${statusColor}${status}\x1b[0m ${request.method} ${url.pathname}`
    );
  });

const app = new Elysia()
  .use(logger)
  .use(
    cors({
      origin: ["http://localhost:3000", "http://localhost:8080"],
      methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
      credentials: true,
      allowedHeaders: ["Content-Type", "Authorization"],
    })
  )
  .mount(auth.handler)
  // Custom endpoint to get user with proper role information
  .get("/api/auth/me", async ({ request, set }) => {
    try {
      const session = await auth.api.getSession({
        headers: request.headers,
      });

      if (!session) {
        set.status = 401;
        return { error: "Unauthorized" };
      }

      // Return user with role from database
      const user = session.user as any; // Type assertion to access role
      return {
        user: {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role || "user",
          image: user.image,
        },
        session: {
          id: session.session.id,
          userId: session.session.userId,
          expiresAt: session.session.expiresAt,
        },
      };
    } catch (error) {
      console.error("Error getting user session:", error);
      set.status = 500;
      return { error: "Internal server error" };
    }
  })
  // Get all staff members (Admin only)
  .get("/api/auth/staff", async ({ request, set }) => {
    try {
      const session = await auth.api.getSession({
        headers: request.headers,
      });

      if (!session) {
        set.status = 401;
        return { error: "Unauthorized" };
      }

      const user = session.user as any;
      if (user.role !== "admin" && user.role !== "ADMIN") {
        set.status = 403;
        return { error: "Forbidden: Admin access required" };
      }

      // Query all users with staff or admin roles
      const result = await db.query(
        `SELECT id, name, email, role, image, "createdAt", "updatedAt" 
         FROM "user" 
         WHERE role IN ('admin', 'staff', 'ADMIN', 'STAFF', 'manager', 'MANAGER')
         ORDER BY "createdAt" DESC`
      );

      return {
        success: true,
        data: result.rows.map((row: any) => ({
          id: row.id,
          name: row.name,
          email: row.email,
          role: row.role,
          image: row.image,
          status: "Active", // You can add a status field to the database if needed
          position:
            row.role === "admin" || row.role === "ADMIN"
              ? "Administrator"
              : row.role === "manager" || row.role === "MANAGER"
              ? "Manager"
              : "Staff",
          createdAt: row.createdAt,
          updatedAt: row.updatedAt,
        })),
      };
    } catch (error) {
      console.error("Error fetching staff:", error);
      set.status = 500;
      return { error: "Internal server error" };
    }
  })
  // Get staff member by ID (Admin only)
  .get("/api/auth/staff/:id", async ({ request, set, params }) => {
    try {
      const session = await auth.api.getSession({
        headers: request.headers,
      });

      if (!session) {
        set.status = 401;
        return { error: "Unauthorized" };
      }

      const user = session.user as any;
      if (user.role !== "admin" && user.role !== "ADMIN") {
        set.status = 403;
        return { error: "Forbidden: Admin access required" };
      }

      const result = await db.query(
        `SELECT id, name, email, role, image, "createdAt", "updatedAt" 
         FROM "user" 
         WHERE id = $1 AND role IN ('admin', 'staff', 'ADMIN', 'STAFF', 'manager', 'MANAGER')`,
        [params.id]
      );

      if (result.rows.length === 0) {
        set.status = 404;
        return { error: "Staff member not found" };
      }

      const staff = result.rows[0];
      return {
        success: true,
        data: {
          id: staff.id,
          name: staff.name,
          email: staff.email,
          role: staff.role,
          image: staff.image,
          status: "Active",
          position:
            staff.role === "admin" || staff.role === "ADMIN"
              ? "Administrator"
              : staff.role === "manager" || staff.role === "MANAGER"
              ? "Manager"
              : "Staff",
          createdAt: staff.createdAt,
          updatedAt: staff.updatedAt,
        },
      };
    } catch (error) {
      console.error("Error fetching staff member:", error);
      set.status = 500;
      return { error: "Internal server error" };
    }
  })
  // Update staff member (Admin only)
  .put("/api/auth/staff/:id", async ({ request, set, params, body }) => {
    try {
      const session = await auth.api.getSession({
        headers: request.headers,
      });

      if (!session) {
        set.status = 401;
        return { error: "Unauthorized" };
      }

      const user = session.user as any;
      if (user.role !== "admin" && user.role !== "ADMIN") {
        set.status = 403;
        return { error: "Forbidden: Admin access required" };
      }

      const { name, email, role } = body as any;

      const result = await db.query(
        `UPDATE "user" 
         SET name = COALESCE($1, name), 
             email = COALESCE($2, email), 
             role = COALESCE($3, role),
             "updatedAt" = NOW()
         WHERE id = $4
         RETURNING id, name, email, role, image, "createdAt", "updatedAt"`,
        [name, email, role, params.id]
      );

      if (result.rows.length === 0) {
        set.status = 404;
        return { error: "Staff member not found" };
      }

      const staff = result.rows[0];
      return {
        success: true,
        message: "Staff member updated successfully",
        data: {
          id: staff.id,
          name: staff.name,
          email: staff.email,
          role: staff.role,
          image: staff.image,
          status: "Active",
          position:
            staff.role === "admin" || staff.role === "ADMIN"
              ? "Administrator"
              : staff.role === "manager" || staff.role === "MANAGER"
              ? "Manager"
              : "Staff",
          createdAt: staff.createdAt,
          updatedAt: staff.updatedAt,
        },
      };
    } catch (error) {
      console.error("Error updating staff member:", error);
      set.status = 500;
      return { error: "Internal server error" };
    }
  })
  // Delete staff member (Admin only)
  .delete("/api/auth/staff/:id", async ({ request, set, params }) => {
    try {
      const session = await auth.api.getSession({
        headers: request.headers,
      });

      if (!session) {
        set.status = 401;
        return { error: "Unauthorized" };
      }

      const user = session.user as any;
      if (user.role !== "admin" && user.role !== "ADMIN") {
        set.status = 403;
        return { error: "Forbidden: Admin access required" };
      }

      // Prevent admin from deleting themselves
      if (user.id === params.id) {
        set.status = 400;
        return { error: "Cannot delete your own account" };
      }

      const result = await db.query(
        `DELETE FROM "user" WHERE id = $1 RETURNING id`,
        [params.id]
      );

      if (result.rows.length === 0) {
        set.status = 404;
        return { error: "Staff member not found" };
      }

      return {
        success: true,
        message: "Staff member deleted successfully",
      };
    } catch (error) {
      console.error("Error deleting staff member:", error);
      set.status = 500;
      return { error: "Internal server error" };
    }
  })
  .macro({
    auth: {
      async resolve({ status, request: { headers } }) {
        const session = await auth.api.getSession({
          headers,
        });
        if (!session) return status(401);
        return {
          user: session.user,
          session: session.session,
        };
      },
    },
  })
  .listen(process.env.PORT || 3001);

console.log(
  `🦊 Elysia is running at ${app.server?.hostname}:${app.server?.port}`
);
