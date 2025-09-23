import { ChatRoom } from "../../domain/model/ChatRoom.ts";
import { DatabaseType, TransactionType } from "./database.ts";

export class ChatRoomRepository {
  constructor(private db: DatabaseType) {}

  public findById(id: string, trx?: TransactionType) {
    const db = trx ?? this.db;
    return db
      .selectFrom("chatRoom").where("id", "=", id)
      .selectAll()
      .executeTakeFirst();
  }

  public async save(chatRoom: ChatRoom, trx?: TransactionType) {
    const db = trx ?? this.db;
    await db
      .insertInto("chatRoom").values({
        id: chatRoom.id,
        name: chatRoom.name,
      })
      .execute();
  }
}
