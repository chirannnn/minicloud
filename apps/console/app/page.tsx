"use client";

import { useEffect } from "react";

export default function Home() {
  useEffect(() => {
    window.location.href = "/dashboard";
  }, []);

  return <main><p>Redirecting to dashboard...</p></main>;
}
