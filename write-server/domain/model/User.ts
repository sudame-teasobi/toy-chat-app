import { ulid } from "@std/ulid";

export class User {
  private constructor(
    public readonly id: string,
    public readonly name: string,
  ) {}

  public static create(name: string) {
    const id = "user-" + ulid();
    return new User(id, name);
  }

  public static __unsafeCreate(id: string, name: string) {
    return new User(id, name);
  }
}
