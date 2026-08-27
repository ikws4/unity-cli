---
name: unity-cli
description: "Use when controlling a running Unity Editor with unity-cli, especially when executing C# through unity-cli exec, inspecting scenes, editing Unity objects or assets, running tests, or diagnosing command failures."
metadata:
  requires:
    bins: ["unity-cli"]
---

# unity-cli

Control the running Unity Editor through `unity-cli`. Read live state before making assumptions:

```bash
unity-cli status
unity-cli list
```

When multiple Editors are open, pass the project explicitly with the global `--project <path>` option.

## Execute C# safely

`unity-cli exec` wraps the supplied code inside a synchronous method. The wrapper already imports:

`System`, `System.Collections.Generic`, `System.IO`, `System.Linq`, `System.Reflection`, `System.Threading.Tasks`, `UnityEngine`, `UnityEngine.SceneManagement`, `UnityEditor`, `UnityEditor.SceneManagement`, and `UnityEditorInternal`.

Follow these rules:

1. Do not put `using` directives in the code string. The code runs inside a method, where `using` directives are invalid. Add project namespaces with repeatable `--usings` flags.
2. `Object` is ambiguous because both `System` and `UnityEngine` are imported. Always write `UnityEngine.Object` for Unity objects. There is no `Unity.Object` type.
3. Use `System.Object` only when the CLR base type is specifically intended.
4. Return a value explicitly. End mutations with `return null;` when no result is needed.
5. Prefer stdin for multi-statement code so shell quoting does not corrupt C# strings.
6. Keep execution synchronous. Async code, coroutines, delayed callbacks, and deferred Unity APIs are blocked by default because the command can return before they finish. Use `--allow-async` only when that behavior is intentional.

Correct examples:

```bash
unity-cli exec 'return UnityEngine.Object.FindObjectsOfType<Camera>().Length;'

printf '%s\n' \
  'var go = new GameObject("Marker");' \
  'UnityEngine.Object.DestroyImmediate(go);' \
  'return null;' | unity-cli exec

unity-cli exec 'return World.All.Count;' \
  --usings Unity.Entities \
  --usings Unity.Mathematics
```

Incorrect examples:

```csharp
Object.FindObjectOfType<Camera>()       // ambiguous: System.Object or UnityEngine.Object
Unity.Object.FindObjectOfType<Camera>() // Unity.Object does not exist
using Unity.Entities;                  // invalid inside the generated Execute method
```

## Work with the live Editor

- Use `unity-cli list` to discover the exact registered tools and parameter schemas before guessing a command.
- Use `unity-cli console --type error --stacktrace user` after compilation or runtime failures.
- Prefer Editor commands and `exec` for scene, GameObject, prefab, and asset changes while Unity is running. Save through Unity after mutations.
- Use `unity-cli test --mode EditMode` or `--mode PlayMode` for project tests.
- If the connector is unreachable while Unity is open, inspect compile errors first; a broken Editor assembly prevents the connector from loading.

Read the copy embedded in the installed binary with `unity-cli skills read unity-cli` so guidance stays aligned with that CLI version.
