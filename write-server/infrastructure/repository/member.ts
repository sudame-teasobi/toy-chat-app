import { err, ok, Result } from "neverthrow";
import { Member } from "../../domain/model/Member.ts";
import { DatabaseType, TransactionType } from "./database.ts";

export class MemberRepository {
  constructor(private db: DatabaseType) {}

  public async findById(
    id: string,
    trx?: TransactionType,
  ): Promise<Result<Member, null>> {
    const db = trx ?? this.db;
    const res = await db.selectFrom("member")
      .where("id", "=", id)
      .selectAll()
      .executeTakeFirst();

    if (res == null) {
      return err(null);
    }

    return ok(Member.__unsafeCreate(res.id, res.userId, res.chatRoomId));
  }

  public async findByUserIdAndChatRoomId(
    userId: string,
    chatRoomId: string,
    trx?: TransactionType,
  ): Promise<Result<Member, null>> {
    const db = trx ?? this.db;
    const res = await db.selectFrom("member")
      .where("userId", "=", userId)
      .where("chatRoomId", "=", chatRoomId)
      .selectAll()
      .executeTakeFirst();

    if (res == null) {
      return err(null);
    }

    return ok(Member.__unsafeCreate(res.id, res.userId, res.chatRoomId));
  }

  public async findByChatRoomId(
    chatRoomId: string,
    trx?: TransactionType,
  ): Promise<Result<Member[], null>> {
    const db = trx ?? this.db;
    const res = await db
      .selectFrom("member").where(
        "chatRoomId",
        "=",
        chatRoomId,
      )
      .selectAll()
      .execute();

    if (res == null) {
      return err(null);
    }

    return ok(
      res.map((r) => Member.__unsafeCreate(r.id, r.userId, r.chatRoomId)),
    );
  }

  public async save(member: Member, trx?: TransactionType) {
    const db = trx ?? this.db;
    await db.insertInto("member").values({
      id: member.id,
      userId: member.userId,
      chatRoomId: member.chatRoomId,
    }).execute();
  }
}
