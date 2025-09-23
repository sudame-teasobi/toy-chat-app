import z from "@zod/zod";
import { Message } from "../../domain/model/Message.ts";
import { ChatRoomPostingPolicy } from "../../domain/service/ChatRoomPostingPolicy.ts";
import { DatabaseType } from "../../infrastructure/repository/database.ts";
import { MessageRepository } from "../../infrastructure/repository/message.ts";

export const postMessageCommandSchema = z.object({
  content: z.string(),
  memberId: z.string(),
  chatRoomId: z.string(),
});

export type PostMessageCommand = z.infer<typeof postMessageCommandSchema>;

export class PostMessage {
  constructor(
    private readonly db: DatabaseType,
    private readonly messageRepository: MessageRepository,
    private readonly chatRoomPostingPolicy: ChatRoomPostingPolicy,
  ) {}

  async execute(command: PostMessageCommand) {
    const canPost = await this.chatRoomPostingPolicy.canPost(
      command.chatRoomId,
      command.memberId,
    );
    if (!canPost) {
      throw new Error("User cannot post message to this chat room");
    }

    const message = Message.create(
      command.content,
      command.memberId,
      command.chatRoomId,
    );

    await this.db.transaction().execute(async (trx) => {
      await this.messageRepository.save(message, trx);
    });
  }
}
