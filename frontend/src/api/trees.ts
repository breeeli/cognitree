import { apiGet, apiPost, apiDelete } from "./client";
import type { Tree, TreeDetail, CreateTreeResponse } from "@/types/tree";

export function listTrees() {
  return apiGet<Tree[]>("/trees");
}

export function getTree(id: string) {
  return apiGet<TreeDetail>(`/trees/${id}`);
}

export function createTree(question: string) {
  return apiPost<CreateTreeResponse>("/trees", { question });
}

export function deleteTree(id: string) {
  return apiDelete(`/trees/${id}`);
}
