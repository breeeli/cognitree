import { useState, useEffect, useCallback, useRef } from "react";

interface TextSelection {
  text: string;
  rect: DOMRect;
  blockId: string;
  startOffset: number;
  endOffset: number;
}

export function useTextSelection(containerSelector: string) {
  const [selection, setSelection] = useState<TextSelection | null>(null);
  const lockedRef = useRef(false);

  const lockSelection = useCallback(() => {
    lockedRef.current = true;
  }, []);

  const handleMouseUp = useCallback(() => {
    if (lockedRef.current) return;

    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || !sel.toString().trim()) {
      return;
    }

    const text = sel.toString().trim();
    if (text.length < 2) return;

    const range = sel.getRangeAt(0);
    const container = range.commonAncestorContainer;

    const el = container instanceof Element ? container : container.parentElement;
    if (!el) return;

    const answerEl = el.closest(containerSelector);
    if (!answerEl) return;

    const blockEl = el.closest("[data-block-id]");
    const blockId = blockEl?.getAttribute("data-block-id");
    if (!blockId) return;

    const rect = range.getBoundingClientRect();

    const beforeRange = document.createRange();
    beforeRange.setStart(blockEl!, 0);
    beforeRange.setEnd(range.startContainer, range.startOffset);
    const startOffset = beforeRange.toString().length;
    const endOffset = startOffset + text.length;

    setSelection({ text, rect, blockId, startOffset, endOffset });
  }, [containerSelector]);

  const clearSelection = useCallback(() => {
    lockedRef.current = false;
    setSelection(null);
    window.getSelection()?.removeAllRanges();
  }, []);

  useEffect(() => {
    document.addEventListener("mouseup", handleMouseUp);
    return () => document.removeEventListener("mouseup", handleMouseUp);
  }, [handleMouseUp]);

  return { selection, clearSelection, lockSelection };
}
