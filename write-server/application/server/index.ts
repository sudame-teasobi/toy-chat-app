import { Hono } from "@hono/hono";
import { validator } from "@hono/hono/validator";
import { createChatRoomCommandSchema } from "../usecase/createChatRoom.ts";
import { postMessageCommandSchema } from "../usecase/postMessage.ts";
import { registerUserCommandSchema } from "../usecase/registerUser.ts";
import { createChatRoom, postMessage, registerUser } from "./deps.ts";

const app = new Hono();

app.all(
  "/health",
  (c) => {
    return c.json({ status: "ok" });
  },
);

app.post(
  "/register-user",
  validator("json", (v, c) => {
    const parsed = registerUserCommandSchema.safeParse(v);
    if (!parsed.success) {
      return c.json({ error: parsed.error }, 400);
    }
    return parsed.data;
  }),
  async (c) => {
    const command = c.req.valid("json");
    await registerUser.execute(command);
    return c.json({ status: "ok" });
  },
);

app.post(
  "/create-chat-room",
  validator("json", (v, c) => {
    const parsed = createChatRoomCommandSchema.safeParse(v);
    if (!parsed.success) {
      return c.json({ error: parsed.error }, 400);
    }
    return parsed.data;
  }),
  async (c) => {
    const command = c.req.valid("json");
    await createChatRoom.execute(command);
    return c.json({ status: "ok" });
  },
);

app.post(
  "/post-message",
  validator("json", (v, c) => {
    const parsed = postMessageCommandSchema.safeParse(v);
    if (!parsed.success) {
      return c.json({ error: parsed.error }, 400);
    }
    return parsed.data;
  }),
  async (c) => {
    const command = c.req.valid("json");
    await postMessage.execute(command);
    return c.json({ status: "ok" });
  },
);

export { app };
