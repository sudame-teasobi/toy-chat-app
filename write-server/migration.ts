import { FileMigrationProvider, Migrator } from "kysely";
import fs from "node:fs/promises";
import path from "node:path";
import { db } from "./infrastructure/repository/database.ts";

async function migrateToLatest() {
  const migrator = new Migrator({
    db,
    provider: new FileMigrationProvider({
      fs,
      path,
      migrationFolder: path.resolve(
        import.meta.dirname ?? Deno.cwd(),
        "./infrastructure/migrations/",
      ),
    }),
  });

  const { error, results } = await migrator.migrateToLatest();

  results?.forEach((it) => {
    if (it.status === "Success") {
      console.log(`migration "${it.migrationName}" was executed successfully`);
    } else if (it.status === "Error") {
      console.error(`failed to execute migration "${it.migrationName}"`);
    }
  });

  if (error) {
    console.error("failed to migrate");
    console.error(error);
  }

  await db.destroy();
}

migrateToLatest();
