import { BASE_URL } from "./client";

export interface StreamCallbacks<TEvent> {
  onEvent: (event: TEvent) => void;
}

export async function postEventStream<TEvent extends { type: string }>(
  path: string,
  body: unknown,
  callbacks: StreamCallbacks<TEvent>,
) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }

  if (!res.body) {
    throw new Error("stream body unavailable");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let eventName = "";
  let dataLines: string[] = [];

  const flushEvent = () => {
    if (!eventName || dataLines.length === 0) return;

    const raw = dataLines.join("\n");
    const event = JSON.parse(raw) as TEvent;
    callbacks.onEvent(event);
  };

  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      let newlineIndex = buffer.indexOf("\n");
      while (newlineIndex >= 0) {
        const line = buffer.slice(0, newlineIndex).replace(/\r$/, "");
        buffer = buffer.slice(newlineIndex + 1);

        if (line === "") {
          flushEvent();
          eventName = "";
          dataLines = [];
        } else if (line.startsWith("event:")) {
          eventName = line.slice("event:".length).trim();
        } else if (line.startsWith("data:")) {
          dataLines.push(line.slice("data:".length).trimStart());
        }

        newlineIndex = buffer.indexOf("\n");
      }
    }

    if (buffer.trim() !== "") {
      const line = buffer.replace(/\r$/, "");
      if (line.startsWith("event:")) {
        eventName = line.slice("event:".length).trim();
      } else if (line.startsWith("data:")) {
        dataLines.push(line.slice("data:".length).trimStart());
      }
    }

    flushEvent();
  } finally {
    reader.releaseLock();
  }
}
