import type { BaseMessage } from "./base.ts";

export interface MemberCreatedMessage extends BaseMessage {
  database: "chotwork";
  table: "member";
  tableID: number;
  type: "INSERT";
  schemaVersion: number;
  data: {
    id: string;
    userId: string;
    chatRoomId: string;
  };
}
