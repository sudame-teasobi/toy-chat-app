import type { BaseMessage } from "./base.ts";

export interface ChatRoomCreatedMessage extends BaseMessage {
  database: "chotwork";
  table: "chatRoom";
  tableID: number;
  type: "INSERT";
  schemaVersion: number;
  data: {
    id: string;
    name: string;
  };
}
