import { useNavigate } from "react-router-dom";
import { CreateTreeWorkspace } from "@/components/workspace/CreateTreeWorkspace";

export function NewTreePage() {
  const navigate = useNavigate();

  return (
    <CreateTreeWorkspace
      onCreated={({ treeId }) => {
        localStorage.setItem("cognitree:currentTreeId", treeId);
        navigate("/", { replace: true });
      }}
    />
  );
}
