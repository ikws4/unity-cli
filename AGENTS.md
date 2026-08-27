# unity-cli

CLI tool to control Unity Editor from the command line.

## Structure

```
cmd/                  # Go CLI — thin passthrough layer
  root.go             # Entry point, flag/arg parsing, default passthrough
  editor.go           # editor command (waitForReady polling)
  test.go             # test command (PlayMode result polling)
  status.go           # status, waitForAlive, heartbeat reading
  update.go           # self-update from GitHub releases
  version_check.go    # periodic update notice (12h interval)
internal/client/      # Unity HTTP client, instance discovery
unity-connector/      # C# Unity Editor package (UPM)
  Editor/
    Core/             # Shared utilities (Response, ParamCoercion, ToolParams, StringCaseUtility)
    Tools/            # Tool implementations (auto-registered via [UnityCliTool] attribute)
    TestRunner/       # Test runner (RunTests, TestRunnerState)
```

## Development

### Adding a Command

1. Add a C# tool in `unity-connector/Editor/Tools/` with `[UnityCliTool(Name = "command_name")]`
2. CLI command name matches the tool name — default passthrough handles dispatch
3. Positional args arrive as `args` array, flags as named params
4. Go-side code is only needed for polling/waiting logic (editor, test)

## Verification

Run all of the following before pushing:

```bash
go clean -testcache
gofmt -w .
~/go/bin/golangci-lint run ./...
~/go/bin/golangci-lint fmt --diff
go test ./...
```

### Integration Tests (requires Unity)

Integration tests are tagged with `//go:build integration` and excluded from the default test run.
Run them manually when Unity Editor is open:

```bash
go test -tags integration ./...
```

CI skips these since Unity is not available.

## Checklist

### 변경 시

CLI option, command, parameter를 수정하면 관련된 모든 곳을 함께 반영한다:
- C# tool (Parameters class, HandleCommand)
- Go help text (root.go의 overview + command별 detailed help)
- README.md

### 버전 관리

CLI(Go)와 Connector(C#)는 같은 릴리스 버전으로 맞춘다.

- 소스 기본 버전은 `0.8.0`부터 시작한다. 새 기능을 개발할 때마다 기본적으로 patch 버전을 하나 올리고, 호환성 범위가 바뀌면 SemVer에 따라 minor/major를 올린다.
- 기능 버전 변경은 해당 기능 커밋에 포함하며 `cmd.DefaultVersion`, `unity-connector/package.json`, `unity-connector/Editor/Heartbeat.cs`를 항상 함께 갱신한다.
- **CLI**: git tag는 `vX.Y.Z`
- **CLI binary**: 소스 빌드는 `cmd.DefaultVersion`을 사용한다. `.github/workflows/release.yml`은 `-X main.Version=${VERSION}`로 tag를 주입하고, `main.go`가 이를 `cmd.Version`에 전달한다.
- **Connector**: `unity-connector/package.json` 버전은 `X.Y.Z`
- **Connector heartbeat**: `unity-connector/Editor/Heartbeat.cs`의 `CONNECTOR_VERSION`도 같은 `X.Y.Z`로 갱신한다.
- `content_embed_test.go`의 버전 일치 테스트를 우회하지 않는다.
- CLI는 실행 전 Connector heartbeat의 `connectorVersion`을 읽어 버전이 다르거나 없으면 에러를 낸다. 버전 갱신 누락을 정상 동작으로 넘기지 않는다.

### 작업 마무리 시

- Verification 항목 전부 실행
- 변경한 기능은 Unity가 열려 있으면 `unity-cli`로 직접 실행해서 동작 확인
- 로컬 임시 파일(테스트용 스크립트, 디버깅 출력 등) 정리
- 관련 없는 변경은 별도 커밋으로 분리

## Git

Commit all unstaged changes before finishing. Unrelated changes should be committed separately.

## 실행 규칙

`go run .`은 테스트 목적일 때만 사용. CLI 기능 실행은 반드시 설치된 바이너리 `unity-cli`로.

## 릴리스 플로우

"커밋하고 올려" 지시 시 아래를 한 번에 수행:

1. Verification 전부 실행
2. CLI tag `vX.Y.Z`와 Connector `unity-connector/package.json` `X.Y.Z` 버전을 같은 번호로 갱신
3. 커밋 + push
4. CLI 변경 있으면 새 tag push
5. CI(CI + Release) 완료 대기 (`gh run watch --exit-status`, background)
6. `go clean -cache -testcache`로 빌드/테스트 캐시 전부 정리
7. 둘 다 성공하면 `unity-cli update`로 설치된 CLI 업데이트

## CI

- `push/PR → main`: build, vet, test, lint, format
- `tag push (v*)`: cross-compile (linux/darwin/windows × amd64/arm64) + GitHub Release
