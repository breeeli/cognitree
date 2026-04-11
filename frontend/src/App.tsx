import { BrowserRouter, Routes, Route } from "react-router-dom";
import { TreePage } from "@/pages/TreePage";

function App() {
  return (
    <BrowserRouter>
      <div className="h-screen flex flex-col">
        <Routes>
          <Route path="/" element={<TreePage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;
