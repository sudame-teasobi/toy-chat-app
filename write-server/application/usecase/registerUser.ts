import z from "@zod/zod";
import { User } from "../../domain/model/User.ts";
import { UserRepository } from "../../infrastructure/repository/user.ts";

export const registerUserCommandSchema = z.object({
  name: z.string(),
});

export type RegisterUserCommand = z.infer<typeof registerUserCommandSchema>;

export class RegisterUser {
  constructor(
    private readonly userRepository: UserRepository,
  ) {}

  async execute(command: RegisterUserCommand) {
    const user = User.create(command.name);
    await this.userRepository.save(user);
  }
}
