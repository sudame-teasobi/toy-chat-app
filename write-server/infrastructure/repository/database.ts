import {
  type Insertable,
  Kysely,
  MysqlDialect,
  type Selectable,
  Transaction,
  type Updateable,
} from "kysely";
import { createPool } from "mysql2";

interface ChatRoomTable {
  id: string;
  name: string;
}

export type ChatRoomRow = Selectable<ChatRoomTable>;
export type NewChatRoomRow = Insertable<ChatRoomTable>;
export type ChatRoomUpdateRow = Updateable<ChatRoomTable>;

interface MemberTable {
  id: string;
  chatRoomId: string;
  userId: string;
}

export type MemberRow = Selectable<MemberTable>;
export type NewMemberRow = Insertable<MemberTable>;
export type MemberUpdateRow = Updateable<MemberTable>;

interface MessageTable {
  id: string;
  chatRoomId: string;
  authorMemberId: string;
  content: string;
  createdAt: Date;
}

export type MessageRow = Selectable<MessageTable>;
export type NewMessageRow = Insertable<MessageTable>;
export type MessageUpdateRow = Updateable<MessageTable>;

interface UserTable {
  id: string;
  name: string;
}

export type Database = {
  chatRoom: ChatRoomTable;
  member: MemberTable;
  message: MessageTable;
  user: UserTable;
};

const dialect = new MysqlDialect({
  pool: createPool({
    host: Deno.env.get("DB_HOST"),
    port: Number(Deno.env.get("DB_PORT")),
    user: Deno.env.get("DB_USER"),
    database: Deno.env.get("DB_NAME"),
  }),
});

export const db = new Kysely<Database>({
  dialect,
});

export type DatabaseType = Kysely<Database>;
export type TransactionType = Transaction<Database>;
