# Testing

- Do **not** run the full Go test suite (`go test ./...` from the `server/` module)
  unless the user explicitly asks for it.
- After changing Go code, run only the tests that cover the touched package(s),
  e.g. `go test ./caf/...` or `go test ./caf -run TestName`.
- If no targeted Go test exists, say so and ask before running the full suite.
- Running the client/web tests (`bun test`, `npm test`, or `vitest run` in
  `registro-web/`) is fine — no restriction there.
