import { apiPost } from "./client";
import type { ChatResponse } from "@/types/tree";

export function chatOnNode(nodeId: string, question: string) {
  return apiPost<ChatResponse>(`/nodes/${nodeId}/chat`, { question });
}
