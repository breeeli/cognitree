export interface Tree {
  id: string;
  title: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Node {
  id: string;
  tree_id: string;
  parent_node_id: string | null;
  anchor_id: string | null;
  question: string;
  status: "draft" | "answered" | "summarized";
  created_at: string;
  updated_at: string;
  qa_pairs?: QAPair[];
}

export interface QAPair {
  id: string;
  question: string;
  blocks: Block[];
  created_at: string;
  status?: "pending" | "complete" | "error";
}

export interface Block {
  id: string;
  type: "paragraph" | "list" | "code" | "quote" | "heading";
  content: string;
}

export interface Anchor {
  id: string;
  block_id: string;
  source_node_id: string;
  start_offset: number;
  end_offset: number;
  quoted_text: string;
  child_node_id: string | null;
}

export interface TreeDetail {
  tree: Tree;
  nodes: Node[];
}

export interface CreateTreeResponse {
  tree: Tree;
  root_node: Node;
}

export interface ThreadResponse {
  nodes: Node[];
}

export interface ChatResponse {
  qa_pair: QAPair;
}

export interface CreateAnchorResponse {
  anchor: Anchor;
  child_node: Node;
}

export interface TreeNode extends Node {
  children: TreeNode[];
}
