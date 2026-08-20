const JUDGE_URL = process.env.JUDGE_URL ?? "http://localhost:3001";

export async function GET(request: Request) {
  const id = new URL(request.url).searchParams.get("id") ?? "";
  const res = await fetch(`${JUDGE_URL}/result?id=${encodeURIComponent(id)}`);
  const text = await res.text();
  return new Response(text, {
    status: res.status,
    headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
  });
}
