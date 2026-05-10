// Phase-1 placeholder. Phase 5 wires this up to the Connect-Web client and
// renders the real observability views (nodes, HTTP stream, SQL stream,
// session inventory).
export default function App(): JSX.Element {
  return (
    <main className="flex min-h-screen items-center justify-center p-8">
      <div className="max-w-xl space-y-3 text-center">
        <h1 className="text-2xl font-semibold">Nucleus Admin · Observability</h1>
        <p className="text-sm text-zinc-400">
          Phase&nbsp;1 skeleton. The real-time views will land in Phase&nbsp;5
          once the admin server (Phase&nbsp;4) and the agent (Phase&nbsp;3) are
          shipping events.
        </p>
        <p className="text-xs text-zinc-500">
          Generated stubs for the <code>nucleus.admin.v1</code> contract live
          in <code>src/gen/</code> and are produced by{' '}
          <code>make proto</code> from the repository root.
        </p>
      </div>
    </main>
  )
}
