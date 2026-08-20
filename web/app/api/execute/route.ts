const JUDGE_URL = process.env.JUDGE_URL ?? "http://localhost:3001";

export async function POST(request: Request) {
  const body = await request.text();
  const res = await fetch(`${JUDGE_URL}/execute`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });
  const text = await res.text();
  return new Response(text, {
    status: res.status,
    headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
  });
}
