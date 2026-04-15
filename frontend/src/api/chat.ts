import { apiPost } from "./client";
import { postEventStream } from "./stream";
import type { ChatResponse } from "@/types/tree";

export function chatOnNode(nodeId: string, question: string) {
  return apiPost<ChatResponse>(`/nodes/${nodeId}/chat`, { question });
}

export interface ChatStreamInput {
  question: string;
}

export interface ChatStreamHandlers {
  onDelta?: (delta: string) => void;
  onCompleted?: () => void;
  onError?: (message: string) => void;
}

interface ChatStreamEvent {
  type: string;
  delta?: string;
  message?: string;
}

export async function streamChatOnNode(
  nodeId: string,
  input: ChatStreamInput,
  handlers: ChatStreamHandlers,
) {
  await postEventStream<ChatStreamEvent>(`/nodes/${nodeId}/chat/stream`, input, {
    onEvent: (event) => {
      switch (event.type) {
        case "answer_delta":
          if (event.delta) handlers.onDelta?.(event.delta);
          break;
        case "error":
          handlers.onError?.(event.message ?? "stream error");
          throw new Error(event.message ?? "stream error");
        case "completed":
          handlers.onCompleted?.();
          break;
        default:
          break;
      }
    },
  });
}
