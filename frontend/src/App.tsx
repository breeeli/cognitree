import { BrowserRouter, Routes, Route } from "react-router-dom";
import { TreePage } from "@/pages/TreePage";
import { NewTreePage } from "@/pages/NewTreePage";

function App() {
  return (
    <BrowserRouter>
      <div className="h-screen flex flex-col">
        <div className="flex-1 min-h-0">
          <Routes>
            <Route path="/" element={<TreePage />} />
            <Route path="/trees/new" element={<NewTreePage />} />
          </Routes>
        </div>
      </div>
    </BrowserRouter>
  );
}

export default App;
