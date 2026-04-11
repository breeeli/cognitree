import { NavLink } from "react-router-dom";

const navItems = [{ to: "/", label: "Home", icon: "🏠" }];

export function Sidebar() {
  return (
    <aside className="w-60 border-r border-border bg-sidebar flex flex-col">
      <div className="p-5 border-b border-border">
        <h1 className="text-lg font-semibold tracking-tight text-sidebar-foreground">
          🌳 Cognitree
        </h1>
        <p className="text-xs text-muted-foreground mt-1">Thinking Tree IDE</p>
      </div>
      <nav className="flex-1 p-3">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-muted-foreground hover:text-sidebar-foreground hover:bg-sidebar-accent"
              }`
            }
          >
            <span>{item.icon}</span>
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
