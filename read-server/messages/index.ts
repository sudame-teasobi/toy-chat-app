import type { ChatRoomCreatedMessage } from "./chatRoom.ts";
import type { MemberCreatedMessage } from "./member.ts";
import type { TiCDCSystemMessage } from "./ticdcSystem.ts";

export type Message =
  | TiCDCSystemMessage
  | ChatRoomCreatedMessage
  | MemberCreatedMessage;
