import { ulid } from "@std/ulid";

export class Message {
  private constructor(
    public readonly id: string,
    public readonly content: string,
    public readonly memberId: string,
    public readonly chatRoomId: string,
    public readonly createdAt: Date,
  ) {}

  public static create(
    content: string,
    memberId: string,
    chatRoomId: string,
  ): Message {
    const id = "message-" + ulid();

    return new Message(
      id,
      content,
      memberId,
      chatRoomId,
      new Date(),
    );
  }
}
