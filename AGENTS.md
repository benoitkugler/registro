# Testing

- Do **not** run the full Go test suite (`go test ./...` from the `server/` module)
  or all tests in a package (e.g. `go test ./caf`, `go test ./caf/...`) unless the
  user explicitly asks for it.
- After changing Go code, run only the specific tests covering the touched code,
  selected by name with `-run` against the single package changed, e.g.
  `go test ./caf -run TestName` or `go test ./caf -run 'TestFoo|TestBar'`.
- If no targeted Go test exists, say so and ask before running the package suite
  or the full suite.
- In **Plan mode**, never propose full-package or full-module test runs. Plans
  must list the exact targeted commands (e.g. `go test ./caf -run TestName`) to
  run after implementation, and ask the user before broadening the scope.
- Running the client/web tests (`bun test`, `npm test`, or `vitest run` in
  `registro-web/`) is fine — no restriction there.
