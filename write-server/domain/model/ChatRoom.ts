import { ulid } from "@std/ulid";

export class ChatRoom {
  private constructor(
    public readonly id: string,
    public readonly name: string,
  ) {}

  public static create(name: string) {
    // TODO: 名前の文字数制限などのバリデーション
    const id = "chatRoom-" + ulid();
    return new ChatRoom(id, name);
  }

  public static __unsafeCreate(id: string, name: string) {
    return new ChatRoom(id, name);
  }
}
