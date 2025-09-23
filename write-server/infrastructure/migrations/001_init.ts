import type { Kysely } from "kysely";

export async function up(db: Kysely<unknown>): Promise<void> {
  await db.schema
    .createTable("user")
    .addColumn("id", "varchar(63)", (col) => col.primaryKey().unique())
    .addColumn("name", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("chatRoom")
    .addColumn("id", "varchar(63)", (col) => col.primaryKey().unique())
    .addColumn("name", "text", (col) => col.notNull())
    .execute();

  await db.schema
    .createTable("member")
    .addColumn("id", "varchar(63)", (col) => col.primaryKey().unique())
    .addColumn(
      "chatRoomId",
      "varchar(63)",
      (col) => col.notNull().references("chatRoom.id"),
    )
    .addColumn(
      "userId",
      "varchar(63)",
      (col) => col.notNull().references("user.id"),
    )
    .execute();

  await db.schema
    .createTable("message")
    .addColumn("id", "varchar(63)", (col) => col.primaryKey().unique())
    .addColumn(
      "chatRoomId",
      "varchar(63)",
      (col) => col.notNull().references("chatRoom.id"),
    )
    .addColumn(
      "authorUserId",
      "varchar(63)",
      (col) => col.notNull().references("user.id"),
    )
    .addColumn("content", "text", (col) => col.notNull())
    .addColumn("createdAt", "datetime", (col) => col.notNull())
    .execute();
}

export async function down(db: Kysely<unknown>): Promise<void> {
  await db.schema.dropTable("message").ifExists().execute();
  await db.schema.dropTable("member").ifExists().execute();
  await db.schema.dropTable("chatRoom").ifExists().execute();
  await db.schema.dropTable("user").ifExists().execute();
}
