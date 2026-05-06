# Project Context

## Knowledge Graph Navigation

**ALWAYS read `graphify-out/GRAPH_REPORT.md` before exploring or answering questions about this codebase.**

The knowledge graph maps:
- 314 nodes · 624 edges · 18 communities
- Navigation hubs: `respondError()`, `Setup()`, `main()`, `ServiceRequestHandler`, `ServiceRequestRepository`, `API surface`, `ExpoPushService`, `NewsHandler`, etc.
- Core abstractions ("God Nodes"): `respondError()` (42 edges), `ServiceRequestRepository` (20), `Setup()` (18)

**Workflow:**
1. Read `graphify-out/GRAPH_REPORT.md` to identify which community/module is relevant
2. Use the community hubs to locate the right files
3. Only then read the specific source files

**When graph is stale** (new commits after the graph was built), run `graphify update .` to refresh it without API cost.
