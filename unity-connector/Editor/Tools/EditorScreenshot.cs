using System;
using System.IO;
using System.Reflection;
using Newtonsoft.Json.Linq;
using UnityEditor;
using UnityEngine;

namespace UnityCliConnector.Tools
{
    [UnityCliTool(Name = "screenshot", Description = "Capture the Game View (default) or Scene View.")]
    public static class EditorScreenshot
    {
        private const int DefaultWidth = 1920;
        private const int DefaultHeight = 1080;

        public class Parameters
        {
            [ToolParameter("View to capture: game (default), scene", Required = false)]
            public string View { get; set; }

            [ToolParameter("Override width (default: configured Game View width; scene: 1920)", Required = false)]
            public int Width { get; set; }

            [ToolParameter("Override height (default: configured Game View height; scene: 1080)", Required = false)]
            public int Height { get; set; }

            [ToolParameter("Output file path, absolute or relative to project root (default: Screenshots/screenshot.png)", Required = false)]
            public string OutputPath { get; set; }
        }

        public static object HandleCommand(JObject @params)
        {
            if (@params == null)
                @params = new JObject();

            var p = new ToolParams(@params);
            var view = p.Get("view", "game").ToLowerInvariant();
            var width = p.GetInt("width");
            var height = p.GetInt("height");
            var outputPath = ResolveOutputPath(p.Get("output_path"));

            try
            {
                var dir = Path.GetDirectoryName(outputPath);
                if (!string.IsNullOrEmpty(dir))
                    Directory.CreateDirectory(dir);

                switch (view)
                {
                    case "scene":
                        return CaptureSceneView(
                            width ?? DefaultWidth,
                            height ?? DefaultHeight,
                            outputPath);
                    case "game":
                        return CaptureGameView(width, height, outputPath);
                    default:
                        return new ErrorResponse($"Unknown view '{view}'. Valid: scene, game.");
                }
            }
            catch (Exception e)
            {
                return new ErrorResponse($"Screenshot failed: {e.Message}");
            }
        }

        private static string ResolveOutputPath(string userPath)
        {
            if (string.IsNullOrEmpty(userPath))
                userPath = "Screenshots/screenshot.png";

            if (Path.IsPathRooted(userPath))
                return Path.GetFullPath(userPath);

            var projectRoot = Path.GetDirectoryName(Application.dataPath);
            return Path.GetFullPath(Path.Combine(projectRoot, userPath));
        }

        private static object CaptureSceneView(int width, int height, string outputPath)
        {
            var sceneView = SceneView.lastActiveSceneView;
            if (!sceneView)
                return new ErrorResponse("No active SceneView found.");

            var camera = sceneView.camera;
            if (!camera)
                return new ErrorResponse("SceneView camera is null.");

            return CaptureCamera(camera, width, height, outputPath);
        }

        private static object CaptureGameView(int? requestedWidth, int? requestedHeight, string outputPath)
        {
            var camera = Camera.main;
            if (!camera)
            {
#if UNITY_2023_1_OR_NEWER
                camera = UnityEngine.Object.FindFirstObjectByType<Camera>();
#else
                camera = UnityEngine.Object.FindObjectOfType<Camera>();
#endif
                if (!camera)
                    return new ErrorResponse("No camera found in scene.");
            }

            var configuredSize = GetConfiguredGameViewSize();
            var width = requestedWidth ?? ResolveDimension(configuredSize.x, camera.pixelWidth, DefaultWidth);
            var height = requestedHeight ?? ResolveDimension(configuredSize.y, camera.pixelHeight, DefaultHeight);

            return CaptureCamera(camera, width, height, outputPath);
        }

        private static Vector2 GetConfiguredGameViewSize()
        {
            try
            {
                var gameViewType = Type.GetType("UnityEditor.GameView,UnityEditor");
                var methodFlags = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Static;
                var getTargetSize = gameViewType?.GetMethod("GetMainGameViewTargetSize", methodFlags)
                    ?? gameViewType?.GetMethod("GetSizeOfMainGameView", methodFlags);
                var result = getTargetSize?.Invoke(null, null);
                if (result is Vector2 size)
                    return size;
                if (result is Vector2Int intSize)
                    return new Vector2(intSize.x, intSize.y);
            }
            catch
            {
                // Older Editor versions can change this internal API; camera dimensions remain usable.
            }

            return Vector2.zero;
        }

        private static int ResolveDimension(float configured, int cameraDimension, int fallback)
        {
            var configuredDimension = Mathf.RoundToInt(configured);
            if (configuredDimension > 0)
                return configuredDimension;
            if (cameraDimension > 0)
                return cameraDimension;
            return fallback;
        }

        private static object CaptureCamera(Camera camera, int width, int height, string outputPath)
        {
            if (width <= 0 || height <= 0)
                return new ErrorResponse("Screenshot width and height must be greater than zero.");

            var previousRT = camera.targetTexture;
            RenderTexture rt = null;
            Texture2D tex = null;

            try
            {
                rt = new RenderTexture(width, height, 24);
                camera.targetTexture = rt;
                camera.Render();

                RenderTexture.active = rt;
                tex = new Texture2D(width, height, TextureFormat.RGB24, false);
                tex.ReadPixels(new Rect(0, 0, width, height), 0, 0);
                tex.Apply();

                File.WriteAllBytes(outputPath, tex.EncodeToPNG());

                return new SuccessResponse($"Screenshot saved to {outputPath}",
                    new { path = outputPath, width, height });
            }
            finally
            {
                camera.targetTexture = previousRT;
                RenderTexture.active = null;
                if (rt) UnityEngine.Object.DestroyImmediate(rt);
                if (tex) UnityEngine.Object.DestroyImmediate(tex);
            }
        }
    }
}
