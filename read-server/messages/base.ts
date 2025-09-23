export interface BaseMessage {
  version: number;
  type: string;
  commitTs: number;
  buildTs: number;
}
