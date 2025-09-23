import { ChatRoomPostingPolicy } from "../../domain/service/ChatRoomPostingPolicy.ts";
import { ChatRoomRepository } from "../../infrastructure/repository/chatRoom.ts";
import { db } from "../../infrastructure/repository/database.ts";
import { MemberRepository } from "../../infrastructure/repository/member.ts";
import { MessageRepository } from "../../infrastructure/repository/message.ts";
import { UserRepository } from "../../infrastructure/repository/user.ts";
import { CreateChatRoom } from "../usecase/createChatRoom.ts";
import { PostMessage } from "../usecase/postMessage.ts";
import { RegisterUser } from "../usecase/registerUser.ts";

// Repositories
export const userRepository = new UserRepository(db);
export const chatRoomRepository = new ChatRoomRepository(db);
export const memberRepository = new MemberRepository(db);
export const messageRepository = new MessageRepository(db);

// Domain Services
export const chatRoomPostingPolicy = new ChatRoomPostingPolicy(
  chatRoomRepository,
  memberRepository,
);

// Use Cases
export const registerUser = new RegisterUser(userRepository);
export const createChatRoom = new CreateChatRoom(
  db,
  chatRoomRepository,
  memberRepository,
);
export const postMessage = new PostMessage(
  db,
  messageRepository,
  chatRoomPostingPolicy,
);
