import { ChatRoomRepository } from "../../infrastructure/repository/chatRoom.ts";
import { MemberRepository } from "../../infrastructure/repository/member.ts";

export class ChatRoomPostingPolicy {
  constructor(
    private readonly chatRoomRepository: ChatRoomRepository,
    private readonly memberRepository: MemberRepository,
  ) {}

  public async canPost(
    chatRoomId: string,
    userId: string,
  ): Promise<boolean> {
    const member = await this.memberRepository.findByUserIdAndChatRoomId(
      userId,
      chatRoomId,
    );

    return member.isOk();
  }
}
