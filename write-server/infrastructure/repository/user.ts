import { User } from "../../domain/model/User.ts";
import { DatabaseType } from "./database.ts";

export class UserRepository {
  constructor(
    private readonly db: DatabaseType,
  ) {}

  async save(user: User) {
    await this.db.insertInto("user").values({
      id: user.id,
      name: user.name,
    }).execute();
  }

  async findById(id: string): Promise<User | null> {
    const row = await this.db
      .selectFrom("user")
      .selectAll()
      .where(
        "id",
        "=",
        id,
      )
      .executeTakeFirst();

    if (!row) {
      return null;
    }

    return User.__unsafeCreate(row.id, row.name);
  }
}
