import z from "@zod/zod";
import { ChatRoom } from "../../domain/model/ChatRoom.ts";
import { Member } from "../../domain/model/Member.ts";
import { ChatRoomRepository } from "../../infrastructure/repository/chatRoom.ts";
import { DatabaseType } from "../../infrastructure/repository/database.ts";
import { MemberRepository } from "../../infrastructure/repository/member.ts";

export const createChatRoomCommandSchema = z.object({
  name: z.string(),
  operatorUserId: z.string(),
});

export type CreateChatRoomCommand = z.infer<
  typeof createChatRoomCommandSchema
>;

export class CreateChatRoom {
  constructor(
    private readonly db: DatabaseType,
    private readonly chatRoomRepository: ChatRoomRepository,
    private readonly memberRepository: MemberRepository,
  ) {}

  async execute(command: CreateChatRoomCommand) {
    const chatRoom = ChatRoom.create(command.name);
    const member = Member.create(
      command.operatorUserId,
      chatRoom.id,
    );

    await this.db.transaction().execute(async (trx) => {
      await this.chatRoomRepository.save(chatRoom, trx);
      await this.memberRepository.save(member, trx);
    });
  }
}
