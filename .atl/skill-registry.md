# Skill Registry — net-probe-exporter

Persistent map of skill-name → path, resolved once per session by the SDD orchestrator
and passed to sub-agents. Sub-agents do NOT search for this registry themselves.

## Project
- Name: net-probe-exporter
- Path: /home/jsojo/Documentos/OPEN_SOURCE/net-probe-exporter
- Stack: Go 1.25 (module github.com/josesojo2828/net-probe-exporter)
- Default artifact store: engram

## Skills

| Skill | Path | When to use |
|-------|------|-------------|
| go-testing | /home/jsojo/.claude/skills/go-testing/SKILL.md | Writing Go tests, using teatest, or adding test coverage (table-driven `*testing.T`, `*_test.go`) |
| sdd-explore | /home/jsojo/.claude/skills/sdd-explore/SKILL.md | Orchestrator: explore/investigate before committing to a change |
| sdd-propose | /home/jsojo/.config/opencode/skills/sdd-propose/SKILL.md | Orchestrator: create a change proposal |
| sdd-spec | /home/jsojo/.claude/skills/sdd-spec/SKILL.md | Orchestrator: write specs with requirements/scenarios |
| sdd-design | /home/jsojo/.config/opencode/skills/sdd-design/SKILL.md | Orchestrator: technical design document |
| sdd-tasks | /home/jsojo/.claude/skills/sdd-tasks/SKILL.md | Orchestrator: task breakdown |
| sdd-apply | /home/jsojo/.claude/skills/sdd-apply/SKILL.md | Orchestrator: implement tasks into code |
| sdd-verify | /home/jsojo/.claude/skills/sdd-verify/SKILL.md | Orchestrator: verify implementation vs specs |
| sdd-archive | /home/jsojo/.claude/skills/sdd-archive/SKILL.md | Orchestrator: archive a completed change |

## Notes
- Convention: project is fully dockerized. Tests/lint/build run via `docker compose`
  (Makefile targets use `golang:1.25-alpine` tester service). When delegating Go work,
  the sub-agent should run tests via `make test` / `make test-race`, not bare `go test`
  unless a local Go toolchain is confirmed available.
- Linter is `gofmt -l .` (no golangci-lint configured).
