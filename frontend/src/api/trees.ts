import { apiDelete, apiGet, apiPost } from "./client";
import { postEventStream } from "./stream";
import type {
  Node,
  QAPair,
  Tree,
  TreeDetail,
  CreateTreeResponse,
} from "@/types/tree";

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

export interface CreateTreeStreamInput {
  question: string;
  tree_id?: string;
  root_node_id?: string;
}

export interface CreateTreeStreamHandlers {
  onTreeReady?: (tree: Tree) => void;
  onRootNodeReady?: (rootNode: Node) => void;
  onDelta?: (delta: string) => void;
  onQAPairReady?: (qaPair: QAPair) => void;
  onCompleted?: () => void;
  onError?: (message: string) => void;
}

interface TreeStreamEvent {
  type: string;
  tree?: Tree;
  root_node?: Node;
  qa_pair?: QAPair;
  delta?: string;
  message?: string;
}

export async function streamCreateTreeFirstQuestion(
  input: CreateTreeStreamInput,
  handlers: CreateTreeStreamHandlers,
) {
  await postEventStream<TreeStreamEvent>("/trees/stream", input, {
    onEvent: (event) => {
      switch (event.type) {
        case "tree_ready":
          if (event.tree) handlers.onTreeReady?.(event.tree);
          break;
        case "root_node_ready":
          if (event.root_node) handlers.onRootNodeReady?.(event.root_node);
          break;
        case "answer_delta":
          if (event.delta) handlers.onDelta?.(event.delta);
          break;
        case "qa_pair_ready":
          if (event.qa_pair) handlers.onQAPairReady?.(event.qa_pair);
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
