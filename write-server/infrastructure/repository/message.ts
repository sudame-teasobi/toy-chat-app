import { Message } from "../../domain/model/Message.ts";
import { DatabaseType, TransactionType } from "./database.ts";

export class MessageRepository {
  constructor(private db: DatabaseType) {}

  public async save(message: Message, trx?: TransactionType) {
    const db = trx ?? this.db;
    await db.insertInto("message").values({
      id: message.id,
      content: message.content,
      authorMemberId: message.memberId,
      chatRoomId: message.chatRoomId,
      createdAt: message.createdAt,
    }).execute();
  }
}
