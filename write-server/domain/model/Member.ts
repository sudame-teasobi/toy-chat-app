import { ulid } from "@std/ulid";

export class Member {
  private constructor(
    public readonly id: string,
    public readonly userId: string,
    public readonly chatRoomId: string,
  ) {}

  public static create(userId: string, chatRoomId: string) {
    const id = "member-" + ulid();
    return new Member(id, userId, chatRoomId);
  }

  public static __unsafeCreate(id: string, userId: string, chatRoomId: string) {
    return new Member(id, userId, chatRoomId);
  }
}
