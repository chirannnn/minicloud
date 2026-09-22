"use client";

import { useEffect, useState } from "react";

type Status = "checking" | "connected" | "unavailable";

export default function Home() {
  const [status, setStatus] = useState<Status>("checking");
  useEffect(() => {
    fetch(`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/health`)
      .then((response) => setStatus(response.ok ? "connected" : "unavailable"))
      .catch(() => setStatus("unavailable"));
  }, []);
  return <main><p className="eyebrow">DEVELOPMENT CONSOLE</p><h1>MiniCloud</h1><p className="intro">A small, real cloud engineering platform. Phase 0 provides the local foundation.</p><section><h2>API connectivity</h2><p data-status={status}>{status}</p></section><section><h2>Next</h2><p>Phase 1 will add the core control-plane resources. The Live Engineering Observatory follows in Phase 2.</p></section></main>;
}
