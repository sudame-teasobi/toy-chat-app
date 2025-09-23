export interface BootstrapMessage {
  version: number;
  type: 'BOOTSTRAP';
  commitTs: number;
  buildTs: number;
}

export interface WatermarkMessage {
  version: number;
  type: 'WATERMARK';
  commitTs: number;
  buildTs: number;
}

export type TiCDCSystemMessage = WatermarkMessage | BootstrapMessage;
