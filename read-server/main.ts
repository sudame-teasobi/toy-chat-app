import { ulid } from "@std/ulid";
import { Kafka } from "kafkajs";
import type { Message } from "./messages/index.ts";

async function main() {
  const kafka = new Kafka({
    clientId: "my-app",
    brokers: [Deno.env.get("KAFKA_BROKER")].filter((b) => b != null),
  });

  const consumer = kafka.consumer({ groupId: `test-group-${ulid()}` });

  await consumer.connect();
  await consumer.subscribe({ topic: "ticdc", fromBeginning: true });

  await consumer.run({
    eachMessage: ({ message }): Promise<void> => {
      const stringValue = message.value?.toString("utf-8");
      if (stringValue == null) {
        return Promise.resolve();
      }

      let value: Message;
      try {
        value = JSON.parse(stringValue);
      } catch {
        console.warn("Failed to parse message value as JSON:", stringValue);
        return Promise.resolve();
      }

      if (value.type === "WATERMARK" || value.type === "BOOTSTRAP") {
        return Promise.resolve();
      }

      console.log({
        value: message.value?.toString("utf-8"),
      });
      return Promise.resolve();
    },
  });
}

void main();
