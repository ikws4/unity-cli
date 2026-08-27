# unity-cli

> 从命令行控制正在运行的 Unity Editor，适合开发者、自动化脚本和 AI Agent。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`unity-cli` 由一个 Go 命令行程序和一个 Unity Editor Connector 包组成。Connector 在本机
Editor 内启动 HTTP 监听，CLI 自动发现实例并发送命令，不需要额外运行 Python、MCP Server
或中转进程。

## 环境要求

- Unity 2022.3 或更高版本
- 使用源码安装时需要 Go 1.24 或更高版本
- Connector 使用 Git URL 安装时，本机需要可供 Unity Package Manager 使用的 Git

## 安装 CLI

推荐直接安装 Go package：

```bash
go install github.com/ikws4/unity-cli@latest
```

如果安装后找不到命令，请确认 Go 的 bin 目录已加入 `PATH`：

```bash
go env GOBIN GOPATH
export PATH="$(go env GOPATH)/bin:$PATH"
```

也可以克隆源码并使用 Makefile：

```bash
git clone https://github.com/ikws4/unity-cli.git
cd unity-cli
make install
```

`make install` 优先使用 `GOBIN`，未设置时安装到第一个 `GOPATH/bin`。安装到系统目录：

```bash
sudo env GOBIN=/usr/local/bin make install
```

macOS 和 Linux 也可以使用安装脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/ikws4/unity-cli/main/install.sh | sh
```

Windows 用户可以使用 `go install`，或从
[GitHub Releases](https://github.com/ikws4/unity-cli/releases) 下载对应的 `.exe`。

验证安装：

```bash
unity-cli version
```

## 安装 Unity Connector

进入 Unity 项目根目录，也就是同时包含 `Assets`、`Packages` 和 `ProjectSettings` 的目录，
运行：

```bash
cd /path/to/MyUnityProject
unity-cli setup
```

`setup` 会检查当前目录并更新 `Packages/manifest.json`：

- 安装 `com.ikws4.unity-cli-connector`
- 发布版 CLI 将 Connector 锁定到相同的 Git tag
- 重复执行时自动更新旧版本，已一致时不改写文件
- 自动迁移旧的 `com.youngwoocho02.unity-cli-connector` 依赖

也可以在 Unity 的 **Package Manager → Add package from git URL** 中手动添加：

```text
https://github.com/ikws4/unity-cli.git?path=unity-connector
```

或直接修改 `Packages/manifest.json`：

```json
{
  "dependencies": {
    "com.ikws4.unity-cli-connector": "https://github.com/ikws4/unity-cli.git?path=unity-connector"
  }
}
```

添加完成后打开 Unity，Connector 会自动启动，无需其他配置。

### 推荐的 Editor 设置

Unity 窗口失去焦点时可能降低 Editor 更新频率，导致需要主线程执行的 CLI 命令延迟。
建议在 **Edit → Preferences → General → Interaction Mode** 中选择 **No Throttling**。

## 快速开始

```bash
# 查看连接状态
unity-cli status

# 进入 Play Mode 并等待完成
unity-cli editor play --wait

# 在 Editor 内执行 C#
unity-cli exec "return Application.dataPath;"

# 读取 Console
unity-cli console --type error,warning,log
```

## 工作原理

```text
终端                                      Unity Editor
unity-cli editor play --wait
  │
  ├─ 扫描 ~/.unity-cli/instances/*.json
  ├─ 按当前目录或 --project 选择实例
  ├─ 向本机 Connector 发送命令 ────────────→ CommandRouter
  │                                           │
  │                                      主线程执行工具
  │                                           │
  └─ 输出 JSON 响应 ←─────────────────────────┘
```

Connector 会：

1. 在 Editor 启动时打开本机 HTTP 监听。
2. 为每个项目写入独立的 instance 文件。
3. 每 0.5 秒更新 Editor 状态与心跳。
4. 通过反射发现内置及项目自定义的 `[UnityCliTool]`。
5. 在 Unity 主线程分发需要访问 Editor API 的命令。
6. 在脚本编译和 Domain Reload 后自动恢复。

CLI 在发送命令前会检查心跳。Unity 正在编译或重载时，CLI 会等待 Editor 恢复响应。

## 命令概览

| 命令 | 说明 |
|---|---|
| `setup` | 在当前 Unity 项目安装或更新 Connector |
| `status` | 查看 Unity Editor 连接状态 |
| `editor` | 控制播放、停止、暂停、刷新和编译 |
| `console` | 读取、筛选或清空 Console 日志 |
| `exec` | 在 Unity Editor 内执行 C# |
| `menu` | 按路径执行 Unity 菜单项 |
| `screenshot` | 截取 Scene View 或 Game View |
| `reserialize` | 通过 Unity Serializer 重新序列化资源 |
| `test` | 运行 EditMode 或 PlayMode 测试 |
| `profiler` | 读取 Profiler 层级并控制录制 |
| `list` | 列出内置和项目自定义工具及参数 |
| `skills` | 查看 CLI 内嵌的 Agent 使用指南 |
| `update` | 检查或安装最新 CLI 版本 |

任何命令都可以使用 `--help` 查看完整参数：

```bash
unity-cli editor --help
unity-cli exec --help
unity-cli profiler --help
```

## Editor 控制

```bash
# 进入播放模式
unity-cli editor play

# 进入播放模式并等待完成
unity-cli editor play --wait

# 停止、暂停或恢复
unity-cli editor stop
unity-cli editor pause

# 刷新资源
unity-cli editor refresh

# 刷新并等待脚本编译完成
unity-cli editor refresh --compile

# 在 Play Mode 中强制刷新
unity-cli editor refresh --force
```

## Console 日志

```bash
# 读取日志
unity-cli console

# 限制数量和类型
unity-cli console --lines 20 --type error,warning,log

# 只读取错误并包含用户代码堆栈
unity-cli console --type error --stacktrace user

# 清空 Console
unity-cli console --clear
```

`--stacktrace` 支持 `none`、`user` 和 `full`。

## 执行 C#

`exec` 可以访问 `UnityEngine`、`UnityEditor`、项目程序集以及已经加载的其他程序集。
使用 `return` 返回结果；没有返回值的修改以 `return null;` 结束。

```bash
unity-cli exec "return Application.dataPath;"
unity-cli exec "return EditorSceneManager.GetActiveScene().name;"
unity-cli exec "return UnityEngine.Object.FindObjectsOfType<Camera>().Length;"

# 添加项目命名空间，可以重复指定 --usings
unity-cli exec "return World.All.Count;" --usings Unity.Entities --usings Unity.Mathematics

# 多语句代码建议通过 stdin 传入，避免 Shell 转义问题
echo 'Debug.Log("hello"); return null;' | unity-cli exec
```

注意：

- 默认已经导入 `System` 和 `UnityEngine`，因此必须写 `UnityEngine.Object`；单独的
  `Object` 存在歧义，`Unity.Object` 类型不存在。
- 不要在代码字符串内写 `using`，请使用 `--usings`。
- 异步、协程和延迟回调默认被阻止，因为命令返回时它们可能尚未完成。仅在明确需要时使用
  `--allow-async`。
- 编译器和 dotnet runtime 默认自动发现，必要时可传入 `--csc` 和 `--dotnet`。

## 菜单与截图

```bash
unity-cli menu "File/Save Project"
unity-cli menu "Assets/Refresh"

unity-cli screenshot
unity-cli screenshot --view game
unity-cli screenshot --output_path captures/game.png --width 1920 --height 1080
```

出于安全考虑，`menu` 不允许执行 `File/Quit`。

## 重新序列化资源

直接修改 `.prefab`、`.unity`、`.asset` 或 `.mat` YAML 后，可以让 Unity 重新加载并使用
自身 Serializer 写回，降低错误缩进、过期 `fileID` 或字段格式导致资源损坏的风险。

```bash
# 整个项目
unity-cli reserialize

# 指定一个或多个资源
unity-cli reserialize Assets/Prefabs/Player.prefab
unity-cli reserialize Assets/Scenes/Main.unity Assets/Scenes/Lobby.unity
```

## Profiler

```bash
unity-cli profiler hierarchy
unity-cli profiler hierarchy --depth 3
unity-cli profiler hierarchy --root SimulationSystem --depth 3
unity-cli profiler hierarchy --parent 4 --depth 2
unity-cli profiler hierarchy --frames 30 --min 0.5
unity-cli profiler hierarchy --from 100 --to 200
unity-cli profiler hierarchy --min 0.5 --sort self --max 10

unity-cli profiler enable
unity-cli profiler disable
unity-cli profiler status
unity-cli profiler clear
```

## 测试

项目需要安装 Unity Test Framework：

```bash
# EditMode
unity-cli test

# PlayMode
unity-cli test --mode PlayMode

# 按完整测试名称筛选
unity-cli test --filter MyNamespace.MyTests.SpecificTest
```

PlayMode 测试可能触发 Domain Reload，CLI 会自动轮询结果文件。

## 自定义工具

在项目的 Editor Assembly 中创建带 `[UnityCliTool]` 的静态类。Connector 会在 Domain
Reload 后自动发现，无需修改 Go CLI。

```csharp
using UnityCliConnector;
using Newtonsoft.Json.Linq;
using UnityEngine;

[UnityCliTool(Name = "spawn", Description = "Spawn an enemy", Group = "gameplay")]
public static class SpawnEnemy
{
    public class Parameters
    {
        [ToolParameter("X position", Required = true)]
        public float X { get; set; }

        [ToolParameter("Prefab name", DefaultValue = "Enemy")]
        public string Prefab { get; set; }
    }

    public static object HandleCommand(JObject parameters)
    {
        var p = new ToolParams(parameters);
        float x = p.GetFloat("x", 0);
        string prefabName = p.Get("prefab", "Enemy");

        var prefab = Resources.Load<GameObject>(prefabName);
        var instance = UnityEngine.Object.Instantiate(
            prefab, new Vector3(x, 0, 0), Quaternion.identity);

        return new SuccessResponse("Enemy spawned", new { instance.name });
    }
}
```

查看并调用：

```bash
unity-cli list
unity-cli spawn --x 1 --prefab Enemy
unity-cli spawn --params '{"x":1,"prefab":"Enemy"}'
```

规则：

- 类必须是 `static`。
- 入口必须是 `public static object HandleCommand(JObject parameters)`，也支持
  `async Task<object>` 变体。
- 返回 `SuccessResponse` 或 `ErrorResponse`。
- 嵌套 `Parameters` 类不是必需的，但建议使用，以便 `unity-cli list` 暴露参数 schema。
- 未指定 `Name` 时，类名会自动转为 snake_case。
- 工具在 Unity 主线程执行，可以访问 Editor API。
- 重名工具只注册第一个，并在 Console 中报告错误。

常用属性：

| Attribute | 属性 | 说明 |
|---|---|---|
| `[UnityCliTool]` | `Name` | 覆盖命令名 |
| `[UnityCliTool]` | `Description` | 工具说明 |
| `[UnityCliTool]` | `Group` | 工具分组 |
| `[ToolParameter]` | `Description` | 参数说明 |
| `[ToolParameter]` | `Required` | 是否必填 |
| `[ToolParameter]` | `Name` | 覆盖参数名 |
| `[ToolParameter]` | `DefaultValue` | 默认值提示 |

## 多 Unity 实例

每个打开的 Unity 项目都会注册独立实例。默认优先匹配当前工作目录中的项目；无法唯一匹配时，
使用全局 `--project`：

```bash
unity-cli status
unity-cli --project /path/to/MyGame editor play
```

全局参数：

| 参数 | 说明 | 默认值 |
|---|---|---|
| `--project <path>` | 按项目路径选择 Unity 实例 | 自动选择 |
| `--timeout <ms>` | 命令超时时间 | `120000` |
| `--ignore-version-mismatch` | 跳过 CLI/Connector 版本检查 | `false` |

## Agent Skill

安装仓库提供的 Agent Skill：

```bash
npx skills add ikws4/unity-cli -y -g
```

CLI 二进制中也内嵌了同一份说明，可以直接读取：

```bash
unity-cli skills list
unity-cli skills read unity-cli
```

## 更新

```bash
unity-cli update --check
unity-cli update
```

更新命令从 `github.com/ikws4/unity-cli` 的最新 GitHub Release 下载当前平台的二进制。

## 开发与验证

```bash
make build
make test

go clean -testcache
gofmt -w .
golangci-lint run ./...
golangci-lint fmt --diff
go test ./...
```

需要打开 Unity Editor 的集成测试：

```bash
go test -tags integration ./...
```

## License

[MIT](LICENSE)
