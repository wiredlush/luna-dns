import { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./Login";
import Layout from "./Layout";
import Dashboard from "./Dashboard";
import DnsServer from "./DnsServer";
import Blocklist from "./Blocklist";
import Security from "./Security";
import Users from "./Users";

export default function App() {
  const [authed, setAuthed] = useState<boolean | null>(null);

  useEffect(() => {
    fetch("/api/health", { credentials: "same-origin" })
      .then((res) => setAuthed(res.ok))
      .catch(() => setAuthed(false));
  }, []);

  if (authed === null) return null;

  if (!authed) {
    return <Login onLogin={() => setAuthed(true)} />;
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout onLogout={() => setAuthed(false)} />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/dns" element={<DnsServer />} />
          <Route path="/blocklist" element={<Blocklist />} />
          <Route path="/security" element={<Security />} />
          <Route path="/users" element={<Users />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
