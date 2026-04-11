import { apiGet, apiPost, apiDelete } from "./client";
import type {
  Node,
  ThreadResponse,
  Anchor,
  CreateAnchorResponse,
} from "@/types/tree";

export function getNode(id: string) {
  return apiGet<Node>(`/nodes/${id}`);
}

export function createChildNode(
  parentId: string,
  question: string,
  anchorId?: string,
) {
  return apiPost<Node>(`/nodes/${parentId}/children`, {
    question,
    anchor_id: anchorId ?? null,
  });
}

export function deleteNode(id: string) {
  return apiDelete(`/nodes/${id}`);
}

export function getThread(nodeId: string) {
  return apiGet<ThreadResponse>(`/nodes/${nodeId}/thread`);
}

export function getAnchors(nodeId: string) {
  return apiGet<Anchor[]>(`/nodes/${nodeId}/anchors`);
}

export function createAnchor(params: {
  block_id: string;
  start_offset: number;
  end_offset: number;
  quoted_text: string;
  child_question: string;
}) {
  return apiPost<CreateAnchorResponse>("/anchors", params);
}
